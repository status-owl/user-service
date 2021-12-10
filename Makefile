GOBIN=${GOPATH}/bin
SPEC_API_V1=spec/api-v1.yaml
SPEC_API_V1_GEN=pkg/api/types.go
OAPI_CODEGEN=$(GOBIN)/oapi-codegen
MOCKGEN=$(GOBIN)/mockgen

default: help

.PHONY = mocks install-tools

help:   ## show this help
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

install-tools:
	@echo "installing mockgen & oapi-codegen"
	@go install github.com/golang/mock/mockgen@v1.6.0
	@go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0

generate:
	@echo "generating mocks and models..."
	@go generate ./...


$(SPEC_API_V1_GEN): $(OAPI_CODEGEN)
	@echo "generating models from api spec..." 
	@oapi-codegen -o $(SPEC_API_V1_GEN) --generate=types --package=api $(SPEC_API_V1)

run:
	@go run ./...  $(ARGS)

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "cleaning up..."
	@rm -f $(SPEC_API_V1_GEN)
	@go clean

services-up:
	@echo "starting services..."
	@docker compose -f docker-compose.yaml up -d

services-down:
	@echo "stopping services..."
	@docker compose -f docker-compose.yaml down

build:
	@echo "Building binary..."
	@go build ./...