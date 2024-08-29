DEV_ID := github.com/anibaldeboni/screech/screenscraper.DevID=${SS_DEV_ID}
DEV_PASSWORD := github.com/anibaldeboni/screech/screenscraper.DevPassword=${SS_DEV_PASSWORD}
CFLAGS := $(shell pkg-config --cflags sdl2)
LDLAGS := $(shell pkg-config --libs SDL2_image SDL2_ttf) -ldl -lpthread -lm

.PHONY: run build package
.DEFAULT: package

package: build
	rm -rf Screech && \
	mkdir -p Screech/assets/logos && \
	cp -r ./assets/*.bmp Screech/assets/ && \
	cp -r ./includes/* Screech/ && \
	cp -r ./assets/logos/* Screech/assets/logos/ && \
	cp ./bin/app Screech/

build:
	CGO_ENABLED=1 \
	CGO_CFLAGS="${CFLAGS}" \
	CGO_LDFLAGS="${LDLAGS}" \
	GOARCH=arm64 \
	CC=aarch64-linux-gnu-gcc \
	go build \
	-tags static \
	-buildvcs=false \
	-ldflags "-X ${DEV_ID} -X ${DEV_PASSWORD}" \
	-o bin/app ./

run:
	go run -ldflags "-X ${DEV_ID} -X ${DEV_PASSWORD}" main.go
