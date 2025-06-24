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

envoy:
	@docker run --rm -v "$$(pwd)"/web/envoy.yaml:/etc/envoy/envoy.yaml:ro --network=host envoyproxy/envoy:v1.22.0

.PHONY: all build gen gen_js gen_go serve run