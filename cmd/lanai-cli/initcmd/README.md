# Lanai-CLI Init and GNU Make

This package provides a set of templates for `GNU Make`. Together with `lanai-cli`, they provide an opinionated way
that automates common software development tasks.

## An Opinionated Approach

The commands and processes described below represents an opinionated approach to software development lifecycle. It addresses
the case where multiple projects are developed in parallel. For example, a team could be developing a service and a library in
tandem, where the service is dependent on the library. In this document, you will see the library in this example being referred to
as `PRIVATE_MODS`, which stands for "private module". It signifies this module is under development (i.e. private to the development team).
You can still use this pattern if your library module is a public repository but being developed in parallel.

In a project with one or more private modules, our workflow is the following:

1. Developers will check out both the service project and the private modules.
2. Service project's `go.mod` file will require the private modules by branch.
3. Service project's `go.mod` file will have `replace` directives to point to checked out copies of the private modules.
4. Developers are expected to check in the `go.mod` file with the `replace` directive.
5. During the CI/CD process, the dependency for the private modules are updated to the head of the specified branch, 
and the `replace` directives in the `go.mod` file are dropped.
6. When a build is published, a copy of the code with the modified `go.mod` is version tagged and pushed.

This process allows developers to work on the service and library project with ease, while allowing the CI/CD process
to build against the expected code base.

If your project does not have any private modules, simply omit the `PRIVATE_MODS` arguments in the commands.

## Bootstrap a Project

To bootstrap a project from scratch, copy the [`Makefile.tmpl`](Makefile.tmpl) into your project root directory, and rename 
it to `Makefile`. This `Makefile` contains the targets that can bootstrap the rest of the targets needed during the development
lifecycle.

### Setting Up Local Environment

```make init-once```

This target is used to set up your local development environment. You should run this target if your project uses any
private modules, and the private module repositories are not publicly accessible. 

This target takes the `PRIVATE_MODS` argument. This argument takes a comma separated list of private modules.

```make init-once PRIVATE_MODS="github.com/<org>/<my-private-dependency>@v0.1.0,github.my-domain.com/<org>/<repo>@0.2.1```

Running this target will configure your `git` command to use `ssh` instead of `https` to access these module repos. It will
also append these module repos to the `GOPRIVATE` variable so that `go` does not use the public Go module proxy of the public
checksum database.

If your project does not use any private modules, you do not need to run this target.

### Generate Project Specific Make File and Docker File

```make init```

Pre-requisite:
1. a `go.mod` file that has `go-lanai` as a required module.
2. a `Module.yml` that describes your project. The `Module.yml` can have the following entries
```yaml
name: europa # name of the project

execs: # your project's executables. these entries will be used to generate the dockerlaunch.sh file
  europa:
    main: cmd/europa/main.go 
    port: 8080 # specify a port to indicate this binary represents a web app. 
  migrate:
    main: cmd/europa/migrate.go
    type: migrator # specify migrator type to indicate this binary is responsible for database migration.

generates: # packages that have the golang generate directive. they will be used to generate Makefile target that triggers the go generate directive.
  - path: "pkg/swagger"
  - path: "pkg/security/example"

resources: # resources that should be published together with the binary. they will be used to generate Makefile target for copying the resources into the build
  - pattern: "configs" # this is the pattern to find them in the source directory
    output: "configs" # this is the directory to copy them to in the dist directory 
  - pattern: "web"
    output: "web"
```

This target does the following:
1. Installs the `lanai-cli` command line tool. 
2. Invoke the `lanai-cli init` command to generate make files and docker files specific to your project.

The following sections will describe these two steps in detail.

#### Installing `lanai-cli`

`make init` installs the `lanai-cli` command line tool by calling the ```make init-cli``` sub-target which handles a 
variety of use cases.

First case is where the `go.mod` file has a `replace` directive that points `go-lanai` to a local copy. This case is most 
common for developers who have local modification to `go-lanai`. In this case, this target will install `lanai-cli` from 
the local checked out location of `go-lanai`. 

If the local location is not valid, it will install `go-lanai` according to the required version. This is most common in 
the CI/CD pipeline.

If the `go.mod` does not have a `replace` directive for `go-lanai`, it will install according to the required version. 

Except when installing using local copy via the `replace` directive, you can use the `CLI_TAG` argument to override the
version of `lanai-cli` to be installed. The `CLI_TAG` argument can be either a branch name or a tag in the `go-lanai` repo.

i.e. `make init-cli CLI_TAG=main`.

#### Invoking `lanai-cli init`

After installing `lanai-cli`, the `make init` target generates a set of files based on the information from `Module.yml`
by invoking the `lanai-cli init` command.

It will generate the following files:
* `Makefile-Build` This file contains the targets specific to your project. Such as executables to build, resources to copy.
* `Makefile-Generated` This file contains targets that are usually used by CI/CD.
* `Dockerfile` This file will be used to build the image for your service. It will expose ports listed in the `Module.yml` file.
* `dockerlaunch.sh` The entry point of the docker image. It will use the executable listed in the `Module.yml` file.

The recommended practice is to check in `Makefile-Build`, `Dockerfile` and `dockerlaunch.sh`. These files are specific
to your project, and you may want to modify them as you need. `make init` will not overwrite existing files, so your 
changes can persist. If you updated `Module.yml` and want to re-generate them. You need to use the `FORCE` flag. i.e
`make init FORCE=true`

`Makefile-Generated` is project agnostic. It's not expected to be checked in and will be re-generated during the CI/CD process.

## Development

### Running Tests and Linters

The following targets are found in `Makefile-Build` which is generated based on `Makefile-Build.tmpl`.

```make test``` 

