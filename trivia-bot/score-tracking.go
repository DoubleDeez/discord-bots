package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Global vars
var (
	Scores map[string]int
)

// Constants
const (
	ScoreSaveFile = "scores.sav"
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

func init() {
	Scores = map[string]int{}
	loadScores()
}

// IncrementUserScore Increments a user's score, returns user's score
func IncrementUserScore(UserID string) int {
	var score = 0
	if val, ok := Scores[UserID]; ok {
		score = val + 1
	} else {
		score = 1
	}

	Scores[UserID] = score
	saveScores()
	return score
}

func saveScores() {
	var file, err = os.Create(ScoreSaveFile)
	if err != nil {
		fmt.Printf("Error making save file: %s\n", err)
		return
	}

	defer file.Close()

	for userID, score := range Scores {
		file.WriteString(userID + "`")
		file.WriteString(strconv.Itoa(score) + "\n")
	}
}

func loadScores() {
	if _, err := os.Stat(ScoreSaveFile); os.IsNotExist(err) {
		return
	}

	var ScoreFile, err = ioutil.ReadFile(ScoreSaveFile)
	if err != nil {
		fmt.Printf("Failed to open score save file (%s): %s\n", QuestionsFilePath, err)
		return
	}

	var ScoresString = string(ScoreFile[:])
	var ScoresLines = strings.Split(ScoresString, "\n")

	for _, ScoreLine := range ScoresLines {
		if ScoreLine == "" {
			continue
		}
		var ScoreIDValue = strings.Split(ScoreLine, "`")
		Scores[ScoreIDValue[0]], err = strconv.Atoi(ScoreIDValue[1])
		if err != nil {
			fmt.Printf("Failed to load score: %s\n", err)
		}
	}
}
