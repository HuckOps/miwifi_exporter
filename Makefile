PROJECT="miwifi_exporter"

MAIN_PATH="main.go"
VERSION="v0.0.1"
DATE= `date +%FT%T%z`

ifeq (${VERSION}, "v0.0.1")
		VERSION=VERSION = "v0.0.1"
endif

version:
		@echo ${VERSION}

.PHONY: build
build:
		@echo version: ${VERSION} date: ${DATE} os: Mac OS
		@go mod tidy
		@rm -rf ./bin
		@mkdir bin
		@go  build -o ./bin/${PROJECT} ${MAIN_PATH}

install:
		@echo download package
		@go mod download

build-linux:
		@echo version: ${VERSION} date: ${DATE} os: linux-centOS
		@GOOS=linux go build -o ${PROJECT} ${MAIN_PATH}

run:   build
		@./bin/${PROJECT}

clean:
		rm -rf ./log
