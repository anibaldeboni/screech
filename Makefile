MODULE := github.com/anibaldeboni/screech
DEV_ID := ${MODULE}/scraper.DevID=${SS_DEV_ID}
DEV_PASSWORD := ${MODULE}/scraper.DevPassword=${SS_DEV_PASSWORD}
CFLAGS := $(shell pkg-config --cflags sdl2)
LDLAGS := $(shell pkg-config --libs SDL2_image SDL2_ttf) -ldl -lpthread -lm
DIST_DIR := dist
.PHONY: run build package lint test
.DEFAULT: package

package: clean build
	@mkdir -p ${DIST_DIR}
	zip -j -r ${DIST_DIR}/ScreechApp.zip assets/*.bmp includes/* bin/app

build:
	@go build \
	-v \
	-tags static \
	-buildvcs=false \
	-ldflags "-s -w -X ${DEV_ID} -X ${DEV_PASSWORD}" \
	-o bin/app ./

clean:
	rm -rf bin/* Screech

run:
	@go run -ldflags "-X ${DEV_ID} -X ${DEV_PASSWORD}" main.go

lint: ##@dev Run lint (download from https://golangci-lint.run/usage/install/#local-installation)
	golangci-lint run -v

test:
	go test -cover -v ./...
