package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type DBConfig struct {
	Host string
	Port int
	Name string
	User string
	Pass string
}

var (
	dbOnce     sync.Once
	dbInstance *DB
	initError  error
)

func Init(config *DBConfig) (*DB, error) {
	dbOnce.Do(func() {
		log.Println("Start init DB")
		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.User, config.Pass, config.Host, config.Port, config.Name)

		rawDB, err := sql.Open("postgres", connStr)
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
