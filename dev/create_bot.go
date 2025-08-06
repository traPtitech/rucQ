package main

import (
	"fmt"
	"log"

	"github.com/traPtitech/rucQ/testutil/bot"
)

func main() {
	const traqURL = "http://localhost:3000/api/v3"

	accessToken, err := bot.CreateBot(traqURL)

	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	fmt.Println("TRAQ_BOT_ACCESS_TOKEN=" + accessToken)
}
