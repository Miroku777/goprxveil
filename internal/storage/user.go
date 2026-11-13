package storage

import (
	"database/sql"

	"github.com/Miroku77/goprxveil/internal/models"
)

type Users struct {
	Id       int //OneToMany
	Email    string
	Password string
}

// SaveUser save new users and return this id
func SaveUser(db *sql.DB, user models.Users) (newUserId int, err error) {
	err = db.QueryRow(`INSERT INTO users(email, password) VALUES($1, $2) RETURNING id`, user.Email, user.Password).Scan(&newUserId)
	return newUserId, err
}

// FindUserByEmail return struct users by email
func FindUserByEmail(db *sql.DB, email string) (userId int, passwd string, err error) {
	err = db.QueryRow(`SELECT id, password FROM users WHERE email = $1`, email).Scan(&userId, &passwd)
	return userId, passwd, err
}
