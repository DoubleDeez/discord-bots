@echo off
go clean
go build
go install
js-bot -t "[TOKEN]"