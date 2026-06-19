// main.go
package main

import (
	"log"
	"os"
	"time"
)

const passwordPath = "/etc/secrets/DB_PASSWORD"

func main() {
	for {
		b, err := os.ReadFile(passwordPath)
		if err != nil {
			log.Printf("error reading password: %v", err)
		} else {
			log.Printf("DB_PASSWORD=%s", string(b))
		}

		time.Sleep(5 * time.Second)
	}
}
