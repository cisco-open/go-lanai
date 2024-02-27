# For Service Developers and DevOps

## Prerequisites

- GO 1.16+
- Git 2.23.0+
- GNU Make 3.81+
- Docker 20.10.5+
- Proper access to any private GitHub
- GO environment variables such as `$GOPATH` and `$GOBIN` are properly set

## Project Setup

### GO Mod File `go.mod`

The `go.mod` should have at least following content:

```
module github.com/my_organization/my_module

go 1.16

require (
	github.com/cisco-open/go-lanai v0.12.0
)
```

It's optional, but recommended for platform developers, to clone `github.com/cisco-open/go-lanai` 
alongside service projects, and add `replace` using relative path in `go.mod`. 

See `go.mod` [Example](res/Example-Go-Mod.mod)

### Module Descriptor `Module.yml`

In addition to `go.mod`, a descriptor file `Module.yml` is required to provide additional information 
about the service.

`Module.yml` is used by `lanai-cli` (See [Tooling](#tooling)) to generate proper `Makefile` for the service

See `Module.yml` [Example](res/Example-Module.yml)

### Bootstrapping `Makefile`

The bootstrapping `Makefile` helps with installing necessary tools and initializing projects and environment
such as generating `Dockerfile` and additional Makefile components 

`Makefile` template can be found [Here](../cmd/lanai-cli/initcmd/Makefile.tmpl). 

> Note 1: This template can be copied directly with simple file rename.

> Note 2: Alternatively, `lanai-cli` tool can be installed manually, and the CLI can generate the Makefile. 
> See [Tooling](#tooling)

### Git Ignore `.gitignore`

See `.gitignore` [Example](res/Example-gitignore)

### Private GO Repository

To help with access to any private module, the `Makefile` from previous section provides with an automated target to set
up the development environment. If the local environment is already setup, this step can be skipped:

```shell
make init-once
```

> Note 1: The target configure the GO CLI Tool to use `SSH` instead of `https` to get modules in `PRIVATE_MODS`.
> So the local environment need to be properly configured to access the modules in `PRIVATE_MODS` 

> Note 2: This is only required to run once per machine

<br>

## Tooling

`github.com/cisco-open/go-lanai/cmd/lanai-cli` is a command line tool to help with common build/code generation/git tasks. 
It is required to properly set up the project and perform common development tasks.

### Install using GNU Make

Assuming `Makefile` exists (copied from this document):

```
make init CLI_TAG=main
```

The value `CLI_TAG` is usually `main` which point to the latest snapshot version or any stable/released version Git Tag.

The target will:

- Install `lanai-cli` to `$GOBIN`
- Create `Makefile-Build` and `Makefile-Generated`
- Create `build/package/Dockerfile`

> Note 1: `Makefile-Build` and `Dockerfile` are supposed to be committed into GitHub and the command won't overwrite 
> if those files already exist. If overwriting is desired, add `FORCE=true`

> Note 2: `Makefile-Generated` is typically used by CI/CD, and always get overwrite. It should be ignored from version control.


### Manual Install

This operation is only required when bootstrapping a new Service at the first time without `Makefile`.

```
go install github.com/cisco-open/go-lanai/cmd/lanai-cli@main
```

> Note: `@main` can be changed to the latest stable version

After successfully install the CLI, it can be used to generate bootstrapping Makefiles to the current directory:

```
lanai-cli init --force --upgrade -o .
```

### CLI Usage

The CLI tool contains many commands. Use `lanai-cli --help` and `lanai-cli help [command]` for Help

<br>

## PR Verify, Nightly Build, Promoting and Release 

Topics are covered in [CI/CD Documentation](CICD.md)

<br>

## Develop, Test and Build

Typically, developer don't need GNU Make or `lanai-cli` to test/build Service. It can be done by

- Run Test from GoLand IDE; OR
- `go run`, `go test` and `go build` commands

However, the generated `Makefile-Build` provide some basic targets:

- `make generate` 
  
  Invoke `//go:generate` on all packages registered in `Module.yml`

- `make test` 

  run tests on all tests in `pkg` folder with coverage report saved in `dist` folder
  
- `make build [VERSION=<version>] [PRIVATE_MODS=<private.module/path@branch_or_version>]`

  Build executable and copy resources registered in `Module.yml` to `dist` folder, where:
    
    - `VERSION` is set to executable's build-info and can be viewed via `admin/info` endpoint
    - `PRIVATE_MODS` is set to executable's build-info as a reference of upstream private modules.
      Also viewable via `admin/info` endpoint
      
- `make clean` 
  
  clean `dist` folder and run `go clean`

<br>

## Keep Upstream Dependencies Up-To-Date

As mentioned before, it's recommended to check out GO Lanai alongside the project during development and use `replace` in 
`go.mod` with relative path.

In addition, it is recommended to keep `require` section up-to-date pointing to the correct GO Lanai version, at least before creating PR.
This can be easily done by following step

- Edit `go.mod` and change the "version" part of `require` section to the desired branch or version. e.g.

  ```
  require (
      github.com/cisco-open/go-lanai main
  )
  ```

- Run any Module-Aware `go` command, such as `go mod tidy` or `go get`. The "branch" in `go.mod` will be updated
  to proper version tag representing the committed content in the given branch/tag
  


