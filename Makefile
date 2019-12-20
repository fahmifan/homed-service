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

build:
	@echo ">>> build binary"
	@go build -ldflags="-w -s" -o output/${BINARY} ./cmd/... && upx output/${BINARY}
	@echo ">>> finished"

build-with-tag:
	@echo ">>> build windows binary"
	@GOARCH=amd64 GOOS=windows go build -ldflags="-w -s" -o output/${BINARY}_windows_${TAG}.exe ./cmd/... && upx output/${BINARY}_windows_${TAG}.exe
	@echo ">>> build linux binary"
	@GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o output/${BINARY}_linux_${TAG} ./cmd/... && upx output/${BINARY}_linux_${TAG}
	@echo ">>> build arm binary"
	@GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -o output/${BINARY}_arm7_${TAG} ./cmd/... && upx output/${BINARY}_arm7_${TAG}
	@echo ">>> finished"

build-multi:
	@git checkout ${VERSION} > /dev/null 2>&1
	@echo ">>> build windows binary"
	@GOARCH=amd64 GOOS=windows go build -ldflags="-w -s" -o output/${BINARY}_windows_${VERSION}.exe ./cmd/... && upx output/${BINARY}_windows_${VERSION}.exe
	@echo ">>> build linux binary"
	@GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o output/${BINARY}_linux_${VERSION} ./cmd/... && upx output/${BINARY}_linux_${VERSION}
	@echo ">>> build arm binary"
	@GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -o output/${BINARY}_arm7_${VERSION} ./cmd/... && upx output/${BINARY}_arm7_${VERSION}
	@echo ">>> finished"
	@git checkout master > /dev/null 2>&1


changelog:
ifdef version
	@git-chglog --next-tag $(version) -o CHANGELOG.md
else
	@git-chglog -o CHANGELOG.md
endif

push-master:
	@git push origin master
	@git push mirror master
