// common/config/config.go
package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}

func Initialize() error {
	fmt.Println("initializing env")

	if err := loadEnv(); err != nil {
		return err
	}

	fmt.Println("env loaded")

	return nil
}