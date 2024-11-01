MODULE := github.com/anibaldeboni/screech
DEV_ID := ${MODULE}/scraper.DevID=${SS_DEV_ID}
DEV_PASSWORD := ${MODULE}/scraper.DevPassword=${SS_DEV_PASSWORD}
# VERSION := ""
DIST_DIR := ScreechApp
BIN_DIR := bin

# Detect the operating system
UNAME_S := $(shell uname -s)

# Default values for CFLAGS and LDFLAGS
CFLAGS := ""
LDFLAGS := ""
CC := ""

ifndef APP_VERSION
VERSION=${MODULE}/config.Version=$(shell git tag --sort=-version:refname | head -n 1)
else
VERSION=${MODULE}/config.Version=${APP_VERSION}
endif

# Set flags based on the operating system
ifeq ($(UNAME_S), Darwin)
    # MacOS specific flags
    CFLAGS = $(shell pkg-config --cflags sdl2)
    LDFLAGS = $(shell pkg-config --libs SDL2_image SDL2_ttf) -ldl -lpthread -lm
		CC = "aarch64-linux-gnu-gcc"
else ifeq ($(UNAME_S), Linux)
    # Linux specific flags
    CFLAGS = -I${SYSROOT}/usr/include -I/usr/aarch64-linux-gnu/include -I/usr/aarch64-linux-gnu/include/SDL2 -I/usr/include/SDL2 -D_REENTRANT
    LDFLAGS = -L${SYSROOT}/usr/lib -L/usr/lib/aarch64-linux-gnu -lSDL2_image -lSDL2_ttf -lSDL2 -ldl -lpthread -lm
		CC = "aarch64-linux-gnu-gcc --sysroot=${SYSROOT}"
endif

.PHONY: run build package lint test
.DEFAULT: package

package: clean build
	@mkdir -p ${DIST_DIR}/assets && \
	cp -r assets/*.bmp ${DIST_DIR}/assets && \
	cp -r includes/* ${DIST_DIR} && \
	cp -r bin/app ${DIST_DIR} && \
	chmod +x ${DIST_DIR}/app && \
	zip -g -r ${DIST_DIR}/ScreechApp.zip ${DIST_DIR}

build:
	@CGO_CFLAGS="${CFLAGS}" \
	CGO_LDFLAGS="${LDFLAGS}" \
	CC=${CC} \
	GOARCH=arm64 \
	GOOS=linux \
	go build \
	-tags static \
	-ldflags "-s -w -X ${DEV_ID} -X ${DEV_PASSWORD} -X ${VERSION}" \
	-o bin/app ./

clean:
	@rm -rf ${BIN_DIR}/* ${DIST_DIR}/*

run:
	@go run -ldflags "-X ${DEV_ID} -X ${DEV_PASSWORD} -X ${VERSION}" main.go --config=./includes/screech.yaml

lint: ##@dev Run lint (download from https://golangci-lint.run/usage/install/#local-installation)
	@golangci-lint run -v

test:
	go test -cover -v ./...
