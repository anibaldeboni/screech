MODULE := github.com/anibaldeboni/screech
DEV_ID := ${MODULE}/scraper.DevID=${SS_DEV_ID}
DEV_PASSWORD := ${MODULE}/scraper.DevPassword=${SS_DEV_PASSWORD}
CFLAGS := $(shell pkg-config --cflags sdl2)
LDLAGS := $(shell pkg-config --libs SDL2_image SDL2_ttf) -ldl -lpthread -lm
DIST_DIR := ScreechApp
BIN_DIR := bin
.PHONY: run build package lint test build-macos
.DEFAULT: package

package: clean build
	@mkdir -p ${DIST_DIR}/assets && \
	cp -r assets/*.bmp ${DIST_DIR}/assets && \
	cp -r includes/* ${DIST_DIR} && \
	cp -r bin/app ${DIST_DIR} && \
	chmod +x ${DIST_DIR}/app && \
	zip -g -r ${DIST_DIR}/ScreechApp.zip ${DIST_DIR}

build:
	@go build \
	-tags static \
	-buildvcs=false \
	-ldflags "-s -w -X ${DEV_ID} -X ${DEV_PASSWORD}" \
	-o bin/app ./

build-macos:
	CGO_CFLAGS="${CFLAGS}" \
	CGO_LDFLAGS="${LDLAGS}" \
	GOARCH=arm64 \
	GOOS=linux \
	CC="aarch64-linux-gnu-gcc" \
	go build \
	-tags static \
	-buildvcs=false \
	-ldflags "-s -w -X ${DEV_ID} -X ${DEV_PASSWORD}" \
	-o bin/app ./

clean:
	@rm -rf ${BIN_DIR}/* ${DIST_DIR}/*

run:
	@go run -ldflags "-X ${DEV_ID} -X ${DEV_PASSWORD}" main.go

lint: ##@dev Run lint (download from https://golangci-lint.run/usage/install/#local-installation)
	@golangci-lint run -v

test:
	go test -cover -v ./...
