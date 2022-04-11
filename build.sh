#!/usr/bin/env bash

build() {
    os=$1
    output=$2
    echo -n "Building for $os..."
    GOOS="$os" GOARCH=amd64 go build -o "$output"
    echo "done: $output"
}

build linux nt5000-serial
build windows nt5000-serial.exe
build darwin nt5000-serial-macos
