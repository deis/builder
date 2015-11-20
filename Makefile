# Short name: Short name, following [a-zA-Z_], used all over the place.
# Some uses for short name:
# - Docker image name
# - Kubernetes service, rc, pod, secret, volume names
SHORT_NAME := builder

# Enable vendor/ directory support.
export GO15VENDOREXPERIMENT=1

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
# VERSION := 0.0.1-$(shell date "+%Y%m%d%H%M%S")

VERSION := 2.0.0
BINARY_DEST_DIR := rootfs/usr/bin
# Common flags passed into Go's linker.
LDFLAGS := "-s -X main.version=${VERSION}"
BINARIES := extract-domain extract-types extract-version generate-buildhook get-app-config get-app-values publish-release-controller yaml2json-procfile
STANDALONE := extract-types  generate-buildhook yaml2json-procfile
# Docker Root FS
BINDIR := ./rootfs

# Legacy support for DEV_REGISTRY, plus new support for DEIS_REGISTRY.
DEV_REGISTRY ?= $$DEV_REGISTRY
DEIS_REGISTY ?= ${DEV_REGISTRY}

# Kubernetes-specific information for RC, Service, and Image.
RCBP := manifests/deis-bp${SHORT_NAME}-rc.yaml
SVCBP := manifests/deis-bp${SHORT_NAME}-service.yaml
# IMAGEBP := ${DEV_REGISTRY}/deis/bp${SHORT_NAME}:${VERSION}
IMAGEBP := deis/bp${SHORT_NAME}:${VERSION}

RCDF := manifests/deis-df${SHORT_NAME}-rc.yaml
SVCDF := manifests/deis-df${SHORT_NAME}-service.yaml
# IMAGEDF := ${DEV_REGISTRY}/deis/df${SHORT_NAME}:${VERSION}
IMAGEDF := deis/df${SHORT_NAME}:${VERSION}

all:
	@echo "Use a Makefile to control top-level building of the project."

# This illustrates a two-stage Docker build. docker-compile runs inside of
# the Docker environment. Other alternatives are cross-compiling, doing
# the build as a `docker build`.
#
# Aaron(2015-11-17) doesn't work, commented out. use 'make docker-build-bpb' instead
#
# build:
# 	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/builder boot.go || exit 1
# 	mkdir -p ${BINDIR}/bin
# 	docker run --rm -v ${PWD}:/app -w /app golang:1.5.1 make docker-compile

docker-build-bpb:
	cp -r bpbuilder/etcd pkg/
	cp bpbuilder/kubectl rootfs/bin/
	cp bpbuilder/entrypoint.sh rootfs/
	cp bpbuilder/deis-slugbuilder.yaml rootfs/etc/
	cp  bpbuilder/builder rootfs/etc/confd/templates/
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/bpbuilder boot.go || exit 1
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/fetcher bpbuilder/fetcher/fetcher.go || exit 1
	@$(call check-static-binary,$(BINARY_DEST_DIR)/bpbuilder)
	@$(call check-static-binary,$(BINARY_DEST_DIR)/fetcher)
	for i in $(BINARIES); do \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/$$i pkg/src/$$i.go || exit 1; \
	done
	@echo "Past go compiling"
	@for i in $(BINARIES); do \
		$(call check-static-binary,$(BINARY_DEST_DIR)/$$i); \
	done
	l1=`grep -n "install docker" rootfs/Dockerfile| cut -d ":" -f1`
	l2=`grep -n "end docker install" rootfs/Dockerfile| cut -d ":" -f1`
	sed "$l1,$l2 d" rootfs/Dockerfile.tmp >rootfs/Dockerfile
	docker build -t $(IMAGEBP) rootfs
	perl -pi -e "s|image: [a-z0-9.:]+\/deis\/bp${SHORT_NAME}:[0-9a-z-.]+|image: ${IMAGEBP}|g" ${RCBP}
	rm -rf pkg/etcd
	rm rootfs/entrypoint.sh
	rm rootfs/bin/kubeclt
	rm rootfs/usr/bin/*
	rm rootfs/Dockerfile
	rm rootfs/etc/confd/templates/builder
# For cases where build is run inside of a container.

docker-build-dfb:
	cp -r dfbuilder/etcd pkg/
	cp dfbuilder/entrypoint.sh rootfs/
	cp dfbuilder/builder rootfs/etc/confd/templates/
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/dfbuilder boot.go || exit 1
	@$(call check-static-binary,$(BINARY_DEST_DIR)/dfbuilder)
	for i in $(BINARIES); do \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags '-s' -o $(BINARY_DEST_DIR)/$$i pkg/src/$$i.go || exit 1; \
	done
	@for i in $(BINARIES); do \
		$(call check-static-binary,$(BINARY_DEST_DIR)/$$i); \
	done
	cp rootfs/Dockerfile.tmp rootfs/Dockerfile
	docker build -t $(IMAGEDF) rootfs
	perl -pi -e "s|image: [a-z0-9.:]+\/deis\/df${SHORT_NAME}:[0-9a-z-.]+|image: ${IMAGEDF}|g" ${RCDF}
	rm -rf pkg/etcd
	rm rootfs/entrypoint.sh
	rm rootfs/Dockerfile
	rootfs/etc/confd/templates/builder


# Push to a registry that Kubernetes can access.
docker-push-bp:
	docker push ${IMAGEBP}

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

define check-static-binary
	  if file $(1) | egrep -q "(statically linked|Mach-O)"; then \
	    echo ""; \
	  else \
	    echo "The binary file $(1) is not statically linked. Build canceled"; \
	    exit 1; \
	  fi
endef
