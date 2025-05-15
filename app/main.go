// app/main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"smuggr.xyz/thughunter/common/config"
	"smuggr.xyz/thughunter/common/models"
	"smuggr.xyz/thughunter/core/datastore"
	"smuggr.xyz/thughunter/core/scraper"
)

var predefined = []string{
	`host.services.vnc.security_types.value = "1" and host.operating_system.product = "linux"`,
}

func main() {
	config.Initialize()
	datastore.Initialize(os.Getenv("DB_PATH"))
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("1. Launch Updater")
		fmt.Println("2. Browse Saved Data")
		fmt.Println("3. Exit")
		fmt.Print("Select: ")
		choiceStr, _ := r.ReadString('\n')
		choice, _ := strconv.Atoi(strings.TrimSpace(choiceStr))

		switch choice {
		case 1:
			launchUpdater(r)
		case 2:
			browseData(r)
		case 3:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func launchUpdater(r *bufio.Reader) {
	fmt.Println("Choose query or enter custom:")
	for i, q := range predefined {
		fmt.Printf("%d) %s\n", i+1, q)
	}
	fmt.Print("0) Custom query\nSelect: ")
	selStr, _ := r.ReadString('\n')
	sel, _ := strconv.Atoi(strings.TrimSpace(selStr))
	var query string
	if sel > 0 && sel <= len(predefined) {
		query = predefined[sel-1]
	} else if sel == 0 {
		fmt.Print("Enter custom query: ")
		q, _ := r.ReadString('\n')
		query = strings.TrimSpace(q)
	} else {
		fmt.Println("Invalid selection")
		return
	}

	newCount, updCount := scraper.LaunchUpdater(query)
	fmt.Printf("Import complete: %d new, %d updated\n", newCount, updCount)
}

func browseData(r *bufio.Reader) {
	fmt.Print("Filter by service (leave blank for all): ")
	f, _ := r.ReadString('\n')
	filter := strings.TrimSpace(f)

	var hosts []models.Host
	datastore.DB.Find(&hosts)

	for _, h := range hosts {
		if filter != "" {
			if _, ok := h.Services[filter]; !ok {
				continue
			}
		}
		fmt.Printf("IP: %s, Hostname: %s, Loc: %s, Labels: %v, Services: %v\n", h.IP, h.Hostname, h.Location, h.Labels, h.Services)
	}
}
