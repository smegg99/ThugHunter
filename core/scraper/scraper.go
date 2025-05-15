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

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"gorm.io/gorm"

	"smuggr.xyz/thughunter/common/models"
	"smuggr.xyz/thughunter/core/datastore"
)

func LaunchUpdater(query string) (newCount, updCount int) {
	u := "https://platform.censys.io/search?q=" + url.QueryEscape(query)

	userDataDir := "./cdp-profile"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("enable-automation", false),
		// persistent profile so cookies/extensions look real
		chromedp.UserDataDir(userDataDir),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	if err := chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(`
            Object.defineProperty(navigator, 'webdriver', {
                get: () => false
            });
            const ua = navigator.userAgent.replace(/HeadlessChrome/, 'Chrome');
            Object.defineProperty(navigator, 'userAgent', {
                get: () => ua
            });
        `, nil),
	); err != nil {
		log.Fatalf("failed to patch webdriver: %v", err)
	}

	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatalf("network enable: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Navigate(u)); err != nil {
		log.Fatalf("navigate error: %v", err)
	}

	fmt.Println("Please complete login/navigation in the browser and press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	var html string
	if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &html, chromedp.ByQuery)); err != nil {
		log.Fatalf("extract html: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatalf("goquery parse: %v", err)
	}

	// extract host cards
	doc.Find("div[data-testid='hostDetailsCard']").Each(func(_ int, s *goquery.Selection) {
		h := models.Host{Services: make(models.JSONServiceMap)}
		// IP & hostname
		link := s.Find("a[title^='View']")
		href, _ := link.Attr("href")
		h.IP = filepath.Base(strings.Split(href, "?")[0])
		h.Hostname = s.Find("._typographyDefault_80wah_2").First().Text()

		// labels
		s.Find("div[data-testid='label-list'] span._label_13xbf_14").Each(func(_ int, lab *goquery.Selection) {
			h.Labels = append(h.Labels, lab.Text())
		})

		// location
		h.Location = s.Find("table.qI1Kw td._typographyDefault_80wah_2").First().Text()

		// services
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

		// upsert
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
