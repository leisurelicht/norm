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

## prepare: Prepare test environment
.PHONY: prepare
prepare:
	echo "prepare test environment"
	@docker run -d --name norm_test_mysql -e MYSQL_ROOT_PASSWORD=123456 -p 3306:3306 mysql:8.4
	@sleep 10
	@mysql -h127.0.0.1 -uroot -p123456 --silent <./test/ddl.sql
	@goctl model mysql ddl --style go_zero --src ./test/ddl.sql --dir ./test;
	echo "prepare test environment over"

## clean: Clean test environment
.PHONY: clean
clean:
	@docker stop norm_test_mysql && docker rm norm_test_mysql