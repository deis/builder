SHORT_NAME ?= builder

include versioning.mk

# dockerized development environment variables
REPO_PATH := github.com/deis/${SHORT_NAME}
DEV_ENV_IMAGE := quay.io/deis/go-dev:0.17.0
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}

# Get the component informtation to a tmp location and get replica count
KUBE := $(shell which kubectl)
ifdef KUBE
$(shell kubectl get rc deis-$(SHORT_NAME) --namespace deis -o yaml > /tmp/deis-$(SHORT_NAME))
DESIRED_REPLICAS=$(shell kubectl get -o template rc/deis-$(SHORT_NAME) --template={{.status.replicas}} --namespace deis)
endif

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
BINARY_DEST_DIR := rootfs/usr/bin
# Common flags passed into Go's linker.
LDFLAGS := "-s -w -X main.version=${VERSION}"
# Docker Root FS
BINDIR := ./rootfs

DEIS_REGISTRY ?= ${DEV_REGISTRY}/

GOTEST := go test --race

all:
	@echo "Use a Makefile to control top-level building of the project."

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

update-changelog:
	${DEV_ENV_PREFIX} -e RELEASE=${WORKFLOW_RELEASE} ${DEV_ENV_IMAGE} gen-changelog.sh \
	  | cat - CHANGELOG.md > tmp && mv tmp CHANGELOG.md

docker-build: build
	docker build --rm -t ${IMAGE} rootfs
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

# Push to a registry that Kubernetes can access.
docker-push:
	docker push ${IMAGE}

deploy: docker-build docker-push
	sed 's#\(image:\) .*#\1 $(IMAGE)#' /tmp/deis-$(SHORT_NAME) | kubectl apply --validate=true -f -
	kubectl scale rc deis-$(SHORT_NAME) --replicas 0 --namespace deis
	kubectl scale rc deis-$(SHORT_NAME) --replicas $(DESIRED_REPLICAS) --namespace deis

.PHONY: all build docker-compile kube-up kube-down deploy
