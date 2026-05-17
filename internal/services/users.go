package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const sessionPrefix = "session:"
const sessionTTL = 24 * time.Hour

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already taken")
	ErrWrongPassword   = errors.New("wrong password")
)

type UsersService struct {
	db    *db.DB
	redis *redis.Client
}

func NewUsersService(db *db.DB, redis *redis.Client) *UsersService {
	return &UsersService{db: db, redis: redis}
}

func (s *UsersService) Register(ctx context.Context, email, password, name string) (models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.db.CreateUser(ctx, email, string(hash), name)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, err
	}

	return user, nil
}

func (s *UsersService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.db.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrWrongPassword
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}

	key := sessionPrefix + token
	if err := s.redis.Set(ctx, key, user.ID, sessionTTL).Err(); err != nil {
		return "", err
	}

	return token, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
