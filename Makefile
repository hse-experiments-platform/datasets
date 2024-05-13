SRC_DIR_RELATIVE=./api
DST_DIR_RELATIVE=./pkg
CONFIGS_PATH=./configs
BIN_DIR=$(CURDIR)/bin

SWAGGER_FILES=$(shell find $(DST_DIR_RELATIVE) -name '*.swagger.json*')

.PHONY: build
build:
	go build -o ./bin/datasets ./cmd/datasets/main.go

.PHONY: run
run:
	DOTENV_FILE=$(CONFIGS_PATH)/dev/.env go run ./cmd/datasets/main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: imports
imports:
	gci write  .
	 goimports -w .

.PHONY: .copy-swagger
.copy-swagger:
	cp $(SWAGGER_FILES) ./swagger

.PHONY: docker-build
docker-build:
	docker build -t datasets:latest -f ./build/Dockerfile .

.PHONY: docker-run
docker-run:
	docker compose -f deployments/dev/docker-compose.yaml up --build

.PHONY: docker-light-run
docker-light-run:
	docker compose -f deployments/dev/docker-compose.yaml up

.PHONY: docker-run
docker-run-background:
	docker compose -f deployments/dev/docker-compose.yaml up --build -d

.PHONY: start-infra
start-infra:
	docker compose -f deployments/dev/docker-compose.yaml up --build -d db-migrator minio

.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies)
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	GOBIN=$(BIN_DIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(BIN_DIR) go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@latest
	GOBIN=$(BIN_DIR) go install github.com/mitchellh/gox@latest
	GOBIN=$(BIN_DIR) go install golang.org/x/tools/cmd/goimports@latest

.PHONY: generate
generate:
	# buf mod update
	buf generate api
	make .copy-swagger

.PHONY: sqlc-generate
make sqlc-generate:
	sqlc generate -f internal/pkg/storage/datasets/gen/sqlc.yaml
	sqlc generate -f internal/pkg/storage/common/gen/sqlc.yaml
	sqlc generate -f internal/pkg/storage/chunks/impl/db/gen/sqlc.yaml