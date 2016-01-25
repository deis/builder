SHORT_NAME ?= builder

# Enable vendor/ directory support.
export GO15VENDOREXPERIMENT=1

# dockerized development environment variables
REPO_PATH := github.com/deis/${SHORT_NAME}
DEV_ENV_IMAGE := quay.io/deis/go-dev:0.3.0
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -e GO15VENDOREXPERIMENT=1 -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
VERSION ?= git-$(shell git rev-parse --short HEAD)
BINARY_DEST_DIR := rootfs/usr/bin
# Common flags passed into Go's linker.
LDFLAGS := "-s -X main.version=${VERSION}"
IMAGE_PREFIX ?= deis
# Docker Root FS
BINDIR := ./rootfs

DEIS_REGISTRY ?= ${DEV_REGISTRY}/

# Kubernetes-specific information for RC, Service, and Image.
RC := manifests/deis-${SHORT_NAME}-rc.yaml
SVC := manifests/deis-${SHORT_NAME}-service.yaml
IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${VERSION}


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
	${DEV_ENV_PREFIX} -e CGO_ENABLED=0 ${DEV_ENV_IMAGE} go build -a -installsuffix cgo -ldflags ${LDFLAGS} -o ${BINARY_DEST_DIR}/boot boot.go || exit 1
	@$(call check-static-binary,$(BINARY_DEST_DIR)/boot)

test:
	${DEV_ENV_CMD} go test `glide nv`

docker-build:
	docker build --rm -t ${IMAGE} rootfs
	perl -pi -e "s|image: [a-z0-9.:]+\/deis\/bp${SHORT_NAME}:[0-9a-z-.]+|image: ${IMAGE}|g" ${RC}

# Push to a registry that Kubernetes can access.
docker-push:
	docker push ${IMAGE}

# Deploy is a Kubernetes-oriented target
deploy: kube-service kube-rc

# Some things, like services, have to be deployed before pods. This is an
# example target. Others could perhaps include kube-secret, kube-volume, etc.
kube-service:
	kubectl create -f ${SVC}

# When possible, we deploy with RCs.
kube-rc:
	kubectl create -f ${RC}

kube-clean:
	kubectl delete rc deis-${SHORT_NAME}
	kubectl delete svc deis-${SHORT_NAME}

.PHONY: all build docker-compile kube-up kube-down deploy

define check-static-binary
	  if file $(1) | egrep -q "(statically linked|Mach-O)"; then \
	    echo ""; \
	  else \
	    echo "The binary file $(1) is not statically linked. Build canceled"; \
	    exit 1; \
	  fi
endef
