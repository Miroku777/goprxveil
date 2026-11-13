package wildberries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/Miroku77/goprxveil/internal/config"
	"github.com/Miroku77/goprxveil/internal/models"
)

type WBParser struct{}

// stratagy collect wb
func (w *WBParser) ParseAll(url string) (models.Product, int64, error) {
	tctx, cleanup, err := config.NewBrowserCxt()
	if err != nil {
		return models.Product{}, 0, err
	}
	defer cleanup()
	if err := chromedp.Do(tctx, chromedp.Navigate(url)); err != nil {
		return models.Product{}, 0, err
	}
	return MergeCardAndDetailResponseInProduct(tctx)
}

func (w *WBParser) ParsePrice(tctx context.Context, url string) (int64, error) {
	if err := chromedp.Do(tctx, chromedp.Navigate(url)); err != nil {
		return 0, err
	}
	return ListenerDetailPrice(tctx)
}

func MergeCardAndDetailResponseInProduct(tctx context.Context) (product models.Product, price int64, err error) {
	card, detail, err := ListenerCardAndDetailResponse(tctx)
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

func ListenerDetailPrice(tctx context.Context) (price int64, err error) {
	_, err = chromedp.Call(tctx, network.Enable, network.EnableParams{})
	if err != nil {
		return 0, err
	}
	detail := models.DetailResponse{}
	events := chromedp.Events(tctx, network.ResponseReceived)
	for ev, err := range events {
		if err != nil {
			log.Println(err)
			continue
		}
		if strings.Contains(ev.Response.URL, "/__internal/u-card/cards/v4/detail") {
			log.Println("found detail.json!")
			time.Sleep(time.Second * 2)
			body, err := chromedp.Call(tctx, network.GetResponseBody, network.GetResponseBodyParams{
				RequestID: ev.RequestID,
			})
			if err != nil {
				log.Println(err)
			}
			log.Println(string(body.Body))
			if err := json.Unmarshal(body.Body, &detail); err != nil {
				log.Println(err)
			} else if len(detail.Products) == 0 || len(detail.Products[0].Sizes) == 0 {
				return 0, errors.New("there is no product price")
			} else {
				price = detail.Products[0].Sizes[0].Price.Product
				break
			}
		}
	}
	return price, nil
}

// cards.json
func ListenerCardAndDetailResponse(tctx context.Context) (card models.CardResponse, detail models.DetailResponse, err error) {
	_, err = chromedp.Call(tctx, network.Enable, network.EnableParams{})
	if err != nil {
		log.Println(err)
	}
	events := chromedp.Events(tctx, network.ResponseReceived)

	for ev, err := range events {
		if err != nil {
			log.Println(err)
			continue
		}
		//log.Println(ev.Response.URL)
		if strings.Contains(ev.Response.URL, "card.json") {
			log.Println("found card.json!")
			time.Sleep(time.Second * 2)
			body, err := chromedp.Call(tctx, network.GetResponseBody, network.GetResponseBodyParams{
				RequestID: ev.RequestID,
			})
			if err != nil {
				log.Println("receivid body card:", err)
				continue
			}
			log.Println(string(body.Body))
			if err := json.Unmarshal(body.Body, &card); err != nil {
				log.Println(err)
				continue
			}
		}
		if strings.Contains(ev.Response.URL, "/__internal/u-card/cards/v4/detail") {
			log.Println("found detail.json!")
			time.Sleep(time.Second * 2)
			body, err := chromedp.Call(tctx, network.GetResponseBody, network.GetResponseBodyParams{
				RequestID: ev.RequestID,
			})
			if err != nil {
				log.Println("receiving body detail:", err)
				continue
			}
			log.Println(string(body.Body))
			if err = json.Unmarshal(body.Body, &detail); err != nil {
				log.Println(err)
				continue
			}
		}
		if card.NmId != 0 && len(detail.Products) > 0 {
			return card, detail, nil
		}
	}
	return card, detail, fmt.Errorf("card.json or detail.json is empty")
}
