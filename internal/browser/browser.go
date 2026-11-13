package browser

import (
	"database/sql"
	"log"
	"math/rand"
	"net/url"
	"time"

	"github.com/Miroku77/goprxveil/internal/config"
	"github.com/Miroku77/goprxveil/internal/parser"
	"github.com/Miroku77/goprxveil/internal/parser/ozon"
	"github.com/Miroku77/goprxveil/internal/parser/wildberries"
	"github.com/Miroku77/goprxveil/internal/storage"
)

// FirstAddProductToDB parses the initial product data (name, seller, price, card.json)
// from the given URL and saves it to the database.
// Called when the url is added for the first time
func FirstAddProductToDB(db *sql.DB, domen string, urlId int, urlstr string) {
	var itemParser parser.Parser
	if domen == "www.wildberries.ru" {
		itemParser = &wildberries.WBParser{}
	} else if domen == "www.ozon.ru" {
		itemParser = &ozon.OzonParser{}
	} else if itemParser == nil {
		log.Println("Undefined domen url")
		return
	}
	product, price, err := itemParser.ParseAll(urlstr)
	if err != nil {
		log.Println(err)
		return
	}
	product.UrlId = urlId

	productId, err := storage.SaveProductToDB(db, product)
	if err != nil {
		log.Println(err)
		return
	}
	if err = storage.SaveToPriceHistory(db, productId, price, time.Now()); err != nil {
		log.Println(err)
	}
}

func JobCallShop(db *sql.DB) {
	for {
		sleepSecond := int((1.0 * 60.0) + rand.Float64()*(1.0*60.0))
		time.Sleep(time.Duration(sleepSecond) * time.Second)

		log.Println("sleepMinutes:", sleepSecond)
		productToParseUpdate, err := storage.FindProductByLastChecked(db, sleepSecond+2)
		if err != nil {
			log.Println("Db not responce: ", err)
			continue
		}
		if len(productToParseUpdate) == 0 {
			log.Println("Everything up-to-date")
			continue
		}

		func() {
			page, cleanup, err := config.NewBrowserContext()
			if err != nil {
				log.Println(err)
				return
			}
			defer cleanup()

			for _, product := range productToParseUpdate {
				time.Sleep(time.Duration(3+rand.Intn(7)) * time.Second)

				u, err := url.Parse(product.Url)
				if err != nil {
					log.Println(err)
					continue
				}
				domen := u.Host
				var itemParser parser.Parser
				if domen == "www.wildberries.ru" {
					itemParser = &wildberries.WBParser{}
				} else if domen == "www.ozon.ru" {
					itemParser = &ozon.OzonParser{}
				} else if itemParser == nil {
					log.Println("Undefined domen url")
					continue
				}

				price, err := itemParser.ParsePrice(page, product.Url)
				if err != nil {
					log.Println(err)
					continue
				}
				if err = storage.RecordPriceUpdate(db, price, time.Now(), product.Id); err != nil {
					log.Println(err)
					continue
				}
			}
		}()
	}
}
