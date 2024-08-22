BINARY_NAME=main
.SHELLFLAGS = -e

.PHONY:

help:
	@echo "Typical commands (more see below):"
	@echo "  make build                      - Build web app and server"
	@echo "  make web                        - Build the web app"
	@echo "  make check                      - Run all tests, vetting/formatting checks and linters"
	@echo
	@echo "Build everything:"
	@echo "  make build                      - Build web app and server"
	@echo "  make clean                      - Clean build/dist folders"
	@echo "  make server                     - Build server"
	@echo
	@echo "Run Development server"
	@echo "  make dev-server                 - Run server dev server with Air"
	@echo "  make dev-web                    - Run web dev server with Vite"
	@echo
	@echo "Build server"
	@echo "  make linux-server               - Build web & server (current arch, Linux)"
	@echo "  make darwin-server              - Build client & server (current arch, macOS)"
	@echo
	@echo "Build web app:"
	@echo "  make web                        - Build the web app"
	@echo "  make web-deps                   - Install web app dependencies (yarn install the universe)"
	@echo "  make web-lint                   - Run eslint and typecheck on the web app"
	@echo "  make web-fmt                    - Run prettier on the web app"
	@echo
	@echo "Test/check:"
	@echo "  make test                       - Run tests"
	@echo
	@echo "Lint/format:"
	@echo "  make fmt                        - Run 'go fmt'"
	@echo "  make vet                        - Run 'go vet'"
	@echo "  make lint                       - Run 'golint'"
	@echo "  make staticcheck                - Run 'staticcheck'"

# Build
build: web server

web: web-deps web-build

web-deps:
	yarn --cwd web

web-build:
	yarn --cwd web build

server: darwin-server

linux-server:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -o bin/$(BINARY_NAME)-linux .

darwin-server:
	GOARCH=amd64 GOOS=darwin CGO_ENABLED=1 go build -o bin/$(BINARY_NAME)-darwin .

# Development
dev: dev-web dev-server

dev-server:
	DEBUG=true air

dev-web:
	yarn --cwd web dev


# Clean
clean: clean-web clean-server

clean-web:
	rm -rf web/dist

clean-server:
	go clean
	rm -f bin/${BINARY_NAME}-darwin
	rm -f bin/${BINARY_NAME}-linux
	rm -f tmp/main # Air's temporary binary

# Test/Check
web-lint:
	yarn --cwd web lint
	yarn --cwd web typecheck

web-fmt:
	yarn --cwd web format

vet:
	go vet ./...
	sqlc vet

test:
	go test ./...

sqlc:
	sqlc generate
