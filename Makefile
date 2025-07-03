SHELL := /bin/bash

PATH := $(HOME)/.local/bin:$(HOME)/go/bin:$(PATH)

gen_js:
	rm -f ./web/pb/* && protoc -I ./proto --js_out=import_style=commonjs:./web/pb ./proto/*.proto

gen_go:
	rm -f ./pb/* && protoc -I ./proto --go_out=./pb --go_opt=paths=source_relative ./proto/*.proto	

gen: gen_js gen_go

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
	@while true; do \
		sh -c 'cd web && npx webpack --mode=development' ; \
		inotifywait -qr -e modify -e create -e delete -e move --exclude '/\..+' web/*.js web/index.html; \
		echo "File change detected, re-building..." ; \
	done

watch:
	@while true; do \
		rm -f ./pb/* && protoc -I ./proto --go_out=./pb --go_opt=paths=source_relative ./proto/*.proto ; \
		rm -f ./web/pb/* && protoc -I ./proto --js_out=import_style=commonjs:./web/pb ./proto/*.proto ; \
		go run . -port 9090 & \
		pid=$$!; \
		inotifywait -qr -e modify -e create -e delete -e move --exclude '/\..+' **/*.go assets/*; \
		echo "File change detected, restarting server..." ; \
		pkill -P $$pid 2> /dev/null && wait $$pid 2> /dev/null; \
	done

.PHONY: all build gen gen_js gen_go serve run watch