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
	@echo "prepare test environment"
	@docker run -d --name norm_test_mysql \
		-e MYSQL_ROOT_PASSWORD=123456 \
		-p 6033:3306 \
		-v $(PWD)/test/ddl.sql:/docker-entrypoint-initdb.d/init.sql \
		mysql:8.4
	@echo "Waiting for MySQL to initialize..."
	@sleep 15
	@goctl model mysql ddl --style go_zero --src ./test/ddl.sql --dir ./test
	@echo "prepare test environment over"

## clean: Clean test environment
.PHONY: clean
clean:
	@docker stop norm_test_mysql && docker rm norm_test_mysql

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem

## benchnote: Run benchmark tests and save results to a timestamped file
.PHONY: benchnote
benchnote:
	@mkdir -p bench
	@DATE=$$(date +%Y-%m-%d); \
	LATEST_NUM=$$(ls -1 bench/$${DATE}_* 2>/dev/null | sed -e "s/bench\/$${DATE}_//" | sort -n | tail -1 || echo "0"); \
	NEXT_NUM=$$(($$LATEST_NUM + 1)); \
	FILENAME=bench/$${DATE}_$${NEXT_NUM}; \
	echo "Running benchmark tests, saving results to $${FILENAME}"; \
	go test -bench=. -benchmem | tee $${FILENAME}