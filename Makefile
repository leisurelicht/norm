.DEFAULT_GOAL := help

## help: Show this help info.
.PHONY: help
help: Makefile
	@echo "Usage: make <TARGETS> ...\n\nTargets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

## format: Run code format
.PHONY: format
format:
	@gofmt -w .; echo "gofmt over"
	@for file in $(shell find . -name '*.go'); do goimports-reviser -rm-unused -set-alias -format $$file; echo "goimports-reviser ["$$file"] over"; done

## test: Run code test
.PHONY: test
test:
	go test .
