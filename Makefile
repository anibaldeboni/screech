DEV_ID := github.com/anibaldeboni/screech/screenscraper.DevID=${SS_DEV_ID}
DEV_PASSWORD := github.com/anibaldeboni/screech/screenscraper.DevPassword=${SS_DEV_PASSWORD}
CFLAGS := $(shell pkg-config --cflags sdl2)
LDLAGS := $(shell pkg-config --libs SDL2_image SDL2_ttf) -ldl -lpthread -lm
DIST_DIR := dist/Screech
.PHONY: run build package lint test
.DEFAULT: package

package: clean build
	mkdir -p ${DIST_DIR}/assets && \
	cp -r ./assets/*.bmp ${DIST_DIR}/assets/ && \
	cp -r ./includes/* ${DIST_DIR}/ && \
	cp ./bin/app ${DIST_DIR}/
	zip -r ${DIST_DIR}/ScreechApp.zip ${DIST_DIR}/

build:
	go build \
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
