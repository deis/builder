# Short name: Short name, following [a-zA-Z_], used all over the place.
# Some uses for short name:
# - Docker image name
# - Kubernetes service, rc, pod, secret, volume names
SHORT_NAME := example

# Enable vendor/ directory support.
export GO15VENDOREXPERIMENT=1

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
VERSION := 0.0.1-$(shell date "+%Y%m%d%H%M%S")

# Common flags passed into Go's linker.
LDFLAGS := "-s -X main.version=${VERSION}"

# Docker Root FS
BINDIR := ./rootfs

# Legacy support for DEV_REGISTRY, plus new support for DEIS_REGISTRY.
DEV_REGISTRY ?= $(eval docker-machine ip deis):5000
DEIS_REGISTY ?= ${DEV_REGISTRY}

# Kubernetes-specific information for RC, Service, and Image.
RC := manifests/deis-${SHORT_NAME}-rc.yaml
SVC := manifests/deis-${SHORT_NAME}-service.yaml
IMAGE := ${DEIS_REGISTRY}/deis/${SHORT_NAME}:${VERSION}

all:
	@echo "Use a Makefile to control top-level building of the project."

# This illustrates a two-stage Docker build. docker-compile runs inside of
# the Docker environment. Other alternatives are cross-compiling, doing
# the build as a `docker build`.
build:
	mkdir -p ${BINDIR}/bin
	docker run --rm -v ${PWD}:/app -w /app golang:1.5.1 make docker-compile

# For cases where build is run inside of a container.
docker-compile:
	go build -o ${BINDIR}/bin/boot -a -installsuffix cgo -ldflags ${LDFLAGS} boot.go

# For cases where we're building from local
# We also alter the RC file to set the image name.
docker-build:
	docker build --rm -t ${IMAGE} rootfs
	perl -pi -e "s|[a-z0-9.:]+\/deis\/${SHORT_NAME}:[0-9a-z-.]+|${IMAGE}|g" ${RC}

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
	kubectl delete rc deis-example

.PHONY: all build docker-compile kube-up kube-down deploy
