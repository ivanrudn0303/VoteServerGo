GOPATH=$(shell pwd)
export GOPATH

all: prepare build

build:
	go install Webserver/

prepare:
	echo '{"address_listen": ":8080", "address_blockchain": "127.0.0.1:8888", "ip_sql": "127.0.0.1", "port_sql": "5432", "user_sql": "ivan", "password_sql": "pass"}' > ~/.vote_server
	go get -u github.com/lib/pq
