package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	Version = "v0.2.0"
)

// Global vars
var (
	Token    string
	Channels map[string]string
)

func init() {
	fmt.Println("Startin Trivia Bot...")
	flag.StringVar(&Token, "t", "", "Discord Authentication Token")
	flag.Parse()

	Channels = map[string]string{}
	InitQuestions()
	InitScoreTracking()
	InitCmds()
}

func main() {
	if Token == "" {
		fmt.Println("You must provide a Discord authentication token (-t)")
		return
	}

	var DiscordSession, err = discordgo.New("Bot " + Token)
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

	if strings.HasPrefix(message.Content, CmdChar) {
		var HandledCommand = HandleCommand(session, message, strings.TrimPrefix(message.Content, CmdChar), message.Author.ID)
		if HandledCommand {
			var err = session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("Could not delete command message, %s\n", err)
			}
		}
	} else if IsTriviaChannel(session, message.ChannelID) && IsQuestionActive() {
		CheckAnswer(session, message)
	}
}

// IsTriviaChannel Returns true if the channelID is the server's select trivia channel
func IsTriviaChannel(session *discordgo.Session, channelID string) bool {
	var channel, err = session.Channel(channelID)
	if err != nil {
		fmt.Printf("Could not find channel, %s\n", err)
		return false
	}

	return Channels[channel.GuildID] == channelID
}
