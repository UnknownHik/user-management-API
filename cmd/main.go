package main

import (
	"log"

	"user-management/internal/pkg/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
