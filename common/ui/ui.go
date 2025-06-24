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
	fmt.Println("\n=== ThugHunter Main Menu ===")
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
	fmt.Printf("%d) Run all predefined queries automatically\n", len(predefined)+1)
	fmt.Print("0) Custom query\nSelect: ")
	selStr, _ := r.ReadString('\n')
	sel, _ := strconv.Atoi(strings.TrimSpace(selStr))

	if sel == len(predefined)+1 {
		runAllQueries(predefined)
		return
	}

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

func runAllQueries(predefined []string) {
	fmt.Printf("Running all %d predefined queries automatically...\n", len(predefined))
	totalNew := 0
	totalUpdated := 0

	for i, query := range predefined {
		fmt.Printf("\n[%d/%d] Running query: %s\n", i+1, len(predefined), query)
		newCount, updCount := scraper.LaunchUpdater(query)
		fmt.Printf("Query %d complete: %d new, %d updated\n", i+1, newCount, updCount)
		totalNew += newCount
		totalUpdated += updCount
	}

	fmt.Printf("\n=== All queries completed ===\n")
	fmt.Printf("Total results: %d new, %d updated\n", totalNew, totalUpdated)
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
