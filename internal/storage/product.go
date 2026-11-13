package storage

import (
	"database/sql"
	"time"

	"github.com/Miroku77/goprxveil/internal/models"
)

// UpdateLastCheckedToProduct update product long ago - set time now by id product
func RecordPriceUpdate(db *sql.DB, price int64, createdAt time.Time, productId int) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
        INSERT INTO price_history (product_id, price, created_at) 
        VALUES ($1, $2, $3) 
    `, productId, price, createdAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE products SET last_checked = $1 WHERE id = $2`, time.Now(), productId)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// SaveProductToDB save a product to the database and return id it's records
func SaveProductToDB(db *sql.DB, p models.Product) (id int, err error) {
	err = db.QueryRow(`
        INSERT INTO products (url_id, seller_name, name, rating, description, options) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        ON CONFLICT (url_id) DO UPDATE SET 
            seller_name = EXCLUDED.seller_name,
            name = EXCLUDED.name,
            rating = EXCLUDED.rating,
            description = EXCLUDED.description,
            options = EXCLUDED.options
        RETURNING id
    `, p.UrlId, p.SellerName, p.Name, p.Rating, p.Description, p.Options).Scan(&id)
	return id, err
}

// FindProductByLastChecked returns products that haven't been checked for a while
func FindProductByLastChecked(db *sql.DB, sleepSecond int) (productToParseUpdate []models.ProductToParseUpdate, err error) {
	rows, err := db.Query(`
        SELECT p.id, u.url
        FROM products p
        JOIN urls u ON p.url_id = u.id
        WHERE u.active = true AND p.last_checked < NOW() - ($1 * INTERVAL '1 second')
    `, sleepSecond)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productToParseUpdate = []models.ProductToParseUpdate{}
	for rows.Next() {
		var p models.ProductToParseUpdate
		if err = rows.Scan(&p.Id, &p.Url); err != nil {
			return nil, err
		}
		productToParseUpdate = append(productToParseUpdate, p)
	}
	return productToParseUpdate, rows.Err()
}
