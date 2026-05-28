//go:build ignore

// Демонстрация production-ready флоу POST /bookings.
// Запустить: go run system_design.go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// Stub-типы — имитируют реальные хранилища
// ---------------------------------------------------------------------------

type Booking struct {
	ID         int64
	UserID     int64
	RoomID     int64
	CheckIn    time.Time
	CheckOut   time.Time
	Status     string
	TotalPrice float64
}

type DB struct{}

// SelectRoomForUpdate — pessimistic lock: другие транзакции ждут пока мы не завершимся.
// В реальности: SELECT ... FOR UPDATE
func (db *DB) SelectRoomForUpdate(ctx context.Context, tx *Tx, roomID int64) (pricePerNight float64, err error) {
	fmt.Printf("  [DB] SELECT * FROM rooms WHERE id=%d FOR UPDATE\n", roomID)
	return 5000.0, nil
}

func (db *DB) IsAvailable(ctx context.Context, tx *Tx, roomID int64, checkIn, checkOut time.Time) (bool, error) {
	fmt.Printf("  [DB] SELECT COUNT(*) FROM bookings WHERE room_id=%d AND dates overlap\n", roomID)
	return true, nil
}

func (db *DB) InsertBooking(ctx context.Context, tx *Tx, b Booking) (Booking, error) {
	b.ID = 42
	b.Status = "pending" // pending до подтверждения оплаты
	fmt.Printf("  [DB] INSERT INTO bookings ... id=%d status=%s\n", b.ID, b.Status)
	return b, nil
}

func (db *DB) InsertOutbox(ctx context.Context, tx *Tx, eventType string, payload any) error {
	fmt.Printf("  [DB] INSERT INTO outbox (event_type='%s') — в той же транзакции\n", eventType)
	return nil
}

func (db *DB) WithTx(ctx context.Context, fn func(*Tx) error) error {
	fmt.Println("  [DB] BEGIN")
	tx := &Tx{}
	if err := fn(tx); err != nil {
		fmt.Println("  [DB] ROLLBACK")
		return err
	}
	fmt.Println("  [DB] COMMIT")
	return nil
}

type Tx struct{} // передаётся в методы DB чтобы работали внутри транзакции

// ---------------------------------------------------------------------------

type Redis struct{}

var ErrAlreadyProcessed = errors.New("idempotency: already processed")
var ErrLockNotAcquired = errors.New("lock: room already being booked")

// IdempotencyCheck — если запрос с таким ключом уже был, возвращаем ошибку-дубль.
// SET NX EX: атомарная операция, работает в distributed среде.
func (r *Redis) IdempotencyCheck(ctx context.Context, key string) error {
	fmt.Printf("  [Redis] SET %s NX EX 86400 → ", key)
	// в реальности: false если ключ уже существует
	fmt.Println("OK (новый запрос)")
	return nil
}

// AcquireLock — distributed lock на (roomID, даты).
// Защищает от параллельных запросов на один номер.
func (r *Redis) AcquireLock(ctx context.Context, roomID int64) (unlock func(), err error) {
	key := fmt.Sprintf("lock:room:%d", roomID)
	fmt.Printf("  [Redis] SET %s NX EX 10 → OK\n", key)
	unlock = func() {
		fmt.Printf("  [Redis] DEL %s\n", key)
	}
	return unlock, nil
}

// ---------------------------------------------------------------------------

type KafkaMessage struct {
	Value  []byte
	commit func()
}

func (m *KafkaMessage) Commit() {
	fmt.Println("  [Kafka] offset committed")
	m.commit()
}

type Kafka struct {
	messages chan *KafkaMessage
}

func NewKafka() *Kafka {
	return &Kafka{messages: make(chan *KafkaMessage, 10)}
}

// Publish вызывается Outbox Relay'ем.
func (k *Kafka) Publish(eventType string, payload any) error {
	fmt.Printf("  [Kafka] topic=%s payload=%v\n", eventType, payload)
	done := make(chan struct{})
	k.messages <- &KafkaMessage{
		Value:  []byte(`{"event_id":1,"booking_id":42,"user_id":1}`),
		commit: func() { close(done) },
	}
	return nil
}

