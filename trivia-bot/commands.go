package main

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	CmdChar            = "+"
	LeaderboardSpacing = 8
	NumRankingsToShow  = 10
)

// cmdFuncType Command function type
type cmdFuncType func(*discordgo.MessageCreate)

// cmdFuncHelpType The type stored in the cmdFuncs map to map a function and helper text to a command
type cmdFuncHelpType struct {
	function          cmdFuncType
	help              string
	triviaChannelOnly bool
}

// cmdFuncsType The type of the CmdFuncs map
type cmdFuncsType map[string]cmdFuncHelpType

// cmdFuncs Commands to functions map
var cmdFuncs cmdFuncsType

func init() {
	cmdFuncs = cmdFuncsType{
		"help":    cmdFuncHelpType{cmdHelp, "Prints this list", false},
		"here":    cmdFuncHelpType{cmdHere, "Sets the channel for the bot to perform trivia (per server)", false},
		"version": cmdFuncHelpType{cmdVersion, "Outputs the current bot version", true},
		"ranking": cmdFuncHelpType{cmdRanking, "Displays the current score rankings", true},
		"stats":   cmdFuncHelpType{cmdStats, "Displays stats about this bot", true},
		"start":   cmdFuncHelpType{cmdQuestion, "Triggers the bot to ask a question", true},
		"stop":    cmdFuncHelpType{cmdStop, "The current channel will no longer get trivia", true},
	}
}

// HandleCommand Called whenever a command is sent
func HandleCommand(message *discordgo.MessageCreate, cmd string) bool {
	cmdFuncHelpPair, ok := cmdFuncs[cmd]
	if ok {
		if cmdFuncHelpPair.triviaChannelOnly == false || IsTriviaChannel(message.ChannelID) {
			cmdFuncHelpPair.function(message)
			return true
		}
	} else if IsTriviaChannel(message.ChannelID) {
		var reply = fmt.Sprintf("I don't understand the command `%s`", cmd)
		DiscordSession.ChannelMessageSend(message.ChannelID, reply)
	}
	return false
}

func cmdHelp(message *discordgo.MessageCreate) {
	// Build array of the keys in CmdFuncs
	var keys []string
	for k := range cmdFuncs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build message (sorted by keys) of the commands
	var cmds = "Command notation: \n`" + CmdChar + "[command]`\n"
	cmds += "Commands:\n```\n"
	for _, key := range keys {
		cmds += fmt.Sprintf("%s - %s\n", key, cmdFuncs[key].help)
	}
	cmds += "```\n"
	cmds += "To answer a question just type it like normal chatting.\n"
	DiscordSession.ChannelMessageSend(message.ChannelID, cmds)
}

func cmdHere(message *discordgo.MessageCreate) {
	var channel, err = DiscordSession.Channel(message.ChannelID)
	if err != nil {
		fmt.Printf("Could not find channel, %s\n", err)
		return
	}

	Channels[channel.GuildID] = channel.ID
	DiscordSession.ChannelMessageSend(channel.ID, "I'm here!")
}

func cmdVersion(message *discordgo.MessageCreate) {
	DiscordSession.ChannelMessageSend(message.ChannelID, "Version: "+Version)
}

func cmdRanking(message *discordgo.MessageCreate) {
	var scoreList = make(ScoreList, len(Scores))
	var i = 0
	for k, v := range Scores {
		var user, err = DiscordSession.User(k)
		if err != nil {
			fmt.Printf("Failed to get user: %s", err)
			continue
		}
		scoreList[i] = UserScore{user.Username, v}
		i++
	}
	sort.Sort(sort.Reverse(scoreList))

	var longestNameLength = 0
	for _, score := range scoreList {
		var nameLength = len(score.name)
		if nameLength > longestNameLength {
			longestNameLength = nameLength
		}
	}

	var rankings = fmt.Sprintf("Top %d Players:\n```\n", NumRankingsToShow)
	var RankingCounter = NumRankingsToShow
	for _, score := range scoreList {
		if RankingCounter <= 0 {
			break
		}
		var numSpaces = longestNameLength + LeaderboardSpacing - len(score.name)
		rankings += score.name
		for i := 0; i < numSpaces; i++ {
			rankings += " "
		}
		rankings += fmt.Sprintf("%d\n", score.score)
		RankingCounter--
	}
	rankings += "```"
	DiscordSession.ChannelMessageSend(message.ChannelID, rankings)
}

func cmdStats(message *discordgo.MessageCreate) {
	var Stats = "Stats:\n```\n"
	Stats += fmt.Sprintf("Number of Valid Questions: %d\n", NumQuestions)
	Stats += fmt.Sprintf("Number of Invalid Questions: %d\n", NumInvalidQuestions)
	Stats += "```"
	DiscordSession.ChannelMessageSend(message.ChannelID, Stats)
}

func cmdQuestion(message *discordgo.MessageCreate) {
	SetChannelActive(message.ChannelID)
	if IsQuestionActive() {
		MessageQuestion(message.ChannelID)
	} else {
		NewQuestion()
	}
}

func cmdStop(message *discordgo.MessageCreate) {
	if IsQuestionActive() == false {
		return
	}

	SetChannelInactive(message.ChannelID)
}
