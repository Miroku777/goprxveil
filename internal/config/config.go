package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)

func ConnectDB() (db *sql.DB, err error) {
	pgConfig := "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable port=5432"

	db, err = sql.Open("pgx", pgConfig)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// GetOzonCookies launches a Chrome browser via go-rod, navigates to the specified URL,
// waits for the page to load, and returns all cookies as a single semicolon-separated string.
func GetOzonCookies(urlstr string) (string, error) {
	l := launcher.New().
		Bin("/usr/bin/google-chrome").
		UserDataDir("/tmp/chrome-profile-ozon").
		Set("window-size", "1920,1080").
		Headless(false).
		Set("user-agent", UserAgent).
		Set("disable-blink-features", "AutomationControlled")

	url, err := l.Launch()
	if err != nil {
		return "", err
	}
	defer l.Cleanup()

	browser := rod.New().ControlURL(url)
	if err = browser.Connect(); err != nil {
		return "", err
	}
	defer browser.MustClose()

	page := stealth.MustPage(browser).Timeout(30 * time.Second)
	page.MustSetViewport(1920, 1080, 1, false)

	if err := page.Navigate(urlstr); err != nil {
		return "", err
	}
	page.MustWaitLoad()
	time.Sleep(5 * time.Second)

	cookies, err := page.Cookies([]string{urlstr})
	if err != nil {
		return "", err
	}

	var cookieStr string
	for _, c := range cookies {
		cookieStr += fmt.Sprintf("%s=%s; ", c.Name, c.Value)
	}
	return cookieStr, nil
}

// NewBrowserContext initializes a new Chrome browser instance with a persistent profile, custom flags,
// and a User-Agent. It returns a rod.Page, a cleanup function, and an error.
func NewBrowserContext() (page *rod.Page, cleanup func(), err error) {
	l := launcher.New().
		Bin("/usr/bin/google-chrome").
		UserDataDir("/tmp/ozon-profile-full").
		Set("window-size", "1920,1080").
		Headless(false).
		Set("user-agent", UserAgent).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-infobars").
		Set("no-first-run").
		Set("no-default-browser-check").
		Leakless(false)

	url, err := l.Launch()
	if err != nil {
		l.Cleanup()
		return nil, nil, err
	}

	browser := rod.New().ControlURL(url)
	if err = browser.Connect(); err != nil {
		l.Cleanup()
		return nil, nil, err
	}

	page = stealth.MustPage(browser)
	page.MustSetViewport(1920, 1080, 1, false)
	page = page.Timeout(time.Hour)

	cleanup = func() {
		browser.MustClose()
		l.Cleanup()
	}
	return page, cleanup, nil
}

// Deprecated
func NewBrowserCxt() (tctx context.Context, cleanup func(), err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir("/tmp/chrome-profile"),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("headless", false),
		chromedp.UserAgent(UserAgent),

		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	tctx, tcancel := context.WithTimeout(ctx, time.Hour)
	cleanup = func() {
		allocCancel()
		cancel()
		tcancel()
	}
	err = chromedp.Do(tctx, chromedp.EmulateViewport(1920, 1080)) //время работы окна
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return tctx, cleanup, nil
}
