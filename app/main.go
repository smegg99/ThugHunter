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
}

func main() {
	loadEnvAndInitDB()
	r := bufio.NewReader(os.Stdin)
	scanner.StartControlServer()
	ui.MainMenuLoop(r, predefined)
}

func loadEnvAndInitDB() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file, using defaults:", err)
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./thughunter.db"
	}
	datastore.Initialize(dbPath)
}
