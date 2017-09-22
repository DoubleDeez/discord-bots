package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// Global constants
const (
	QuestionsFilePath = "trivia.txt"
)

// Global vars
var (
	NumQuestions         int
	NumInvalidQuestions  int
	CurrentQuestionIndex int
	Questions            []Question
)

// Question holds the question and answers
type Question struct {
	question string
	answers  []string
}

// InitQuestions Initializes the questions
func InitQuestions() {
	CurrentQuestionIndex = -1
	var QuestionFileData, err = ioutil.ReadFile(QuestionsFilePath)
	if err != nil {
		fmt.Printf("Failed to open questions file (%s): %s\n", QuestionsFilePath, err)
		return
	}

	var QuestionDataString = string(QuestionFileData[:])
	var QuestionLines = strings.Split(QuestionDataString, "\n")
	NumQuestions = len(QuestionLines)
	Questions = make([]Question, NumQuestions)

	NumInvalidQuestions = 0
	var ValidQuestionIndex = 0
	for _, QuestionLine := range QuestionLines {
		var QuestionComponents = strings.Split(QuestionLine, "`")
		if len(QuestionComponents) < 2 {
			NumInvalidQuestions++
			fmt.Printf("Invalid question (%s)\n", QuestionLine)
			continue
		}
		var NextQuestion Question
		NextQuestion.question = QuestionComponents[0]
		NextQuestion.answers = QuestionComponents[1:]
		Questions[ValidQuestionIndex] = NextQuestion
		ValidQuestionIndex++
	}
	NumQuestions = ValidQuestionIndex
	fmt.Printf("There are %d invalid questions\n", NumInvalidQuestions)
}

// IsQuestionActive returns true if a question is active
func IsQuestionActive() bool {
	return CurrentQuestionIndex > 0
}

// MessageQuestion messages the active question if there is one
func MessageQuestion(session *discordgo.Session, message *discordgo.MessageCreate) {
	var currQuestion = Questions[CurrentQuestionIndex]
	var QuestionMessage = fmt.Sprintf("The question is: \n`%s`", currQuestion.question)
	session.ChannelMessageSend(message.ChannelID, QuestionMessage)
}

// NewQuestion Starts a new question
func NewQuestion(session *discordgo.Session, message *discordgo.MessageCreate) {
	if IsQuestionActive() == false {
		setRandomQuestion()
		session.ChannelMessageSend(message.ChannelID, "A new question!")
	}

	MessageQuestion(session, message)
}

// CheckAnswer Checks if the input is the correct answer
func CheckAnswer(session *discordgo.Session, message *discordgo.MessageCreate) {
	if IsQuestionActive() {
		var currQuestion = Questions[CurrentQuestionIndex]
		var Input = strings.ToLower(stripWhiteSpace(message.Content))
		for _, ans := range currQuestion.answers {
			var Answer = strings.ToLower(stripWhiteSpace(ans))
			if Answer == Input {
				correctAnswer(session, message)
				return
			}
		}
	}
}

func correctAnswer(session *discordgo.Session, message *discordgo.MessageCreate) {
	CurrentQuestionIndex = -1

	var user = message.Author
	var answerMessage = fmt.Sprintf("%s had the correct answer with `%s`", user.Mention(), message.Content)
	session.ChannelMessageSend(message.ChannelID, answerMessage)
	var score = IncrementUserScore(user.ID)
	var scoreMessage = fmt.Sprintf("%s now has %d point(s)!", user.Mention(), score)
	session.ChannelMessageSend(message.ChannelID, scoreMessage)
}

func setRandomQuestion() {
	rand.Seed(time.Now().Unix())
	CurrentQuestionIndex = rand.Intn(NumQuestions - 1)
}

func stripWhiteSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
