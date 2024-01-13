# Example of Bootstrapping OPA Enabled Service

## Prerequisites

- Golang 1.20+
- GNU Makefile
- Docker for Mac
- `$GOBIN` & `$GOROOT` are correctly set
- `GO111MODULE` is "on"
- OPA Commandline

### Install OPA CLI

#### Option 1: Homebrew
```shell
brew install opa
```

#### Option 2: Manual Install

See [Download OPA](https://www.openpolicyagent.org/docs/latest/#1-download-opa)

## Step-by-Step Bootstrapping

### 1. Bootstrap the service via `lanai-cli codegen`
See Instructions in [skeleton-service](../skeleton/README.md)

### 2. Write Policies

After code generation, OPA policy stubs are generated in `./policies` folder.  

Follow the instruction documented in generated `./policies/README.md` to compose and upload
polices into Policy Service.

### 3. Run Service

The generated service requires [Policy Service](https://cto-github.cisco.com/NFV-BU/cda-policy-service) 
and [Auth Service](https://cto-github.cisco.com/NFV-BU/cda-auth-service) to run at localhost

## Appendix 1: Useful Admin Endpoints

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