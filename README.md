# Deis Builder v2

[![Build Status](https://travis-ci.org/deis/builder.svg?branch=master)](https://travis-ci.org/deis/builder) [![Go Report Card](http://goreportcard.com/badge/deis/builder)](http://goreportcard.com/report/deis/builder) [![Docker Repository on Quay](https://quay.io/repository/deisci/builder/status "Docker Repository on Quay")](https://quay.io/repository/deisci/builder)

Deis (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage
applications on your own servers. Deis builds on [Kubernetes](http://kubernetes.io/) to provide
a lightweight, [Heroku-inspired](http://heroku.com) workflow.

## Work in Progress

![Deis Graphic](https://s3-us-west-2.amazonaws.com/get-deis/deis-graphic-small.png)

Deis Builder v2 is changing quickly. Your feedback and participation are more than welcome, but be
aware that this project is considered a work in progress.

# About

This package provides a the Deis Builder, a git server to respond to `git push`es from clients. When it receives a push, it takes the following high level steps:

1. Accepts the code and writes to the local file system
2. Calls `git archive` to produce a tarball (i.e. a `.tar.gz` file) on the local file system
3. Saves the tarball according to the following rules:
  - If the `DEIS_MINIO_SERVICE_HOST` and `DEIS_MINIO_SERVICE_PORT` environment variables exist, uses the [`mc`](https://github.com/minio/mc) client to save to the [Minio](https://github.com/minio/minio) server at `http://$DEIS_MINIO_SERVICE_HOST:$DEIS_MINIO_SERVICE_HOST`
  - Otherwise, if the `DEIS_OUTSIDE_STORAGE_HOST` and `DEIS_OUTSIDE_STORAGE_PORT` environment variables exist, uses the [`mc`](https://github.com/minio/mc) client to save to S3 server (or server that adheres to the S3 API) at `https://$DEIS_OUTSIDE_STORAGE_HOST:$DEIS_OUTSIDE_STORAGE_PORT` (this functionality is currently waiting for merge at https://github.com/deis/builder/pull/21).
4. Starts a builder pod according to these rules:
  - If a `Dockerfile` is present, starts a [`dockerbuilder`](https://github.com/deis/dockerbuilder) pod, configured to download the code to build from the URL computed in the previous step (`dockerbuilder` and Dockerfile builder are not currently supported. See https://github.com/deis/dockerbuilder/pull/1 for prototype `dockerbuilder` code).
  - Otherwise, starts a [`slugbuilder`](https://github.com/deis/slugbuilder) pod, configured to download the code to build from the URL computed in the previous step.

# Hacking Builder

First, install [helm](http://helm.sh) and [boot up a kubernetes cluster][install-k8s]. Next, add the
`deis` repository to your chart list:

```console
$ helm repo add deis https://github.com/deis/charts
```

Then, install the Deis chart!

```console
$ helm install deis/deis
```

The chart will install the entire Deis platform onto Kubernetes. You can monitor all the pods that it installs by running:

```console
$ kubectl get pods --namespace=deis
```

Once this is done, SSH into a Kubernetes minion, and run the following:

```
$ curl -sSL http://deis.io/deis-cli/install.sh | sh
$ sudo mv deis /bin
$ kubectl get service deis-workflow
$ deis register 10.247.59.157 # or the appropriate CLUSTER_IP
$ ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
$ eval $(ssh-agent) && ssh-add ~/.ssh/id_rsa
$ deis keys:add ~/.ssh/id_rsa.pub
$ deis create --no-remote
Creating Application... done, created madras-radiator
$ deis pull deis/example-go -a madras-radiator
Creating build... ..o
```

If you want to hack on a new feature, rebuild your code, build the deis/builder image and push it to a Docker registry. The `$DEIS_REGISTRY` environment variable must point to a registry accessible to your Kubernetes cluster. If you're using a locally hosted Docker registry, you may need to configure the Docker engines on your Kubernetes nodes to allow `--insecure-registry 192.168.0.0/16` (or the appropriate address range).

```console
$ make build docker-build docker-push
```

Next, you'll want to remove the `deis-builder` [replication controller](http://kubernetes.io/v1.1/docs/user-guide/replication-controller.html) and re-create it to run your new image.

```console
kubectl delete --namespace=deis rc deis-builder
make kube-rc
```

## Installation

The following steps assume that you have the [Docker CLI](https://docs.docker.com/) and [Kubernetes CLI](http://kubernetes.io/v1.1/docs/user-guide/kubectl-overview.html) installed and correctly configured.

```
make deploy kube-service
```

## License

Copyright 2013, 2014, 2015 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[install-k8s]: http://kubernetes.io/gettingstarted/
