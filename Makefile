github-webhook: **/*
	go build -o main

.PHONY: dev
dev:
	godo --watch

.PHONY: init
init:
	dep ensure