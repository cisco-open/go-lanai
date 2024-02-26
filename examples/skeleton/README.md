# Example of Bootstrapping New Service

## Prerequisites

- Golang 1.20+
- GNU Makefile
- Docker for Mac
- `$GOBIN` & `$GOROOT` are correctly set
- `GO111MODULE` is "on"

## Step-by-Step Bootstrapping

1. Update [Module.yml](Module.yml) for correct service name, port and other attributes if applicable. 
   This file will be used for to setup the Makefile and CI/CD related scripts.
2. Update [go.mod](go.mod) to make sure latest `go-lanai` library is used 
3. Run
   ```shell
   make init CLI_TAG="main"
   ```
   *Note*: If using released version of `go-lanai`, "version" can be used instead of "branch" in `CLI_TAG="..."`

   This step would setup the workspace as following:
   - Setup Golang private repository.
   - Install `lanai-cli`, which is the CLI tool that come with `go-lanai`.
   - Generate `Makefile-Build` for developer day-to-day tasks (build, test, lint, etc.).
   - Generate `Makefile-Generated` for CI/CD tasks, this file is typically excluded form Source Control.
   - Generate `Dockerfile`
   - Install Golang CLI utilities that are required for this project. (`Module.yml` can be used to overwrite this)
4. Update [codegen.yml](codegen.yml) to prepare for code generation. Make sure service name, port, context-path, etc. to be in sync
   with [Module.yml](Module.yml)
5. Update OpenAPI contract document [configs/api-docs-v3.yml](configs/api-docs-v3.yml)
6. Run
   ```shell
   lanai-cli codegen -o ./
   ```
   This step would generate skeleton code based on provided OpenAPI contract and `codegen.yml`.
7. Review `configs/application.yml` and `configs/bootstrap.yml` 
8. Try run the generated service
   ```shell
   go run cmd/<service-name>/main.go
   ```
9. Verify
   - Service is registered with consul
   - Service is healthy
   - APIs are serving
   - Swagger page can be accessed (this would require [auth service](../auth))

## Appendix 1:  Useful Admin Endpoints

### `GET /context-path/admin/info`
Show application information.

### `GET /context-path/admin/health`
Show health status.

### `GET /context-path/admin/env`
Show all properties applied to the running service and their sources.

### `GET/POST /context-path/admin/loggers`
Show and Modify log levels 

## Appendix 1: Common CLI Commands

### Build

```shell
make build
```

### Test

```shell
make test
```

### Code Lint

```shell
make lint
```

### Force Makefile/Dockerfile

```shell
make init FORCE=true
```

### Init with different `lanai-cli` Version

```shell
make init CLI_TAG="<github-branch>"
```

e.g.

```shell
make init CLI_TAG="develop"
```