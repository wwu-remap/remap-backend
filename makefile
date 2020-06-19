.PHONY: all build update

all: build

bin:
	@mkdir -p $@

build: bin/remap-server

bin/remap-server: main.go | bin
	@go build -o $@

clean:
	@rm -rf bin

update:
	@go get -u
	@go mod tidy
