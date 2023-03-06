build:
	@go build -o bin/chip8 -x

run: build
	./bin/chip8

test:
	go test ./...
