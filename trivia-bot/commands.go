package main

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	CmdChar = "+"
)

// CmdFuncType Command function type
type CmdFuncType func(*discordgo.Session, *discordgo.MessageCreate, string)

// CmdFuncHelpType The type stored in the CmdFuncs map to map a function and helper text to a command
type CmdFuncHelpType struct {
	function          CmdFuncType
	help              string
	triviaChannelOnly bool
}

// CmdFuncsType The type of the CmdFuncs map
type CmdFuncsType map[string]CmdFuncHelpType

// CmdFuncs Commands to functions map
var CmdFuncs CmdFuncsType

// InitCmds Initializes the cmds map
func InitCmds() {
	CmdFuncs = CmdFuncsType{
		"help":     CmdFuncHelpType{cmdHelp, "Prints this list", false},
		"here":     CmdFuncHelpType{cmdHere, "Sets the channel for the bot to perform trivia (per server)", false},
		"version":  CmdFuncHelpType{cmdVersion, "Outputs the current bot version", true},
		"ranking":  CmdFuncHelpType{cmdRanking, "Displays the current score rankings", true},
		"stats":    CmdFuncHelpType{cmdStats, "Displays stats about this bot", true},
		"question": CmdFuncHelpType{cmdQuestion, "Triggers the bot to ask a question", true},
	}
}

// HandleCommand Called whenever a command is sent
func HandleCommand(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, cmder string) bool {
	CmdFuncHelpPair, ok := CmdFuncs[cmd]
	if ok {
		if CmdFuncHelpPair.triviaChannelOnly == false || IsTriviaChannel(session, message.ChannelID) {
			CmdFuncHelpPair.function(session, message, cmder)
			return true
		}
	} else if IsTriviaChannel(session, message.ChannelID) {
		var reply = fmt.Sprintf("I don't understand the command `%s`", cmd)
		session.ChannelMessageSend(message.ChannelID, reply)
	}
	return false
}

func cmdHelp(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
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
	cmds += "```\n"
	cmds += "To answer a question just type it like normal chatting.\n"
	session.ChannelMessageSend(message.ChannelID, cmds)
}

func cmdHere(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	var channel, err = session.Channel(message.ChannelID)
	if err != nil {
		fmt.Printf("Could not find channel, %s\n", err)
		return
	}

	Channels[channel.GuildID] = channel.ID
	session.ChannelMessageSend(channel.ID, "I'm here!")
}

func cmdVersion(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	session.ChannelMessageSend(message.ChannelID, "Version: "+Version)
}

func cmdRanking(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
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

func cmdStats(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	var Stats = "Stats:\n```\n"
	Stats += fmt.Sprintf("Number of Valid Questions: %d\n", NumQuestions)
	Stats += fmt.Sprintf("Number of Invalid Questions: %d\n", NumInvalidQuestions)
	Stats += "```"
	session.ChannelMessageSend(message.ChannelID, Stats)
}

func cmdQuestion(session *discordgo.Session, message *discordgo.MessageCreate, cmder string) {
	NewQuestion(session, message.ChannelID)
}
