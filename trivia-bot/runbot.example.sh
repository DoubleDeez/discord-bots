#!/bin/bash
go clean
go build
go install
trivia-bot -t "[TOKEN]"