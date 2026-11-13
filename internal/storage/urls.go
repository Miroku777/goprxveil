package storage

import (
	"database/sql"
	"time"

	"github.com/Miroku77/goprxveil/internal/models"
)

type GroupUrls struct {
	GroupId int `json:"group_id"`
	UrlId   int `json:"url_id"`
}

type Urls struct {
	Id        int
	Url       string
	CreatedAt time.Time
	Name      string
	//for run sckrup
	Checked bool
	Active  bool
}

// SaveUrlForGroup creates or updates URL for group. For rederect in new group
func SaveUrlForGroup(db *sql.DB, urlstr string, groupId int) (id int, err error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	err = tx.QueryRow(`
		INSERT INTO urls (url, checked) VALUES ($1, 'false') 
		ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
		RETURNING id
	`, urlstr).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`
		INSERT INTO group_urls (group_id, url_id) VALUES ($1, $2) 
		ON CONFLICT DO NOTHING
	`, groupId, id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// FindAllUrlsFromGroup get all urls where checked = true from group_id
func FindAllUrlsFromGroup(db *sql.DB, group_id int) (urlsstr []Urls, err error) {
	rows, err := db.Query(`
		SELECT u.id, u.url, COALESCE(p.name, 'Загрузка...') as name, u.active
		FROM urls u
		JOIN group_urls gu ON u.id = gu.url_id
		LEFT JOIN products p ON u.id = p.url_id
		WHERE gu.group_id = $1
	`, group_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url Urls
		if err = rows.Scan(&url.Id, &url.Url, &url.Name, &url.Active); err != nil {
			return nil, err
		}
		urlsstr = append(urlsstr, url)
	}
	return urlsstr, rows.Err()
}

// DeleteFromUrlById deletes url by it's id and group id
func DeleteFromUrlById(db *sql.DB, urlId, groupId int) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`DELETE FROM group_urls WHERE url_id = $1 AND group_id = $2`, urlId, groupId)
	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM group_urls WHERE url_id = $1`, urlId).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = tx.Exec(`DELETE FROM urls WHERE id = $1`, urlId)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// NOT REFACTOR
// FindProductsByUrlid accepts url, return id product and struct analitics
func FindProductsByUrlid(db *sql.DB, urlId int) (productId int, analitics models.Analitics, err error) {
	err = db.QueryRow(`
		SELECT id, seller_name, name, rating, description, options FROM products WHERE url_id = $1`,
		urlId).Scan(&productId, &analitics.SellerName, &analitics.Name, &analitics.Rating, &analitics.Description,
		&analitics.Options)
	return productId, analitics, err
}

// NOT REFACTOR
// FindPriceHistoryByProductId accepts id product, return current price and history price
func FindPriceHistoryByProductId(db *sql.DB, productId int) (currentPrice string, history []models.ChartPoint, err error) {
	rows, err := db.Query(`
		SELECT price, created_at FROM price_history 
		WHERE product_id = $1 AND created_at >= CURRENT_DATE 
		ORDER BY created_at ASC
	`, productId)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var chPoint models.ChartPoint
		if err = rows.Scan(&chPoint.Price, &chPoint.Date); err != nil {
			return "", nil, err
		}
		history = append(history, chPoint)
	}
	if len(history) > 0 {
		currentPrice = history[len(history)-1].Price
	}
	return currentPrice, history, rows.Err()
}

// NOT REFACTOR
// UpdateActiveUrl reverse field active by id url
func UpdateActiveUrl(db *sql.DB, urlId int) (err error) {
	_, err = db.Exec(`UPDATE urls SET active = NOT active WHERE id = $1`, urlId)
	return err
}

type GroupComparisonUrls struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}

// get name and last price products, if in a this products there is group
func GetGroupPrices(db *sql.DB, groupId int) ([]GroupComparisonUrls, error) {
	rows, err := db.Query(`SELECT p.name, 
	(SELECT price FROM price_history WHERE product_id = p.id ORDER BY created_at DESC LIMIT 1) as last_price
		FROM products p 
		JOIN urls u ON p.url_id = u.id
		JOIN group_urls gu ON u.id = gu.url_id
		WHERE gu.group_id = $1 
	`, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GroupComparisonUrls
	for rows.Next() {
		var g GroupComparisonUrls
		if err = rows.Scan(&g.Name, &g.Price); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}
