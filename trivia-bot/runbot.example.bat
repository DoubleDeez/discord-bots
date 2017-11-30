@echo off
go clean
go build
go install
trivia-bot -t "[TOKEN]"