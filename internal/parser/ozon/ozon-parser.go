package ozon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Miroku77/goprxveil/internal/models"
	"github.com/go-rod/rod"
	"github.com/mxschmitt/playwright-go"
)

var (
	scraperInstance *OzonPlaywright
	scraperMutex    sync.Mutex
)

type OzonPlaywright struct {
	pw         *playwright.Playwright
	context    playwright.BrowserContext
	page       playwright.Page
	challenged bool
}

// GetScraper returns the global OzonPlaywright instance.
// If it doesn't exist or the challenge hasn't been passed,
// it initializes a new browser context and launches Chromium.
func GetScraper() (*OzonPlaywright, error) {
	scraperMutex.Lock()
	defer scraperMutex.Unlock()

	if scraperInstance != nil && scraperInstance.challenged {
		return scraperInstance, nil
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	userDataDir := filepath.Join(os.Getenv("HOME"), ".ozon-mcp-userdata")

	args := []string{
		"--disable-blink-features=AutomationControlled",
		"--no-sandbox",
		// "--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
		"--mute-audio",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		"--disable-background-networking",
		// "--window-position=-32000,-32000", // Окно за экраном
	}

	browserPath := "/usr/bin/google-chrome"
	if _, err := os.Stat(browserPath); os.IsNotExist(err) {
		browserPath = "/usr/bin/chromium"
	}
	context, err := pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath: playwright.String(browserPath), // <-- ИСПОЛЬЗУЕМ АВТООПРЕДЕЛЕНИЕ
		Headless:       playwright.Bool(false),
		Args:           args,
		Viewport:       &playwright.Size{Width: 1920, Height: 1080},
	})
	if err != nil {
		return nil, err
	}

	scraperInstance = &OzonPlaywright{
		pw:      pw,
		context: context,
	}

	return scraperInstance, nil
}

// ensureContext ensures the Playwright context is initialized and the anti-bot challenge is passed by navigating
// to the Ozon homepage.
func (o *OzonPlaywright) ensureContext() error {
	if o.challenged {
		return nil
	}

	pages := o.context.Pages()
	if len(pages) > 0 {
		o.page = pages[0]
	} else {
		var err error
		o.page, err = o.context.NewPage()
		if err != nil {
			return err
		}
	}

	_, err := o.page.Goto("https://www.ozon.ru/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(90000),
	})
	if err != nil {
		return err
	}
	time.Sleep(12 * time.Second)

	title, err := o.page.Title()
	if err == nil && regexp.MustCompile(`(?i)antibot|ограничен|доступ|нет соединения`).MatchString(title) {
		return fmt.Errorf("challenge не пройден, title: %s", title)
	}
	o.challenged = true
	return nil
}

// GetPrice retrieves the price, product name, rating, and description from the Ozon API using the product ID.
func (o *OzonPlaywright) fetchJson(path string) (text string, err error) {
	if err = o.ensureContext(); err != nil {
		return "", err
	}

	apiURL := "https://www.ozon.ru/api/composer-api.bx/page/json/v2?url=" + url.QueryEscape(path)

	res, err := o.page.Evaluate(`async (url) => {
		const r = await fetch(url, { headers: { accept: "application/json" } });
		return { status: r.status, text: await r.text() };
	}`, apiURL)
	if err != nil {
		return "", err
	}

	data, ok := res.(map[string]any)
	if !ok {
		return "", errors.New("invalid JS response")
	}

	var status int
	switch v := data["status"].(type) {
	case float64:
		status = int(v)
	case int:
		status = v
	}
	text = data["text"].(string)

	if status != 200 {
		if status == 403 {
			o.challenged = false
		}
		return "", fmt.Errorf("Ozon вернул HTTP %d", status)
	}
	return text, nil
}

