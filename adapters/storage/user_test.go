package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	t.Run("successfully creates user", func(t *testing.T) {
		store := setupTestStore(t)

		userID, err := store.CreateUser("test@email.com", "password_hash")
		assert.NoError(t, err)
		assert.NotZero(t, userID)

		var id int
		var email, passwordHash string
		err = store.db.QueryRow("SELECT id, email, password_hash FROM users").Scan(&id, &email, &passwordHash)
		assert.NoError(t, err)
		assert.Equal(t, userID, id)
		assert.Equal(t, "test@email.com", email)
		assert.Equal(t, "password_hash", passwordHash)
	})
	t.Run("fails when email already exists", func(t *testing.T) {
		store := setupTestStore(t)

		_, err := store.CreateUser("test@email.com", "password_hash")
		assert.NoError(t, err)
		_, err = store.CreateUser("test@email.com", "password_hash")
		assert.Error(t, err)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("successfully get user by email", func(t *testing.T) {
		store := setupTestStore(t)

		userID, err := store.CreateUser("test@email.com", "password_hash")
		assert.NoError(t, err)
		assert.NotZero(t, userID)

		user, err := store.GetUserByEmail("test@email.com")
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@email.com", user.Email)
		assert.Equal(t, "password_hash", user.PasswordHash)
	})
	t.Run("fails when user not found", func(t *testing.T) {
		store := setupTestStore(t)

		_, err := store.GetUserByEmail("test@email.com")
		assert.Error(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("successfully get user by id", func(t *testing.T) {
		store := setupTestStore(t)

		userID, err := store.CreateUser("test@email.com", "password_hash")
		assert.NoError(t, err)
		assert.NotZero(t, userID)

		user, err := store.GetUserByID(userID)
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@email.com", user.Email)
		assert.Equal(t, "password_hash", user.PasswordHash)
	})
	t.Run("fails when user not found", func(t *testing.T) {
		store := setupTestStore(t)

		_, err := store.GetUserByID(99999)
		assert.Error(t, err)
	})
}

func TestEmailExists(t *testing.T) {
	t.Run("successfully check email", func(t *testing.T) {
		store := setupTestStore(t)

		userID, err := store.CreateUser("test@email.com", "password_hash")
		assert.NoError(t, err)
		assert.NotZero(t, userID)

		exists, err := store.EmailExists("test@email.com")
		assert.NoError(t, err)
		assert.True(t, exists)

	})
	t.Run("get false when email not found", func(t *testing.T) {
		store := setupTestStore(t)

		exists, err := store.EmailExists("test@email.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
