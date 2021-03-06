#!/bin/bash
## execute unit test

go get -v -t -d ./...
go test -v -coverprofile=c.out ./pkg/...
go tool cover -html=c.out -o coverage.html