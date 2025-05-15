// main.go
package main

import (
	"bufio"
	"os"

	"github.com/joho/godotenv"
	"smuggr.xyz/thughunter/common/ui"
	"smuggr.xyz/thughunter/core/datastore"
)

var predefined = []string{
	`host.services.vnc.security_types.value = "1" and host.operating_system.product = "linux"`,
}

func main() {
	loadEnvAndInitDB()
	r := bufio.NewReader(os.Stdin)
	ui.MainMenuLoop(r, predefined)
}

func loadEnvAndInitDB() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./thughunter.db"
	}
	datastore.Initialize(dbPath)
}
