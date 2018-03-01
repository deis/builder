
|![](https://upload.wikimedia.org/wikipedia/commons/thumb/1/17/Warning.svg/156px-Warning.svg.png) | Deis Workflow is no longer maintained.<br />Please [read the announcement](https://deis.com/blog/2017/deis-workflow-final-release/) for more detail. |
|---:|---|
| 09/07/2017 | Deis Workflow [v2.18][] final release before entering maintenance mode |
| 03/01/2018 | End of Workflow maintenance: critical patches no longer merged |
| | [Hephy](https://github.com/teamhephy/workflow) is a fork of Workflow that is actively developed and accepts code contributions. |

# Deis Builder v2

[![Build Status](https://ci.deis.io/job/builder/badge/icon)](https://ci.deis.io/job/builder) [![codecov](https://codecov.io/gh/deis/builder/branch/master/graph/badge.svg)](https://codecov.io/gh/deis/builder)
[![Go Report Card](https://goreportcard.com/badge/github.com/deis/builder)](https://goreportcard.com/report/github.com/deis/builder)[![codebeat badge](https://codebeat.co/badges/e29e5e2b-531d-4374-810b-f05053c47688)](https://codebeat.co/projects/github-com-deis-builder) [![Docker Repository on Quay](https://quay.io/repository/deisci/builder/status "Docker Repository on Quay")](https://quay.io/repository/deisci/builder)

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes][k8s-home] cluster, making it easy to deploy and manage applications on your own servers.

For more information about Deis Workflow, please visit the main project page at https://github.com/deis/workflow.

We welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

# About

The builder is primarily a git server that responds to `git push`es by executing either the `git-receive-pack` or `git-upload-pack` hook. After it executes one of those hooks, it takes the following high level steps in order:

1. Calls `git archive` to produce a tarball (i.e. a `.tar.gz` file) on the local file system
2. Saves the tarball to centralized object storage according to the following rules:
	- If the `BUILDER_STORAGE` environment variable is other than `minio`, attempts to create the appropriate storage driver and saves using this driver.
  - Otherwise, if `BUILDER_STORAGE` is `minio` and the `DEIS_MINIO_SERVICE_HOST` and `DEIS_MINIO_SERVICE_PORT` environment variables exist (these are standard [Kubernetes service discovery environment variables](http://kubernetes.io/docs/user-guide/services/#environment-variables)), saves to the [S3 API][s3-api-ref] compatible server at `http://$DEIS_MINIO_SERVICE_HOST:$DEIS_MINIO_SERVICE_HOST`
3. Starts a new [Kubernetes Pod](http://kubernetes.io/docs/user-guide/pods/) to build the code, according to the following rules:
  - If a `Dockerfile` is present in the codebase, starts a [`dockerbuilder`](https://github.com/deis/dockerbuilder) pod, configured to download the code to build from the URL computed in the previous step.
  - Otherwise, starts a [`slugbuilder`](https://github.com/deis/slugbuilder) pod, configured to download the code to build from the URL computed in the previous step.

# Supported Off-Cluster Storage Backends

Builder currently supports the following off-cluster storage backends:

* GCS
* AWS/S3
* Azure
* Swift

# Development

The Deis project welcomes contributions from all developers. The high level process for development matches many other open source projects. See below for an outline.

* Fork this repository
* Make your changes
* [Submit a pull request][prs] (PR) to this repository with your changes, and unit tests whenever possible
	* If your PR fixes any [issues][issues], make sure you write `Fixes #1234` in your PR description (where `#1234` is the number of the issue you're closing)
* The Deis core contributors will review your code. After each of them sign off on your code, they'll label your PR with `LGTM1` and `LGTM2` (respectively). Once that happens, a contributor will merge it

## Docker Based Development Environment

The preferred environment for development uses [the `go-dev` Docker image](https://github.com/deis/docker-go-dev). The tools described in this section are used to build, test, package and release each version of Deis.

To use it yourself, you must have [make](https://www.gnu.org/software/make/) installed and Docker installed and running on your local development machine.

If you don't have Docker installed, please go to https://www.docker.com/ to install it.

After you have those dependencies, grab Go dependencies with `make bootstrap`, build your code with `make build` and execute unit tests with `make test`.

## Native Go Development Environment

You can also use the standard `go` toolchain to build and test if you prefer. To do so, you'll need [glide](https://github.com/Masterminds/glide) 0.9 or above and [Go 1.6](http://golang.org) or above installed.

After you have those dependencies, you can build and unit-test your code with `go build` and `go test $(glide nv)`, respectively.

Note that you will not be able to build or push Docker images using this method of development.

# Testing

The Deis project requires that as much code as possible is unit tested, but the core contributors also recognize that some code must be tested at a higher level (functional or integration tests, for example).

The [end-to-end tests](https://github.com/deis/workflow-e2e) repository has our integration tests. Additionally, the core contributors and members of the community also regularly [dogfood](https://en.wikipedia.org/wiki/Eating_your_own_dog_food) the platform. Since this particular component is at the center of much of the Deis Workflow platform, we find it especially important to dogfood it.

## Running End-to-End Tests

Please see [README.md](https://github.com/deis/workflow-e2e/blob/master/README.md) on the end-to-end tests repository for instructions on how to set up your testing environment and run the tests.

## Dogfooding

Please follow the instructions on the [official Deis docs](http://docs-v2.readthedocs.org/en/latest/installing-workflow/installing-deis-workflow/) to install and configure your Deis Workflow cluster and all related tools, and deploy and configure an app on Deis Workflow.


[s3-api-ref]: http://docs.aws.amazon.com/AmazonS3/latest/API/APIRest.html
[install-k8s]: http://kubernetes.io/gettingstarted/
[k8s-home]: http://kubernetes.io
[issues]: https://github.com/deis/builder/issues
[prs]: https://github.com/deis/builder/pulls
[v2.18]: https://github.com/deis/workflow/releases/tag/v2.18.0
