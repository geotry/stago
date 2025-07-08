SHELL := /bin/bash
MAKEFLAGS += -j2
PATH := $(HOME)/.local/bin:$(HOME)/go/bin:$(PATH)

gen_js:
	rm -f ./web/src/pb/* && protoc -I ./proto --js_out=import_style=commonjs:./web/src/pb ./proto/*.proto

gen_go:
	rm -f ./pb/* && protoc -I ./proto --go_out=./pb --go_opt=paths=source_relative ./proto/*.proto	

gen: gen_js gen_go

install:
	@cd web && npm i

build_web: gen_js
	@cd web && npx webpack --mode=development

build: gen_go build_web
	@go build -o .out/build .

race: gen_go build_web
	@go run -race . -port 9090

run: gen_go build_web
	@go run . -port 9090

test: gen_go
	@go test -v ./...

bench: gen_go
	@go test -bench=. ./...

watch_web:
	@cd web && npx webpack serve --mode=development

watch_go:
	@while true; do \
		rm -f ./pb/* && protoc -I ./proto --go_out=./pb --go_opt=paths=source_relative ./proto/*.proto ; \
		rm -f ./web/src/pb/* && protoc -I ./proto --js_out=import_style=commonjs:./web/src/pb ./proto/*.proto ; \
		go run . -port 9090 & \
		pid=$$!; \
		inotifywait -qr -e modify -e create -e delete -e move --exclude '/\..+' **/*.go assets/*; \
		echo "File change detected, restarting server..." ; \
		pkill -P $$pid 2> /dev/null && wait $$pid 2> /dev/null; \
	done

watch: watch_go watch_web

.PHONY: all build gen gen_js gen_go serve web run watch