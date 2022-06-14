# This makes sure the commands are run within a BASH shell.
SHELL := /bin/bash
EXEDIR := ./bin
BIN_NAME=./bin/documents

# The .PHONY target will ignore any file that exists with the same name as the target
# in your makefile, and built it regardless.
.PHONY: all init build run clean

# The all target is the default target when make is called without any arguments.
all: clean | run

init:
	- rm go.mod
	- rm go.sum
	go mod init github.com/find-in-docs/documents
	go mod tidy -compat=1.17

${EXEDIR}:
	mkdir ${EXEDIR}

# Best way to keep track of your dependencies with your own repos are to get the modules
# from your own directory. This way, you update the source module, check it into github,
# but access it locally. To do this, issue the following commands:
#   go mod edit -replace=github.com/samirgadkari/sidecar@v0.0.0-unpublished=../sidecar
#   go get -d github.com/samirgadkari/sidecar@v0.0.0-unpublished
# This will get the repo from ../sidecar, and use it as if it is the latest version of
# github.com/samirgadkari/sidecar

build: | ${EXEDIR}
	go build -o ${BIN_NAME} pkg/main/main.go

run: build
	./${BIN_NAME}

clean:
	go clean
	go clean -cache -modcache -i -r
	- rm ${BIN_NAME}
	go mod tidy -compat=1.17

docker:
	cd ..; docker build --no-cache -t documents:v0.0.1 . -f documents/Dockerfile

