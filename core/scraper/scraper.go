// core/scraper/scraper.go
package scraper

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"gorm.io/gorm"

	"smuggr.xyz/thughunter/common/models"
	"smuggr.xyz/thughunter/core/datastore"
)

func LaunchUpdater(query string) (newCount, updCount int) {
	u := "https://platform.censys.io/search?q=" + url.QueryEscape(query)
	userDataDir := os.Getenv("CDP_USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = "./cdp-profile"
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(userDataDir),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("no-first-run", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	stealthScript := `
	Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
	window.chrome = { runtime: {} };
	Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
	Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3] });
	Object.defineProperty(navigator, 'platform', { get: () => 'Win32' });
	const originalQuery = window.navigator.permissions.query;
	window.navigator.permissions.query = (parameters) => (
		parameters.name === 'notifications'
			? Promise.resolve({ state: Notification.permission })
			: originalQuery(parameters)
	);
	const newUA = navigator.userAgent.replace('HeadlessChrome', 'Chrome');
	Object.defineProperty(navigator, 'userAgent', { get: () => newUA });
	`

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(stealthScript, nil),
	); err != nil {
		log.Fatalf("failed to inject stealth JS: %v", err)
	}

	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatalf("network enable: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Navigate(u)); err != nil {
		log.Fatalf("navigate error: %v", err)
	}

	var avatarExists bool
	fmt.Println("Checking login status...")
	
	if err := chromedp.Run(ctx,
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`document.querySelector('div[class*="_avatar_"][role="img"]') !== null`, &avatarExists),
	); err != nil {
		log.Printf("Error checking login status: %v", err)
	}

	if !avatarExists {
		fmt.Println("Not logged in. Please complete login/navigation in the browser and press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	} else {
		fmt.Println("Already logged in, continuing automatically...")
	}

	var html string
	if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &html)); err != nil {
		log.Fatalf("extract html: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatalf("goquery parse: %v", err)
	}

	doc.Find("div[data-testid='hostDetailsCard']").Each(func(_ int, s *goquery.Selection) {
		h := models.Host{Services: make(models.JSONServiceMap)}

		link := s.Find("a[title^='View']")
		href, _ := link.Attr("href")
		h.IP = filepath.Base(strings.Split(href, "?")[0])
		h.Hostname = s.Find("._typographyDefault_80wah_2").First().Text()

		s.Find("div[data-testid='label-list'] span._label_13xbf_14").Each(func(_ int, lab *goquery.Selection) {
			h.Labels = append(h.Labels, lab.Text())
		})

		h.Location = s.Find("table.qI1Kw td._typographyDefault_80wah_2").First().Text()

		s.Find(".BNqaQ a[title]").Each(func(_ int, svc *goquery.Selection) {
			t := svc.Find("span._label_13xbf_14").Text()
			parts := strings.Split(t, " /")
			if len(parts) >= 1 {
				srv := strings.TrimSpace(parts[len(parts)-1])
				portStr := strings.Fields(parts[0])[0]
				p, _ := strconv.Atoi(portStr)
				h.Services[srv] = p
			}
		})

		var existing models.Host
		r := datastore.DB.First(&existing, "ip = ?", h.IP)
		if r.Error != nil && r.Error == gorm.ErrRecordNotFound {
			datastore.DB.Create(&h)
			newCount++
		} else if r.Error == nil {
			h.IP = existing.IP
			datastore.DB.Save(&h)
			updCount++
		}
	})

	return
}
