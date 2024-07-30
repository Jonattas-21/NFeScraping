DSN=host=localhost port=5432 user=admin password=admin dbname=FinalConcurrency sslmode=disable
REDIS="127.0.0.1:6379"
BINARY_NAME=myapp.exe

## build: builds all binaries
build:
	@go build ./cmd/.
	@echo backend built!

run: build
	@go run ./cmd/main.go

start: run

restart: stop start

test:
	@echo "Testing..."
	go test -v ./...