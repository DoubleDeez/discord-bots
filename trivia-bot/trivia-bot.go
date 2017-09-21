package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	Version = "v0.0.0"
	CmdChar = "+"
)

// Global vars
var (
	Token   string
	Channel string
)

// CmdFuncType Command function type
type CmdFuncType func(*discordgo.Session, *discordgo.MessageCreate, string)

// CmdFuncHelpType The type stored in the CmdFuncs map to map a function and helper text to a command
type CmdFuncHelpType struct {
	function CmdFuncType
	help     string
}

// CmdFuncsType The type of the CmdFuncs map
type CmdFuncsType map[string]CmdFuncHelpType

// CmdFuncs Commands to functions map
var CmdFuncs CmdFuncsType

func init() {
	fmt.Println("Startin Trivia Bot...")
	flag.StringVar(&Token, "t", "", "Discord Authentication Token")
	flag.Parse()

	CmdFuncs = CmdFuncsType{
		"help": CmdFuncHelpType{CmdHelp, "Prints this list"},
		"here": CmdFuncHelpType{CmdHere, "Sets the channel for the bot to perform trivia"},
	}
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
		HandleCommand(session, message, strings.TrimPrefix(message.Content, CmdChar), message.Author.ID)
	}
}

// HandleCommand Called whenever a command is sent
func HandleCommand(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, cmder string) {
	CmdFuncHelpPair, ok := CmdFuncs[cmd]
	if ok {
		CmdFuncHelpPair.function(session, message, cmder)
	} else {
		var reply = fmt.Sprintf("I don't understand the command `%s`", cmd)
		session.ChannelMessageSend(message.ChannelID, reply)
	}
}

// CmdHelp Lists all available cmds
func CmdHelp(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	// Build array of the keys in CmdFuncs
	var keys []string
	for k := range CmdFuncs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build message (sorted by keys) of the commands
	var cmds = "Commands:\n```\n"
	for _, key := range keys {
		cmds += fmt.Sprintf("%s - %s\n", key, CmdFuncs[key].help)
	}
	cmds += "```"
	session.ChannelMessageSend(message.ChannelID, cmds)
}

// CmdHere Command function to tell the bot where to host trivia
func CmdHere(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	Channel = message.ChannelID
	session.ChannelMessageSend(Channel, "I'm here!")
}
