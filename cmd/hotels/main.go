package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/stdpmk/hotels/internal/config"
	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/http/handlers"
	"github.com/stdpmk/hotels/internal/http/middleware"
	"github.com/stdpmk/hotels/internal/httpstat"
	"github.com/stdpmk/hotels/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Init(&db.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Name:     cfg.DBName,
		User:     cfg.DBUser,
		Pass:     cfg.DBPass,
		LogQuery:      cfg.SQLLogQuery,
		LogTime:       cfg.SQLLogTime,
		SlowThreshold: cfg.SQLSlowThreshold,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	hotelsSvc := services.NewHotelsService(database)
	hotelsHandler := handlers.NewHotelsHandler(hotelsSvc)

	usersSvc := services.NewUsersService(database, redisClient)
	authHandler := handlers.NewAuthHandler(usersSvc)

	bookingsSvc := services.NewBookingsService(database)
	bookingsHandler := handlers.NewBookingsHandler(bookingsSvc)

	stats := httpstat.NewStatistics(time.Minute)
	stats.StartWindowReset()

	r := mux.NewRouter()
	r.HandleFunc("/debug/stats", httpstat.Handler(stats)).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(httpstat.Middleware(stats))

	api.HandleFunc("/auth/register", authHandler.Register).Methods(http.MethodPost)
	api.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)

	api.HandleFunc("/hotels", hotelsHandler.GetHotelsHandler).Methods(http.MethodGet)
	api.HandleFunc("/hotels/{id}", hotelsHandler.GetHotelByIDHandler).Methods(http.MethodGet)
	api.HandleFunc("/hotels/{id}/rooms", hotelsHandler.GetHotelRoomsHandler).Methods(http.MethodGet)

	protected := api.NewRoute().Subrouter()
	protected.Use(middleware.Auth(redisClient))
	protected.HandleFunc("/user/bookings", bookingsHandler.GetMyBookings).Methods(http.MethodGet)
	protected.HandleFunc("/bookings", bookingsHandler.CreateBooking).Methods(http.MethodPost)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
