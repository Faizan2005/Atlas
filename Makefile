build:
		@go build -o bin/atlas

run: build
		@./bin/atlas

test:
		@go test -v ./...