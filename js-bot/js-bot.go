package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/robertkrimen/otto"
)

// Constants
const (
	Version = "v0.0.1"
)

// Global vars
var (
	Token          string
	DiscordSession *discordgo.Session
)

func init() {
	fmt.Println("Starting JS Bot...")
	flag.StringVar(&Token, "t", "", "Discord Authentication Token")
	flag.Parse()
}

func main() {
	if Token == "" {
		fmt.Println("You must provide a Discord authentication token (-t)")
		return
	}

	var err error
	DiscordSession, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Printf("Error creating Discord session, %s\n", err)
		return
	}

	err = DiscordSession.Open()
	if err != nil {
		fmt.Printf("Error opening connection to Discord, %s\n", err)
		return
	}

	// Setup handlers
	DiscordSession.AddHandler(OnMessageCreate)

	// Wait for a CTRL-C
	fmt.Println("Now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Clean up
	DiscordSession.Close()
}

// OnMessageCreate Called whenever a message is sent to the bot
func OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		// Ignore self
		return
	}

	if !strings.HasPrefix(message.Content, "```js") || !strings.HasSuffix(message.Content, "```") {
		return
	}

	var messageLength = len(message.Content)
	var code = message.Content[5 : messageLength-3]
	var JSInterp = otto.New()

	JSInterp.Set("print", func(call otto.FunctionCall) otto.Value {
		logString := call.Argument(0).String()
		DiscordSession.ChannelMessageSend(message.ChannelID, logString)
		return otto.Value{}
	})

	_, err := JSInterp.Run(code)
	if err != nil {
		DiscordSession.ChannelMessageSend(message.ChannelID, "Error: "+err.Error())
	}
}
