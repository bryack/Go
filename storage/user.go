package storage

import (
	"database/sql"
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStorage interface {
	CreateUser(email string, passwordHash string) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	EmailExists(email string) (bool, error)
}

func (ds *DatabaseStorage) CreateUser(email, passwordHash string) (int, error) {
	result, err := ds.db.Exec(
		"INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		email, passwordHash,
	)
	if err != nil {
		return 0, mapSQLiteError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, mapSQLiteError(err)
	}
	return int(id), nil
}

func (ds *DatabaseStorage) GetUserByEmail(email string) (*User, error) {
	var user User
	err := ds.db.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, mapSQLiteError(err)
	}

	return &user, nil
}

func (ds *DatabaseStorage) GetUserByID(id int) (*User, error) {
	var user User
	err := ds.db.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, mapSQLiteError(err)
	}

	return &user, nil
}

func (ds *DatabaseStorage) EmailExists(email string) (exists bool, err error) {
	err = ds.db.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM users WHERE email = ?)",
		email,
	).Scan(&exists)

	if err != nil {
		return exists, mapSQLiteError(err)
	}

	return exists, nil
}
