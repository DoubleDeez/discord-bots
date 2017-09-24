package main

// Global vars
var (
	Scores map[string]int
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

	return score
}
