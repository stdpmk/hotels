package db

import (
	"context"

	"github.com/stdpmk/hotels/internal/models"
)

func (db *DB) CreateUser(ctx context.Context, email, passwordHash, name string) (models.User, error) {
	var user models.User
	err := db.DB.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name, created_at`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
