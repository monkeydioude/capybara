.PHONY: start

start: build run

build:
	@go build

run:
	@./capybara ./config.json
