.version: $(shell go run github.com/simon-engledew/gosrc@master ./cmd/sqljsondot) $(shell go run github.com/simon-engledew/gosrc@master ./cmd/sqljsondump)
	docker build . --iidfile $@ -t ghcr.io/simon-engledew/sqljson:latest

Dockerfile: .version
