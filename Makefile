include .env
export $(shell sed 's/=.*//' .env)

run_cmd := go run

BINARY		:= homed-service
NAME		:= ${REGISTRY}
VERSION		:= $$(git describe --tags `git rev-list --tags --max-count=1`)
TAG			:= ${VERSION}-$$(git rev-parse --short HEAD)
IMAGE		:= ${NAME}\:${TAG}
LATEST		:= ${NAME}\:latest

run-http: 
	@$(run_cmd) cmd/http/main.go

check-swagger:
	@which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
	@cd cmd/http && swagger generate spec -o ./swagger.yaml --scan-models

serve-swagger: check-swagger
	@cd cmd/http && swagger serve -F=swagger swagger.yaml

build:
	@echo ">>> build binary"
	@go build -ldflags="-w -s" -o output/${BINARY} ./cmd/... && upx output/${BINARY}
	@echo ">>> finished"

build-with-tag:
	@echo ">>> build binary"
	@GOARCH=amd64 GOOS=windows go build -ldflags="-w -s" -o output/${BINARY}_windows_${TAG}.exe ./cmd/... && upx output/${BINARY}_windows_${TAG}.exe
	@GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o output/${BINARY}_linux_${TAG} ./cmd/... && upx output/${BINARY}_linux_${TAG}
	@echo ">>> finished"