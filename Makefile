BINARY = memcached-dump
GOARCH = amd64

RELEASE ?= $(shell grep -Eo '([0-9].[0-9].[0-9])' internal/version/version.go)
BUILD_TIME?=$(shell date '+%Y-%m-%d %H:%M:%S')

CURRENT_DIR = $(shell pwd)
RELEASE_DIR = ${CURRENT_DIR}/build
GO_VERSION = $(shell go version | awk '{print $$3,$$4}')

LDFLAGS = -ldflags "-X main.Release=${RELEASE} -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GoVersion=${GO_VERSION}'"

CMD_DUMP = ${CURRENT_DIR}

default:
	go run ${CMD_DUMP}

dev:
	go build -ldflags="-s -w" -o ${BINARY} ${CMD_DUMP}

release:
	go clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=${GOARCH} go build -ldflags="-s -w" ${LDFLAGS} -o ${RELEASE_DIR}/${BINARY}_darwin ${CMD_DUMP}
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -ldflags="-s -w" ${LDFLAGS} -o ${RELEASE_DIR}/${BINARY} ${CMD_DUMP}
	CGO_ENABLED=0 GOOS=windows GOARCH=${GOARCH} go build -ldflags="-s -w" ${LDFLAGS} -o ${RELEASE_DIR}/${BINARY}.exe ${CMD_DUMP}