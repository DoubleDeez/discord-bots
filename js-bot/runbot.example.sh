#!/bin/bash
go clean
go build
go install
js-bot -t "[TOKEN]"