Runs tests and produce test coverage report. This is the same target used in PR verification as well so any verification
issue is repeatable locally.

```make lint```

Invoke `go vet` and `golangci-lint` linter.

### Making Builds

```make build```

This command will create a `dist` folder. This folder will contain the executable binaries and any files that the binaries
needs for execution.

This target takes `VERSION` and `PRIVATE_MODS` arguments. `VERSION` is the version number you want to give to this build. 
`PRIVATE_MODS` represents the private modules in your project's dependency. The build tool will look up the entries in
`PRIVATE_MODS` in your service's `go.mod` and record their versions. All the information will be set into variables in the
binary via ldflags, so that the binary can report its own build info.

The following variable will be set:

```
	BuildVersion
	BuildTime
	BuildHash
	BuildDeps
```

See the bootstrap package and [pkg/bootstrap/build_info.go](../../../pkg/bootstrap/build_info.go) on how to access these 
variables in your binary.

```make clean```

Cleans the `dist` directory.

## PR Verification

PR verification needs to check out the PR branch and ensure it builds and tests successfully. Same as on a development environment,
```make init``` needs to be executed to bootstrap `lanai-cli` and `GNU Make`. Usually we don't expect `go-lanai` to be
checked out in this environment. So ```make init``` will use the version of `go-lanai` in the `go.mod` file. If you need
a specific version of the `lanai-cli`, you can use the `CLI_TAG` argument.

The ```make verify``` target is designed to facilitate the PR verification process. 
This target takes `PRIVATE_MODS` as an argument. You can use this argument to specify the private module versions you want to verify this PR with.
i.e. `make verify PRIVATE_MODS=github.com/<org>/<my-private-dependency>@develop`. If your project does not use private modules,
or your private module is not being developed in tandem (i.e. it has a fixed version in `go.mod` instead of a `replace` directive),
you can omit this argument. 

This command executes the following steps.

1. Run `make pre` to create a git tag `tmp/pr-verify`
2. Run `make update-dependency` to update module dependency based on `PRIVATE_MODS`. This is done by invoking `lanai-cli deps --modules=$(PRIVATE_MODS)` This 
will modify the version of the private modules in `go.mod` according to the value of `PRIVATE_MODS`. If the `UPDATE_DEPS_MARK` argument is
provided, a git `tag` will be created to save a record of this change.
3. Run `make drop-replace` to drop the `replace` directive based on `PRIVATE_MODS`. This is done by invoking `lanai-cli drop-replace --modules=$(PRIVATE_MODS)`
This will drop the `replace` directives for the `PRIVATE_MODS`. If the `DROP_REPLACE_MARK` argument is passed, a git tag will be created to save
a record of this change.
4. Run `make clean`.  
5. Run `make test`. 
6. Run `make lint`.
7. Run `make build`. The `PRIVATE_MODS` argument will pass through to this target.
8. Run `make report`. This will generate the test report
9. RUN `make post` to revert the code to the git tag `tmp/pre-verify`  

Note in step 2 `lanai-cli` will drop invalid `replace` directive before update dependency. It will run `go mod tidy` after
updating dependency, and add the `replace` directive back. All the tags created are local and not pushed upstream.

## Build For Distribution

### Build

Similar to PR verification, making a distribution build involves checking out a copy of the project and build it with 
specific version of the private modules. In addition, source code needs to be tagged so that there's a record of the
source code for this build.

The ```make dist``` target is designed for this purpose. This command also takes the `PRIVATE_MODS` argument. 
i.e. `make dist PRIVATE_MODS=github.com/<org>/<my-private-dependency>@develop VERSION=4.0.0-20`

This command executes the following steps:

1. Run `make pre` to create a git tag `tmp/pre-dist`
2. Run `make update-dependency` to update module dependency based on `PRIVATE_MODS`. This is done by invoking `lanai-cli deps --modules=$(PRIVATE_MODS)` This
   will modify the version of the private modules in `go.mod` according to the value of `PRIVATE_MODS`. A git tag will be created to save a record of this change.
3. Run `make drop-replace` to drop the `replace` directive based on `PRIVATE_MODS`. This is done by invoking `lanai-cli drop-replace --modules=$(PRIVATE_MODS)`
   This will drop the `replace` directives for the `PRIVATE_MODS`. A git tag will be created to save
   a record of this change.
4. Run `make clean`.
5. Run `make test`.
6. Run `make pre-build-docker` to make sure the code is at the git tag created by step 3. 
7. Run `make build-docker`. This command trigger the `docker` command to build a docker image. The image is built using
the `Dockerfile`. The `Dockerfile` uses a builder pattern. Source code will be copied into the builder and 
the `make build` command will be run inside the builder to produce the desired distribution package. See [Dockerfile.tmpl](Dockerfile.tmpl) 
for more details.
8. Run `make post-dist` to revert the code to the git tab `tmp/pre-dist`. Create a release tag `v$(VERSION)` using the tag
created in step 3. Optionally merge back to the source branch if the `SRC_BRANCH` argument is provided.

All the tags created are local and not pushed upstream.

### Publish

Use the ```make publish``` command to publish the image to docker hub. This should be run after `make dist`.
For example, after `make dist PRIVATE_MODS=github.com/<org>/<my-private-dependency>@develop VERSION=4.0.0-20)` 
run `make publish VERSION=4.0.0-20`. This commands executes the following steps:

1. Run `make push-docker` to push the image to docker registry. The required args are `DOCKER_TAG` and `DOCKER_REPO`. `DOCKER_TAG` will default to `$(VERSION)`
`DOCKER_REPO` will default to `registry-1.docker.io`. 
2. Run `git-push-tag` to push the release tag created by `make dist`.