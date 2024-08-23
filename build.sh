#!/bin/bash
export CGO_ENABLED=1
export CGO_CFLAGS="$(pkg-config --cflags SDL2_ttf)"
export CGO_LDFLAGS="$(pkg-config --libs SDL2_ttf) -ldl -lpthread -lm"
export GOARCH="arm64"
export CC="aarch64-linux-gnu-gcc"

# -ldflags "-s -w"
go build -tags static -o ./DemoApp/app ./ \
&& cp -r ./includes/* DemoApp/