func (o *OzonPlaywright) GetPrice(productID string) (price int64, productName string, rating float64, description string, err error) {
	path := "/product/" + productID + "/"
	jsonText, err := o.fetchJson(path)
	if err != nil {
		return 0, "", 0, "", err
	}

	var data map[string]any
	if err = json.Unmarshal([]byte(jsonText), &data); err != nil {
		return 0, "", 0, "", err
	}
	widgetStates, _ := data["widgetStates"].(map[string]any)

	var priceStr string

	for key, val := range widgetStates {
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		var widget map[string]any
		json.Unmarshal([]byte(strVal), &widget)

		if strings.HasPrefix(key, "webPrice-") {
			if p, ok := widget["price"].(string); ok {
				priceStr = p
			}
		}
		if strings.HasPrefix(key, "webProductHeading-") {
			if t, ok := widget["title"].(string); ok {
				productName = t
			}
		}
		if strings.HasPrefix(key, "webReviewProductScore-") {
			if s, ok := widget["score"].(float64); ok {
				rating = s
			}
		}
	}

	path2 := path + "?layout_container=pdpPage2column&layout_page_index=2"
	jsonText2, err2 := o.fetchJson(path2)
	if err2 == nil {
		var data2 map[string]any
		json.Unmarshal([]byte(jsonText2), &data2)
		if ws2, ok := data2["widgetStates"].(map[string]any); ok {
			for key, val := range ws2 {
				if strings.HasPrefix(key, "webDescription-") {
					strVal, _ := val.(string)
					var widget map[string]any
					json.Unmarshal([]byte(strVal), &widget)
					if t, ok := widget["text"].(string); ok {
						description = t
					}
				}
			}
		}
	}

	cleanPrice := regexp.MustCompile(`\D`).ReplaceAllString(priceStr, "")
	price, err = strconv.ParseInt(cleanPrice, 10, 64)
	if err != nil {
		return 0, "", 0, "", err
	}
	price = price * 100
	return price, productName, rating, description, nil
}

type OzonParser struct{}

// extractProductID extracts the numeric product ID from the given Ozon URL.
func extractProductID(urlstr string) (string, error) {
	re := regexp.MustCompile(`-(\d+)/`)
	matches := re.FindStringSubmatch(urlstr)
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`/product/(\d+)/`)
		matches2 := re2.FindStringSubmatch(urlstr)
		if len(matches2) < 2 {
			return "", errors.New("cannot extract product ID")
		}
		return matches2[1], nil
	}
	return matches[1], nil
}

// GetFullProductInfo navigates to the product page, extracts JSON-LD data from the DOM,
// and returns the price, name, rating, and description.
func (o *OzonPlaywright) GetFullProductInfo(productID string) (price int64, name string, rating float64, desc string, err error) {
	if err := o.ensureContext(); err != nil {
		return 0, "", 0, "", err
	}

	url := "https://www.ozon.ru/product/" + productID + "/"
	_, err = o.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return 0, "", 0, "", err
	}

	time.Sleep(3 * time.Second) // waiting render
	res, err := o.page.Evaluate(`() => document.querySelector('script[type="application/ld+json"]')?.innerText || ""`)
	if err != nil || res == nil {
		return 0, "", 0, "", errors.New("не удалось найти JSON-LD")
	}

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Rating      struct {
			Value string `json:"ratingValue"`
		} `json:"aggregateRating"`
		Offers struct {
			Price string `json:"price"`
		} `json:"offers"`
	}

	jsonText := res.(string)
	if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
		return 0, "", 0, "", err
	}

	cleanPrice := regexp.MustCompile(`\D`).ReplaceAllString(data.Offers.Price, "")
	price, err = strconv.ParseInt(cleanPrice, 10, 64)
	if err != nil {
		return 0, "", 0, "", err
	}
	price = price * 100
	rating, err = strconv.ParseFloat(data.Rating.Value, 64)
	if err != nil {
		return 0, "", 0, "", err
	}
	return price, data.Name, rating, data.Description, nil
}

// ParseAll parses the full product information from the given URL and returns the product model along with its price.
func (w *OzonParser) ParseAll(urlstr string) (models.Product, int64, error) {
	productID, _ := extractProductID(urlstr)
	scraper, err := GetScraper()
	if err != nil {
		fmt.Println("err start PLAYWRIGHT:", err)
		return models.Product{}, 0, err
	}

	price, name, rating, desc, err := scraper.GetFullProductInfo(productID)
	if err != nil {
		return models.Product{}, 0, err
	}

	product := models.Product{
		SellerName:  "-",
		Name:        name,
		Rating:      rating,
		Description: desc,
		Options:     json.RawMessage(`[]`),
		LastChecked: time.Now(),
	}

	return product, price, nil
}

// ParsePrice retrieves the current price of the product from the given URL.
// It accepts a rod.Page but ignores it, using Playwright instead.
func (w *OzonParser) ParsePrice(page *rod.Page, urlstr string) (int64, error) {
	productID, err := extractProductID(urlstr)
	if err != nil {
		return 0, err
	}

	scraper, err := GetScraper()
	if err != nil {
		return 0, err
	}
	price, _, _, _, err := scraper.GetPrice(productID)
	if err != nil {
		return price, err
	}

	return price, err
}
