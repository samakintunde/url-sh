BINARY_NAME=main
.SHELLFLAGS = -e

.PHONY: build-server build-web build dev-server dev-web dev install clean-server vet

build-server:
	GOARCH=amd64 GOOS=darwin go build -o bin/$(BINARY_NAME)-darwin .
	GOARCH=amd64 GOOS=linux go build -o bin/$(BINARY_NAME)-linux .
	GOARCH=amd64 GOOS=windows go build -o bin/$(BINARY_NAME)-windows .

build-web:
	yarn -cwd web
	yarn --cwd web build

build:
	$(MAKE) build-web
	$(MAKE) build-server

dev-server:
	air

dev-web:
	yarn --cwd web dev

dev:
	$(MAKE) dev-server & $(MAKE) dev-web

install:
	yarn --cwd web

clean-web:
	rm -rf web/dist

clean-server:
	go clean
	rm bin/${BINARY_NAME}-darwin
	rm bin/${BINARY_NAME}-linux
	rm bin/${BINARY_NAME}-windows

clean:
	$(MAKE) clean-web & $(MAKE) clean-server

vet:
	sqlc vet

start-darwin:
	yarn --cwd web build
	CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin go build -o bin/$(BINARY_NAME)-darwin .
	bin/${BINARY_NAME}-darwin
