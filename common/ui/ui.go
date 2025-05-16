package ui

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"smuggr.xyz/thughunter/common/models"
	"smuggr.xyz/thughunter/core/datastore"
	"smuggr.xyz/thughunter/core/scanner"
	"smuggr.xyz/thughunter/core/scraper"
)

func MainMenuLoop(r *bufio.Reader, predefined []string) {
	for {
		switch showMainMenu(r) {
		case 1:
			launchUpdater(r, predefined)
		case 2:
			browseData(r)
		case 3:
			scanner.RunScan(r)
		case 4:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func showMainMenu(r *bufio.Reader) int {
	fmt.Println("1. Launch Updater")
	fmt.Println("2. Browse Saved Data")
	fmt.Println("3. Check VNC Services and Snapshot")
	fmt.Println("4. Exit")
	fmt.Print("Select: ")
	choiceStr, _ := r.ReadString('\n')
	choice, _ := strconv.Atoi(strings.TrimSpace(choiceStr))
	return choice
}

func launchUpdater(r *bufio.Reader, predefined []string) {
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

	if filter == "" {
		for _, h := range hosts {
			fmt.Printf("IP: %s, Hostname: %s, Loc: %s, Labels: %v, Services: %v\n",
				h.IP, h.Hostname, h.Location, h.Labels, h.Services)
		}

		fmt.Printf("Loaded %d hosts from DB\n", len(hosts))
		fmt.Println("No filter applied, showing all hosts.")
		return
	}

	lowerFilter := strings.ToLower(filter)
	matchFound := false
	fmt.Printf("Hosts with '%s' service:\n", filter)

	for _, h := range hosts {
		if h.Services == nil {
			continue
		}
		for svc, port := range h.Services {
			if strings.Contains(strings.ToLower(svc), lowerFilter) {
				fmt.Printf("%s:%d (%s)\n", h.IP, port, svc)
				matchFound = true
			}
		}
	}

	fmt.Printf("Loaded %d hosts from DB\n", len(hosts))

	if !matchFound {
		fmt.Printf("No hosts found with service matching '%s'\n", filter)
	}
}
