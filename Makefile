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
	go test -count=1 .

## prepare: Prepare test environment
.PHONY: prepare
prepare:
	@echo "Prepare MYSQL Test Environment"
	@docker run -d --name norm_test_mysql \
        -e MYSQL_ROOT_PASSWORD=123456 \
        -e MYSQL_CHARSET=utf8mb4 \
        -e MYSQL_COLLATION=utf8mb4_unicode_ci \
        -e MYSQL_CHARACTER_SET_SERVER=utf8mb4 \
        -e MYSQL_COLLATION_SERVER=utf8mb4_unicode_ci \
        -p 6033:3306 \
        mysql:8.4
	@echo "Waiting for MySQL To Initialize..."
	@sleep 20
	@docker exec -i norm_test_mysql mysql -uroot -p123456 --default-character-set=utf8mb4 < ./test/ddl.sql
	@goctl model mysql ddl --style go_zero --src ./test/ddl.sql --dir ./test
	@echo "Prepare MYSQLTest Environment Over"
    @echo "Prepare ClickHouse Test Environment"
	@docker run -d --name norm_test_clickhouse \
		-p 8123:8123 \
		-p 9000:9000 \
		yandex/clickhouse-server:latest
	@echo "Waiting for ClickHouse To Initialize..."
	@sleep 20


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

## coverage: Run tests with coverage and generate a report
.PHONY: coverage
coverage:
	@mkdir -p test
	@echo "Running tests with coverage..."
	@go test -coverprofile=./test/coverage.out ./...
	@echo "Generating coverage report..."
	@go tool cover -html=./test/coverage.out
	@echo "Coverage report generated at ./test/coverage.html"