package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/natsukagami/amabot"
)

// The bot token
var token string

func init() {
	token = os.Getenv("TOKEN")
	if token == "" {
		// No token was found
		panic("Missing TOKEN environment variable.")
	}
}

func main() {
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.Open()
	fmt.Println("Bot online")

	stopListening := dg.AddHandler(amabot.Handle)

	amabot.Loop(dg, stopListening)
}
