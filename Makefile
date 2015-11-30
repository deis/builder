# Short name: Short name, following [a-zA-Z_], used all over the place.
# Some uses for short name:
# - Docker image name
# - Kubernetes service, rc, pod, secret, volume names
SHORT_NAME ?= builder

# Enable vendor/ directory support.
export GO15VENDOREXPERIMENT=1

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
VERSION ?= 0.0.1-$(shell date "+%Y%m%d%H%M%S")
BINARY_DEST_DIR := rootfs/usr/bin
# Common flags passed into Go's linker.
LDFLAGS := "-s -X main.version=${VERSION}"
IMAGE_PREFIX ?= deis
BINARIES := extract-domain extract-types extract-version generate-buildhook get-app-config get-app-values publish-release-controller yaml2json-procfile
STANDALONE := extract-types  generate-buildhook yaml2json-procfile
# Docker Root FS
BINDIR := ./rootfs

# Legacy support for DEV_REGISTRY, plus new support for DEIS_REGISTRY.
DEV_REGISTRY ?= $$DEV_REGISTRY
DEIS_REGISTRY ?= ${DEV_REGISTRY}/

# Kubernetes-specific information for RC, Service, and Image.
RC := manifests/deis-${SHORT_NAME}-rc.yaml
SVC := manifests/deis-${SHORT_NAME}-service.yaml
IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${VERSION}


all:
	@echo "Use a Makefile to control top-level building of the project."

bootstrap:
	glide up

# This illustrates a two-stage Docker build. docker-compile runs inside of
# the Docker environment. Other alternatives are cross-compiling, doing
# the build as a `docker build`.
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/builder boot.go || exit 1
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/fetcher fetcher/fetcher.go || exit 1
	@$(call check-static-binary,$(BINARY_DEST_DIR)/builder)
	@$(call check-static-binary,$(BINARY_DEST_DIR)/fetcher)
	for i in $(BINARIES); do \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/$$i pkg/src/$$i.go || exit 1; \
	done
	@for i in $(BINARIES); do \
		$(call check-static-binary,$(BINARY_DEST_DIR)/$$i); \
	done

docker-build: build
	docker build --rm -t $(IMAGE) rootfs
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
	kubectl delete rc deis-builder

test:
	@echo "Implement functional tests in _tests directory"

.PHONY: all build docker-compile kube-up kube-down deploy

define check-static-binary
	  if file $(1) | egrep -q "(statically linked|Mach-O)"; then \
	    echo ""; \
	  else \
	    echo "The binary file $(1) is not statically linked. Build canceled"; \
	    exit 1; \
	  fi
endef
