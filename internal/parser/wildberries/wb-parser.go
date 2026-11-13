package wildberries

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/Miroku77/goprxveil/internal/config"
	"github.com/Miroku77/goprxveil/internal/models"
)

type WBParser struct{}

func (w *WBParser) ParseAll(urlstr string) (models.Product, int64, error) {
	page, cleanup, err := config.NewBrowserContext()
	if err != nil {
		return models.Product{}, 0, err
	}
	defer cleanup()
	if err := page.Navigate(urlstr); err != nil {
		return models.Product{}, 0, err
	}
	return MergeCardAndDetailResponseInProduct(page)
}

func (w *WBParser) ParsePrice(page *rod.Page, urlstr string) (int64, error) {
	if err := page.Navigate(urlstr); err != nil {
		return 0, err
	}
	return ListenerDetailPrice(page)
}

func MergeCardAndDetailResponseInProduct(page *rod.Page) (product models.Product, price int64, err error) {
	card, detail, err := ListenerCardAndDetailResponse(page)
	if err != nil {
		return product, 0, err
	}
	product = models.Product{
		SellerName:  detail.Products[0].SellerName,
		Name:        detail.Products[0].Name,
		Rating:      detail.Products[0].Rating,
		Description: card.Description,
		Options:     card.Options,
	}
	if len(detail.Products[0].Sizes) > 0 {
		price = detail.Products[0].Sizes[0].Price.Product
	} else {
		return product, 0, errors.New("there is no product price")
	}
	return product, price, nil
}

func ListenerDetailPrice(page *rod.Page) (price int64, err error) {
	err = proto.NetworkEnable{}.Call(page)
	if err != nil {
		return 0, err
	}
	detail := models.DetailResponse{}
	var innerErr error

	page.EachEvent(func(ev *proto.NetworkResponseReceived) bool {
		if strings.Contains(ev.Response.URL, "/__internal/u-card/cards/v4/detail") {
			log.Println("found detail.json!")
			time.Sleep(time.Second * 2)
			body, err := proto.NetworkGetResponseBody{RequestID: ev.RequestID}.Call(page)
			if err != nil {
				log.Println(err)
			}
			log.Println(body.Body)
			if err := json.Unmarshal([]byte(body.Body), &detail); err != nil {
				log.Println(err)
			} else if len(detail.Products) == 0 || len(detail.Products[0].Sizes) == 0 {
				innerErr = errors.New("there is no product price")
				return true
			} else {
				price = detail.Products[0].Sizes[0].Price.Product
				return true
			}
		}
		return false
	})()
	if innerErr != nil {
		return 0, innerErr
	}
	return price, nil
}

// cards.json
func ListenerCardAndDetailResponse(page *rod.Page) (card models.CardResponse, detail models.DetailResponse, err error) {
	err = proto.NetworkEnable{}.Call(page)
	if err != nil {
		log.Println(err)
	}
	page.EachEvent(func(ev *proto.NetworkResponseReceived) bool {
		if strings.Contains(ev.Response.URL, "card.json") {
			log.Println("found card.json!")
			time.Sleep(time.Second * 2)
			body, err := proto.NetworkGetResponseBody{RequestID: ev.RequestID}.Call(page)
			if err != nil {
				log.Println("receivid body card:", err)
				return false
			}
			log.Println(body.Body)
			if err := json.Unmarshal([]byte(body.Body), &card); err != nil {
				log.Println(err)
				return false
			}
		}
		if strings.Contains(ev.Response.URL, "/__internal/u-card/cards/v4/detail") {
			log.Println("found detail.json!")
			time.Sleep(time.Second * 2)
			body, err := proto.NetworkGetResponseBody{RequestID: ev.RequestID}.Call(page)
			if err != nil {
				log.Println("receiving body detail:", err)
				return false
			}
			log.Println(body.Body)
			if err = json.Unmarshal([]byte(body.Body), &detail); err != nil {
				log.Println(err)
				return false
			}
		}
		if card.NmId != 0 && len(detail.Products) > 0 {
			return true
		}
		return false
	})()

	if card.NmId != 0 && len(detail.Products) > 0 {
		return card, detail, nil
	}
	return card, detail, fmt.Errorf("card.json or detail.json is empty")
}