// Subscribe возвращает канал сообщений — консьюмер читает из него в цикле.
func (k *Kafka) Subscribe(ctx context.Context, topic string) <-chan *KafkaMessage {
	fmt.Printf("  [Kafka] subscribed to topic=%s\n", topic)
	return k.messages
}

// ---------------------------------------------------------------------------
// Сам флоу
// ---------------------------------------------------------------------------

type BookingService struct {
	db    *DB
	redis *Redis
	kafka *Kafka
}

type CreateBookingRequest struct {
	IdempotencyKey string // заголовок X-Idempotency-Key от клиента
	UserID         int64
	RoomID         int64
	CheckIn        time.Time
	CheckOut       time.Time
}

func (s *BookingService) CreateBooking(ctx context.Context, req CreateBookingRequest) (Booking, error) {
	// 1. Idempotency — защита от дублей на уровне клиентского retry
	fmt.Println("\nШаг 1: Idempotency check")
	if err := s.redis.IdempotencyCheck(ctx, "idempotency:"+req.IdempotencyKey); err != nil {
		return Booking{}, ErrAlreadyProcessed
	}

	// 2. Distributed lock — защита от параллельных запросов на тот же номер
	fmt.Println("\n Шаг 2: Distributed lock на room")
	unlock, err := s.redis.AcquireLock(ctx, req.RoomID)
	if err != nil {
		return Booking{}, ErrLockNotAcquired
	}
	defer unlock()

	// 3. Транзакция в БД
	fmt.Println("\n Шаг 3: Транзакция")
	var booking Booking
	err = s.db.WithTx(ctx, func(tx *Tx) error {

		// SELECT FOR UPDATE — pessimistic lock на уровне БД.
		// Даже если два запроса прошли через Redis lock одновременно —
		// второй будет ждать здесь пока первый не закоммитится.
		pricePerNight, err := s.db.SelectRoomForUpdate(ctx, tx, req.RoomID)
		if err != nil {
			return err
		}

		// Финальная проверка доступности — уже под локом, консистентно
		available, err := s.db.IsAvailable(ctx, tx, req.RoomID, req.CheckIn, req.CheckOut)
		if err != nil {
			return err
		}
		if !available {
			return errors.New("room not available")
		}

		nights := req.CheckOut.Sub(req.CheckIn).Hours() / 24
		booking, err = s.db.InsertBooking(ctx, tx, Booking{
			UserID:     req.UserID,
			RoomID:     req.RoomID,
			CheckIn:    req.CheckIn,
			CheckOut:   req.CheckOut,
			TotalPrice: pricePerNight * nights,
		})
		if err != nil {
			return err
		}

		// Outbox — событие записывается в ТУ ЖЕ транзакцию.
		// Если коммит упадёт — событие тоже не запишется. Атомарность гарантирована.
		return s.db.InsertOutbox(ctx, tx, "booking.created", booking)
	})
	if err != nil {
		return Booking{}, err
	}

	// 4. Ответ клиенту — НЕМЕДЛЕННО, не ждём Kafka
	fmt.Printf("\n Шаг 4: Ответ клиенту, booking.id=%d status=%s\n", booking.ID, booking.Status)
	return booking, nil
}

// ---------------------------------------------------------------------------
// Outbox Relay — отдельная горутина/сервис
// ---------------------------------------------------------------------------

// OutboxRelay читает неотправленные события из outbox и публикует в Kafka.
// Запускается независимо от HTTP-сервера.
// At-least-once: если упадёт после Publish но до UPDATE — отправит повторно.
// Консьюмер Kafka должен быть идемпотентен.
func (s *BookingService) OutboxRelay(ctx context.Context) {
	fmt.Println("\n── Outbox Relay: polling outbox...")
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// SELECT id, event_type, payload FROM outbox WHERE sent_at IS NULL LIMIT 100
			// for each row:
			//   kafka.Publish(row.EventType, row.Payload)
			//   UPDATE outbox SET sent_at = NOW() WHERE id = row.ID
			s.kafka.Publish("booking.created", map[string]any{"booking_id": 42})
			fmt.Println("  [OutboxRelay] событие доставлено, outbox обновлён")
			return // для демо выходим после первой итерации
		}
	}
}

// ---------------------------------------------------------------------------
// Consumer (Notification Service) — отдельный сервис, слушает Kafka
// ---------------------------------------------------------------------------

