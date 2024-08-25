#!/bin/bash
export CGO_ENABLED=1
export CGO_CFLAGS="$(pkg-config --cflags sdl2)"
export CGO_LDFLAGS="$(pkg-config --libs SDL2_image) $(pkg-config --libs SDL2_ttf) -ldl -lpthread -lm"
export GOARCH="arm64"
export CC="aarch64-linux-gnu-gcc"

# -ldflags "-s -w"
rm -rf Screech \
&& go build -tags static -buildvcs=false -o ./Screech/app ./ \
&& mkdir -p Screech/assets \
&& cp -r ./assets/*.bmp Screech/assets/ \
&& cp -r ./includes/* Screech/ \
