package main

import (
	"log"
	"time"

	"github.com/geniusdex/broawp/accbroadcast"
	"github.com/geniusdex/broawp/accrace"
	"github.com/geniusdex/broawp/frontend"
)

func main() {
	log.Printf("Starting BROAWP")

	client, err := accbroadcast.NewClient("localhost", 9000)
	if err != nil {
		log.Fatalf("Cannot create UDP client: %v", err)
	}

	defer client.Close()

	if err := client.Register("broawp", "asd", 100*time.Millisecond, ""); err != nil {
		log.Fatalf("cannot register: %v", err)
	}

	state := accrace.NewState(client)
	defer state.Close()

	if err := frontend.Run(state); err != nil {
		log.Printf("Error running frontend: %v", err)
	}
}
