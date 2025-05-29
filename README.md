# go-lanai

![Tests](https://github.com/cisco-open/go-lanai/actions/workflows/ci.yml/badge.svg?branch=main)
[![Coverage](https://cisco-open.github.io/go-lanai/reports/main/coverage-badge.svg)](https://cisco-open.github.io/go-lanai/reports/main/code-coverage-results.html)
[![Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-green.svg)](https://opensource.org/license/apache-2-0/)

go-lanai is an application frameworks and a set of modules that make writing applications easy. It provides
everything you need to embrace go-lang in an enterprise environment. You can use it for many kinds of architectures 
depending on your need, from microservice to standalone application. 

go-lanai is inspired by the Spring framework and the family of Spring projects. Developers looking to port a Spring application
to go-lang can use go-lanai as a feature by feature replacement of Spring in most cases.

go-lanai's dependency injection functionality is provided by [uber-go/fx](https://github.com/uber-go/fx). Basic understanding
of dependency injection is helpful when using go-lanai.

## Examples

- The [auth](examples/auth/README.md) example is a web service that can serve as the Oauth2 authorization server.
- The [data](examples/database/README.md) example is a web service that stores and retrieves data from the database.
- In the [skeleton](examples/skeleton/README.md) example, we demonstrate how to create a service from scratch using an OpenAPI spec.
- In the [opa](examples/opa/README.md) example, we create a web service that uses open policy agent for RBAC.

## Tutorials

- The [web app security](docs/tutorials/Web-app-security.md) tutorial shows you how to build a web application with a private
API, while using SAML sign on as the authentication method.

## Documentations

- [docs/Develop.md](docs/Develop.md) - guide on setting up go-lanai based project for developers.
- [docs/CICD.md](docs/CICD.md) - guide on setting up CI/CD for go-lanai based project.
- [cmd/lanai-cli/initcmd/README.md](cmd/lanai-cli/initcmd/README.md) - documentation for the make file templates and cli tools provided by go-lanai.
- [cmd/lanai-cli/codegen/README.md](cmd/lanai-cli/codegen/README.md) - documentation for cli tool that generates code based on an OpenAPI contract.

### Documentation for go-lanai Modules

Explore the go-lanai modules:

- actuator
- [appconfig](pkg/appconfig/README.md)
- aws
- [bootstrap](pkg/bootstrap/README.md)
- [certs](pkg/certs/README.md)
- consul
- [data](pkg/data/README.md)
- discovery
- dsync
- integrate
- [kafka](pkg/kafka/README.md)
- log
- [migration](pkg/migration/README.md)
- opa
- [opensearch](pkg/opensearch/README.md)
- [profiler](pkg/profiler/README.md)
- redis
- scheduler
- [security](pkg/security/README.md)
- swagger
- tenancy
- tracing
- vault
- [web](pkg/web/README.md)
- [test](test/README.md)

# Contributing to `go-lanai`

Thanks for your interest in contributing! There are many ways to contribute to this project. 

Get started with our [Contributing Guide (WIP)](CONTRIBUTING.md).

Please note that our contributing guide is still "work in progress".