type BookingCreatedEvent struct {
	EventID   int64 // outbox.id — используется для дедупликации
	BookingID int64
	UserID    int64
}

type NotificationDB struct{}

func (db *NotificationDB) IsProcessed(ctx context.Context, tx *Tx, eventID int64) (bool, error) {
	fmt.Printf("  [NotifDB] SELECT EXISTS processed_events WHERE event_id=%d\n", eventID)
	return false, nil // false = ещё не обрабатывали
}

func (db *NotificationDB) MarkProcessed(ctx context.Context, tx *Tx, eventID int64) error {
	fmt.Printf("  [NotifDB] INSERT INTO processed_events (event_id=%d)\n", eventID)
	return nil
}

func (db *NotificationDB) WithTx(ctx context.Context, fn func(*Tx) error) error {
	fmt.Println("  [NotifDB] BEGIN")
	if err := fn(&Tx{}); err != nil {
		fmt.Println("  [NotifDB] ROLLBACK")
		return err
	}
	fmt.Println("  [NotifDB] COMMIT")
	return nil
}

type EmailSender struct{}

func (e *EmailSender) Send(userID int64, msg string) error {
	fmt.Printf("  [Email] → userID=%d: %s\n", userID, msg)
	return nil
}

type NotificationConsumer struct {
	db    *NotificationDB
	email *EmailSender
}

// Handle вызывается для каждого сообщения из Kafka topic=booking.created.
// Идемпотентен: повторная обработка одного EventID безопасна.
func (c *NotificationConsumer) Handle(ctx context.Context, event BookingCreatedEvent) error {
	return c.db.WithTx(ctx, func(tx *Tx) error {
		// Проверяем — уже обрабатывали это событие?
		processed, err := c.db.IsProcessed(ctx, tx, event.EventID)
		if err != nil {
			return err
		}
		if processed {
			fmt.Println("  [Consumer] дубль, пропускаем")
			return nil
		}

		// Фиксируем факт обработки ДО отправки email.
		// Если email упадёт — при следующем retry мы сюда не дойдём (уже processed).
		// Принимаем: лучше один раз не отправить, чем отправить дважды.
		if err := c.db.MarkProcessed(ctx, tx, event.EventID); err != nil {
			return err
		}

		// Отправляем email — уже вне транзакции (email не транзакционный)
		return c.email.Send(event.UserID, fmt.Sprintf("Бронь #%d подтверждена", event.BookingID))
	})
}

// Run читает сообщения из Kafka в цикле.
// Offset коммитится только после успешного Handle — при ошибке сообщение придёт повторно.
func (c *NotificationConsumer) Run(ctx context.Context, kafka *Kafka) {
	fmt.Println("  [Consumer] подписался на topic=booking.created")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-kafka.Subscribe(ctx, "booking.created"):
			// В реальности: json.Unmarshal(msg.Value, &event)
			event := BookingCreatedEvent{EventID: 1, BookingID: 42, UserID: 1}

			if err := c.Handle(ctx, event); err != nil {
				// Не коммитим offset — Kafka доставит это сообщение повторно
				fmt.Printf("  [Consumer] ошибка обработки, offset не коммитим: %v\n", err)
				continue
			}

			// Коммитим offset только после успешной обработки
			msg.Commit()
		}
	}
}

// ---------------------------------------------------------------------------

func main() {
	kafka := NewKafka()

	svc := &BookingService{
		db:    &DB{},
		redis: &Redis{},
		kafka: kafka,
	}

	ctx := context.Background()

	booking, err := svc.CreateBooking(ctx, CreateBookingRequest{
		IdempotencyKey: "client-uuid-abc123",
		UserID:         1,
		RoomID:         5,
		CheckIn:        time.Now().Add(24 * time.Hour),
		CheckOut:       time.Now().Add(72 * time.Hour),
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("\n Бронь создана: id=%d price=%.0f\n", booking.ID, booking.TotalPrice)

	// Outbox Relay публикует событие в Kafka
	svc.OutboxRelay(ctx)

	// Consumer читает из Kafka и обрабатывает
	fmt.Println("\n── Notification Consumer запущен")
	consumer := &NotificationConsumer{db: &NotificationDB{}, email: &EmailSender{}}
	consumer.Run(ctx, kafka)
}
