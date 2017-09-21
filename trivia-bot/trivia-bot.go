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
	Version = "v0.1.0"
	CmdChar = "+"
)

// Global vars
var (
	Token    string
	Channels map[string]string
	Scores   map[string]int
)

// UserScore Struct representing a username and score
type UserScore struct {
	name  string
	score int
}

func (s ScoreList) Len() int           { return len(s) }
func (s ScoreList) Less(i, j int) bool { return s[i].score < s[j].score }
func (s ScoreList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// ScoreList List of user's score
type ScoreList []UserScore

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

	Channels = map[string]string{}
	Scores = map[string]int{}

	CmdFuncs = CmdFuncsType{
		"help":    CmdFuncHelpType{CmdHelp, "Prints this list"},
		"here":    CmdFuncHelpType{CmdHere, "Sets the channel for the bot to perform trivia (per server)"},
		"version": CmdFuncHelpType{CmdVersion, "Outputs the current bot version"},
		"score":   CmdFuncHelpType{CmdScore, "Increments the caller's score"},
		"ranking": CmdFuncHelpType{CmdRanking, "Displays the current score rankings"},
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
		var err = session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			fmt.Printf("Could not delete command message, %s\n", err)
		}
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
	var cmds = "Command notation: \n`" + CmdChar + "[command]`\n"
	cmds += "Commands:\n```\n"
	for _, key := range keys {
		cmds += fmt.Sprintf("%s - %s\n", key, CmdFuncs[key].help)
	}
	cmds += "```"
	session.ChannelMessageSend(message.ChannelID, cmds)
}

// CmdHere Command function to tell the bot where to host trivia
func CmdHere(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	var channel, err = session.Channel(message.ChannelID)
	if err != nil {
		fmt.Printf("Could not find channel, %s\n", err)
		return
	}

	Channels[channel.GuildID] = channel.ID
	session.ChannelMessageSend(channel.ID, "I'm here!")
}

// CmdVersion Outputs the current bot version
func CmdVersion(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	session.ChannelMessageSend(message.ChannelID, "Version: "+Version)
}

// CmdScore Increments the cmder's score
func CmdScore(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	if IsTriviaChannel(session, message.ChannelID) == false {
		return
	}

	var user = message.Author
	var score = 0
	if val, ok := Scores[user.ID]; ok {
		score = val + 1
	} else {
		score = 1
	}

	Scores[user.ID] = score
	var reply = fmt.Sprintf("%s now has %d point(s)!", user.Mention(), score)
	session.ChannelMessageSend(message.ChannelID, reply)
}

// CmdRanking Displays the current score rankings
func CmdRanking(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	if IsTriviaChannel(session, message.ChannelID) == false {
		return
	}

	var scoreList = make(ScoreList, len(Scores))
	var i = 0
	for k, v := range Scores {
		var user, err = session.User(k)
		if err != nil {
			fmt.Printf("Failed to get user: %s", err)
			continue
		}
		scoreList[i] = UserScore{user.Username, v}
		i++
	}
	sort.Sort(sort.Reverse(scoreList))

	var rankings = "The Current Rankings:\n```\n"
	for idx := range scoreList {
		rankings += fmt.Sprintf("%s \t\t\t %d\n", scoreList[idx].name, scoreList[idx].score)
	}
	rankings += "```"
	session.ChannelMessageSend(message.ChannelID, rankings)
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
