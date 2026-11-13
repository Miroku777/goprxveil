package storage

import (
	"database/sql"
	"time"
)

// SaveToPriceHistory save new price and time by id product
func SaveToPriceHistory(db *sql.DB, productId int, price int64, createdAt time.Time) (err error) {
	_, err = db.Exec(`
        INSERT INTO price_history (product_id, price, created_at) 
        VALUES ($1, $2, $3) 
    `, productId, price, createdAt)
	return
}
