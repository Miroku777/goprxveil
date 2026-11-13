package storage

import (
	"database/sql"

	"github.com/Miroku77/goprxveil/internal/models"
)

// SaveGroupForUser create group for user and return it's ID
func SaveGroupForUser(db *sql.DB, nameGroup string, userId int) (id int, err error) {
	err = db.QueryRow(
		"INSERT INTO groups (name, user_id) VALUES ($1, $2) RETURNING id",
		nameGroup, userId).Scan(&id)
	return id, err
}

// FindAllFromGroup returns all groups for the user
func FindAllFromGroup(db *sql.DB, userId int) (groups []models.Groups, err error) {
	rows, err := db.Query(`SELECT id, name FROM groups WHERE user_id = $1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Groups
		if err = rows.Scan(&group.Id, &group.Name); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func DeleteFromGroupById(db *sql.DB, id, userId int) (err error) {
	_, err = db.Exec(`DELETE FROM groups WHERE id = $1 AND user_id = $2`, id, userId)
	return err
}
