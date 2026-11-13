package wildberries

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

// Deprecated!
func SckrapParseProductPage(tctx context.Context) (sellerName, name, rating, price string) {
	err := chromedp.Do(tctx, chromedp.WaitReady(".priceBlockFinalPrice--yEKcc"))
	if err != nil {
		log.Println("Ошибка:", err)
		return
	}
	// <span class="mo-typography mo-typography_variant_description mo-typography_variable-weight_description mo-typography_variable sellerInfoNameDefaultText--_ePyu">Автозапчасти_Москвич, ИЖ, ОДА</span>
	sellerName, err = chromedp.Run(tctx, chromedp.Text(`[class*="sellerInfoNameDefaultText"]`))
	if err != nil {
		log.Println("Ошибка:", err)
		return
	}
	name, err = chromedp.Run(tctx, chromedp.Text(".productTitle--ktP8Z"))
	if err != nil {
		log.Println("Ошибка:", err)
		return
	}
	rating, err = chromedp.Run(tctx, chromedp.Text(`[class*="ratingNumber"]`))
	if err != nil {
		log.Println("Ошибка:", err)
		return
	}

	price, err = chromedp.Run(tctx, chromedp.Text(".priceBlockFinalPrice--yEKcc"))
	if err != nil {
		log.Println("Ошибка:", err)
		return
	}
	// log.Printf("Название товара: %s\nОценка: %s\nИмя продавца: %s\nцена товара: %s\n", name, rating, sellerName, price)
	return
}
