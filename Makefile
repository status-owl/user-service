GOBIN=${GOPATH}/bin
SPEC_API_V1=spec/api-v1.yaml
SPEC_API_V1_GEN=spec/types.go
OAPI_CODEGEN=$(GOBIN)/oapi-codegen

default: help

help:   ## show this help
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

$(OAPI_CODEGEN):
	@echo "installing oapi-codegen..."
	@go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen

$(SPEC_API_V1_GEN): $(OAPI_CODEGEN)
	@echo "generating models from api spec..." 
	@oapi-codegen -o $(SPEC_API_V1_GEN) --generate=types --package=spec $(SPEC_API_V1)

run: $(SPEC_API_V1_GEN)
	@go run ./...

test: $(SPEC_API_V1_GEN)
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "cleaning up..."
	@rm -f $(SPEC_MODELS)
	@go clean
