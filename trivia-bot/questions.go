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
	TimeToAnswer      = 180 // TODO - command based
)

// Global vars
var (
	NumQuestions         int
	NumInvalidQuestions  int
	CurrentQuestionIndex int
	Questions            []Question
	QuestionTimer        *time.Timer
	NumWrongAnswers      int
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
func MessageQuestion(session *discordgo.Session, channelID string) {
	var currQuestion = Questions[CurrentQuestionIndex]
	var QuestionMessage = fmt.Sprintf("The question is: \n%s", currQuestion.question)
	session.ChannelMessageSend(channelID, QuestionMessage)
}

// NewQuestion Starts a new question
func NewQuestion(session *discordgo.Session, channelID string) {
	if IsQuestionActive() == false {
		NumWrongAnswers = 0
		setRandomQuestion()
		var NewQuestionMessage = fmt.Sprintf("A new question, you have %d seconds to answer it!", TimeToAnswer)
		session.ChannelMessageSend(channelID, NewQuestionMessage)
		QuestionTimer = time.NewTimer(time.Second * TimeToAnswer)
		go func(session *discordgo.Session, channelID string) {
			<-QuestionTimer.C
			onTimeRanOut(session, channelID)
		}(session, channelID)
	}

	MessageQuestion(session, channelID)
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
		NumWrongAnswers++
	}
}

func correctAnswer(session *discordgo.Session, message *discordgo.MessageCreate) {
	CurrentQuestionIndex = -1
	QuestionTimer.Stop()
	var user = message.Author
	var answerMessage = fmt.Sprintf("%s had the correct answer with `%s`. There were %d wrong answer(s)", user.Mention(), message.Content, NumWrongAnswers)
	session.ChannelMessageSend(message.ChannelID, answerMessage)
	var score = IncrementUserScore(user.ID)
	var scoreMessage = fmt.Sprintf("%s now has %d point(s)!", user.Mention(), score)
	session.ChannelMessageSend(message.ChannelID, scoreMessage)

	NewQuestion(session, message.ChannelID)
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

func onTimeRanOut(session *discordgo.Session, channelID string) {
	if IsQuestionActive() == false {
		return
	}

	var currQuestion = Questions[CurrentQuestionIndex]
	CurrentQuestionIndex = -1

	var answerMessage = fmt.Sprintf("Time ran out!\nThe correct answer was `%s` There were %d wrong answer(s)", currQuestion.answers[0], NumWrongAnswers)
	session.ChannelMessageSend(channelID, answerMessage)

	if NumWrongAnswers > 0 {
		NewQuestion(session, channelID)
	}
}
