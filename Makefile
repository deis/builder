SHORT_NAME ?= builder

include versioning.mk

# dockerized development environment variables
REPO_PATH := github.com/deis/${SHORT_NAME}
DEV_ENV_IMAGE := quay.io/deis/go-dev:0.20.0
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
BINARY_DEST_DIR := rootfs/usr/bin
# Common flags passed into Go's linker.
LDFLAGS := "-s -w -X main.version=${VERSION}"
# Docker Root FS
BINDIR := ./rootfs

DEIS_REGISTRY ?= ${DEV_REGISTRY}/

GOTEST := go test --race

bootstrap:
	${DEV_ENV_CMD} glide install

glideup:
	${DEV_ENV_CMD} glide up

# This illustrates a two-stage Docker build. docker-compile runs inside of
# the Docker environment. Other alternatives are cross-compiling, doing
# the build as a `docker build`.
build:
	${DEV_ENV_CMD} go build -ldflags ${LDFLAGS} -o ${BINARY_DEST_DIR}/boot boot.go
	${DEV_ENV_CMD} upx -9 ${BINARY_DEST_DIR}/boot

test: test-style test-unit

test-style:
	${DEV_ENV_CMD} lint

test-unit:
	${DEV_ENV_CMD} sh -c '${GOTEST} $$(glide nv)'

test-cover:
	${DEV_ENV_CMD} test-cover.sh

docker-build: build
	docker build --rm -t ${IMAGE} rootfs
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

check-kubectl:
	@if [ -z $$(which kubectl) ]; then \
		echo "kubectl binary could not be located"; \
		exit 2; \
	fi

deploy: check-kubectl docker-build docker-push
	kubectl --namespace=deis patch deployment deis-$(SHORT_NAME) --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"$(IMAGE)"}]'

.PHONY: bootstrap glideup build docker-build test test-style test-unit test-cover deploy
