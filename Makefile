# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
BIN_NAME=./bin/db

LATESTVER := "$(shell go list -m -u github.com/samirgadkari/sidecar | rg -o 'v[^\]]*')"

# The .PHONY target will ignore any file that exists with the same name as the target
# in your makefile, and built it regardless.
.PHONY: all build run clean

# The all target is the default target when make is called without any arguments.
all: run

${EXEDIR}:
	mkdir ${EXEDIR}

build: | ${EXEDIR}
	go get github.com/samirgadkari/sidecar@$(LATESTVER)
	go build -o ${BIN_NAME} cli/main.go

run: build
	./${BIN_NAME} serve

clean:
	go clean
	rm ${BIN_NAME}
	go clean -cache -modcache -i -r
	go get github.com/samirgadkari/sidecar@$(LATESTVER)
	go mod tidy
