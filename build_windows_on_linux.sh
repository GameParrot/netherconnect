#!/bin/bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc fyne package --release --name NetherConnect --appID com.gameparrot.netherconnect
