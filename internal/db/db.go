package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/qustavo/sqlhooks/v2"
)

// To not have key collisions
// empty struct is zero bytes
type ctxKeyStart struct{}
type ctxKeyQuery struct{}
type ctxKeyArgs struct{}

var globalHooks = &logHooks{}

func init() {
	sql.Register("pq-logged", sqlhooks.Wrap(&pq.Driver{}, globalHooks))
}

type logHooks struct {
	logQuery bool
	logTime  bool
}

func (h *logHooks) Before(ctx context.Context, query string, args ...any) (context.Context, error) {
	if h.logQuery {
		ctx = context.WithValue(ctx, ctxKeyQuery{}, query)
		ctx = context.WithValue(ctx, ctxKeyArgs{}, args)
	}
	if h.logTime {
		ctx = context.WithValue(ctx, ctxKeyStart{}, time.Now())
	}
	return ctx, nil
}

func (h *logHooks) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	var parts []string
	if h.logTime {
		if start, ok := ctx.Value(ctxKeyStart{}).(time.Time); ok {
			parts = append(parts, time.Since(start).String())
		}
	}
	if h.logQuery {
		if q, ok := ctx.Value(ctxKeyQuery{}).(string); ok {
			parts = append(parts, q)
		}
		if a, ok := ctx.Value(ctxKeyArgs{}).([]any); ok {
			parts = append(parts, fmt.Sprintf("args: %v", a))
		}
	}
	if len(parts) > 0 {
		log.Printf("[SQL] %s", strings.Join(parts, " | "))
	}
	return ctx, nil
}

// PostgreSQL error codes
const (
	PgErrorUniqueViolation     = "23505"
	PgErrorForeignKeyViolation = "23503"
	PgErrorNotNullViolation    = "23502"
	PgErrorCheckViolation      = "23514"
)

type DB struct {
	*sql.DB
}

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Pass     string
	LogQuery bool
	LogTime  bool
}

var (
	dbOnce     sync.Once
	dbInstance *DB
	initError  error
)

func Init(config *DBConfig) (*DB, error) {
	dbOnce.Do(func() {
		log.Println("Start init DB")
		globalHooks.logQuery = config.LogQuery
		globalHooks.logTime = config.LogTime

		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.User, config.Pass, config.Host, config.Port, config.Name)

		rawDB, err := sql.Open("pq-logged", connStr)
		if err != nil {
			initError = fmt.Errorf("error opening DB: %w", err)
			return
		}
		rawDB.SetMaxOpenConns(25)
		rawDB.SetMaxIdleConns(5)
		rawDB.SetConnMaxLifetime(5 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := rawDB.PingContext(ctx); err != nil {
			rawDB.Close()
			initError = fmt.Errorf("ping to DB error: %w", err)
			return
		}
		dbInstance = &DB{DB: rawDB}
	})

	if initError != nil {
		return nil, initError
	}

	log.Println("DB inited successfully")
	return dbInstance, nil
}
