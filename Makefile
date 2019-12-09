GOPATH=$(shell pwd)
export GOPATH

all: prepare build

build:
	go install Webserver/

prepare:
	go get -u github.com/lib/pq
