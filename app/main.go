// main.go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"smuggr.xyz/thughunter/common/ui"
	"smuggr.xyz/thughunter/core/datastore"
	"smuggr.xyz/thughunter/core/scanner"
)

var predefined = []string{
	`host.services.vnc.security_types.value = "1" and host.operating_system.product = "linux"`,
	`host.services.vnc.security_types.value = "1" and host.services.vnc.desktop_name= "QEMU"`,
	`host.services.vnc.security_types.value = "1" and host.services.vnc.security_types.name = "None"`,
	`host.services.vnc.security_types.value = "1" and (host.operating_system.product = "linux" or host.operating_system.product = "unix")`,
	`host.operating_system.product = "linux"`,
	`host.operating_system.product = "unix"`,
	`host.services.vnc.desktop_name= "QEMU`,
	`host.services.vnc.security_types.name = "None"`,
}

func main() {
	fmt.Println("Starting ThugHunter...")
	loadEnvAndInitDB()
	r := bufio.NewReader(os.Stdin)
	scanner.StartControlServer()
	ui.MainMenuLoop(r, predefined)
}

func loadEnvAndInitDB() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using defaults")
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./thughunter.db"
	}
	fmt.Printf("Initializing database: %s\n", dbPath)
	datastore.Initialize(dbPath)
}
