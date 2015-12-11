# Deis Builder v2

[![Build Status](https://travis-ci.org/deis/builder.svg?branch=master)](https://travis-ci.org/deis/builder) [![Go Report Card](http://goreportcard.com/badge/deis/builder)](http://goreportcard.com/report/deis/builder)

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
