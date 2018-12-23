#!/bin/sh -x

if [ ! -d ./mac ] ; then
  mkdir mac
fi

if [ ! -d ./linux_amd64 ] ; then
  mkdir linux_amd64
fi

if [ ! -d ./windows_386 ] ; then
  mkdir windows_386
fi

## Mac用コマンドをコンパイル
go build -o mac/c2ptcli c2ptcli.go

## Linux 64bit用コマンドをコンパイル
GOOS=linux GOARCH=amd64 go build -o linux_amd64/c2ptcli c2ptcli.go

## Windows 32bit用コマンドをコンパイル
GOOS=windows GOARCH=386 go build -o windows_386/c2ptcli.exe c2ptcli.go
