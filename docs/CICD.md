# Prerequisites for CI/CD Worker VM

- GO 1.16+
- Git 2.23.0+
- GNU Make 3.81+
- Docker 20.10.5+
- Nodejs 15.14.0+ & NPM 7.9.0+ (for swagger UI and login UI)
- [Optional] Maven 3.6.0+ (for webjars dependencies)
- Access to any required private Repo via `git@` and `id_rsa` is properly set
- GO environment variables such as `$GOPATH` and `$GOBIN` are properly set

<br>

# Tooling

`lanai-cli` is a command line tool to help with common build/code generation/git tasks.

### To install via GNU Make:

```shell
make init-cli [CLI_TAG=git_brand_or_tag]
```

Where the `CLI_TAG` is supported and optional for Services. 
It typically should match the branch/version of `cto-github.cisco.com/NFV-BU/go-lanai`

<br>

# CI/CD Operations

CI/CD operations are done via GNU Make with help of `lanai-cli` and `Makefile-Generated`

## General `make` Variables

Regardless CI/CD scenarios, CI/CD operators typically need to provide following variables to `make` command:

### `VERSION`

The [GO Semantic Version](https://golang.org/doc/modules/version-numbers) **without** leading "v". 

Example: `VERSION=4.0.0-40`, `VERSION=4.0.0-RC.1`

Default Value: `0.0.0-SNAPSHOT`

> This variable is REQUIRED in all CI/CD targets unless specified explicitly

### `PRIVATE_MODS`

Private Modules' path and branch/version which the current project depends on, in format of `private.url/path/to/module@branch_or_tag`.
Multiple modules are separated by ","

> "Private Module" means extra modules that are developed/released by same team and are under rapid development.

Example: `PRIVATE_MODS=cto-github.cisco.com/NFV-BU/go-lanai@develop` or `PRIVATE_MODS=cto-github.cisco.com/NFV-BU/go-lanai@v4.0.0-50`

Default Value: `cto-github.cisco.com/NFV-BU/go-lanai@develop` for Services, empty string for GO-Lanai libraries

> This variable is REQUIRED in many CI/CD targets unless specified explicitly

### `SRC_BRANCH`

In case `go.mod` or other changes (e.g. generated source code) need to be merged back to develop branch, this is required. 

Example: `SRC_BRANCH=develop`

Default Value: empty string

> By not providing this variable, "merging back" step is skipped in all applicable scenarios 

> Note: More information could be found in generated `Makefile-Generated` file

## Temporary Changes and Local Git Tags

During the process of CI/CD operations, `make` targets would modify `go.mod` files or generate additional source code. 
To track those changes, `make` targets would create some local commits and tags as savepoints. Those savepoints/tags 
is also known as `Marks`. 

`lanai-cli git` sub-commands provide some useful operations around `Marks`. 

During the process, all `make` targets guarantee following behaviors:

1. Git worktree never switch branch/tag
2. Git `HEAD` doesn't change. i.e. it always points to the commit at which the process is started
   
If the process is successful, all `make` targets also guarantee following:

1. All changes during the process are unstaged 

> **Note**: the Git local Worktree should be discarded after use. 

<br>

# CI/CD Scenarios for GO-Lanai Libraries

Applicable to `cto-github.cisco.com/NFV-BU/go-lanai`

## PR Verify

Command:

```shell
make init
make verify
```

Variables: `VERSION` is not required and `PRIVATE_MODS` is not typically required 

Do:

- When `PRIVATE_MODS` is provided, `go.mod` is updated accordingly
- run `generate`, `test` and `build` on packages
- coverage report saved to `dist`
- code analysis

## Nightly Build or Build from Branch

Command:

```shell
make init
make dist VERSION=4.0.0-20
make publish VERSION=4.0.0-20
```

Variables:

- `PRIVATE_MODS` not typically required
- `SRC_BRANCH` specify if merging changes back to branch is required. Typically, not required for Nightly Build

Do:

- When `PRIVATE_MODS` is provided, `go.mod` is updated accordingly
- run `generate`, `test` and `build` on packages
- ready-to-distribute source code is tagged as `v$VERSION` in Git
- the git tag is pushed to remote

## Promote Tagged Source Code

This operation is typically performed on previously released Git Tag

Command:

```shell
make init
make redist VERSION=4.0.0
make publish VERSION=4.0.0
```

Variables: `PRIVATE_MODS` not required

Do:

- All files are used as-is
- NO `generate`, `test` or `build`
- current Git worktree is re-tagged as `v$VERSION` in Git
- the git tag is pushed to remote

<br>

# CI/CD Scenarios for GO-Lanai Services

During development cycle, service depends on go-lanai project using branch name.
For example, service is being developed in develop (or feature) branch, and go-lanai is being developed in develop (or feature) branch as well.
Service may have a "replace" directive that points to a local checked out version of Go-Lanai to help facilitate develop.

During most CI/CD scenarios, the `make` target would attempt to "fix" `go.mod` and `go.sum` by dropping `replace` directive
and update corresponding dependency versions. 

## PR Verify

Command:

```shell
make init CLI_TAG=feature/branch
make verify PRIVATE_MODS=cto-github.cisco.com/NFV-BU/go-lanai@feature/branch
```

Variables: 

- `CLI_TAG` required and should match the `PRIVATE_MODS`
- `VERSION` is not required  
- `PRIVATE_MODS` is required if the PR only works on particular branch of GO-Lanai

Do:

- `go.mod`'s `require` section is updated according to `PRIVATE_MODS`. 
  Result marked as `tmp/0.0.0-SNAPSHOT/update-deps`
- `go.mod`'s `replace` is dropped. Result marked as `tmp/0.0.0-SNAPSHOT/drop-replace`
- run `generate`, `test` and `build` on packages
- coverage report saved to `dist`
- code analysis

## Nightly Build or Build from Branch

Command:

```shell
make init CLI_TAG=4.0.0-90
make dist VERSION=4.0.0-20 PRIVATE_MODS=cto-github.cisco.com/NFV-BU/go-lanai@v4.0.0-90 #SRC_BRANCH=any_branch
make publish VERSION=4.0.0-20 DOCKER_REPO=dockerhub.cisco.com/vms-platform-dev-docker
```

Variables:

- `CLI_TAG` required and should match the `PRIVATE_MODS`
- `DOCKER_REPO` required for publish
- `SRC_BRANCH` specify if merging changes back to branch is required. Typically, not required for Nightly Build

Do:

- `go.mod`'s `require` section is updated according to `PRIVATE_MODS`.
  Result marked as `tmp/$VERSION/update-deps`
- `go.mod`'s `replace` is dropped. Result marked as `tmp/$VERSION/drop-replace`
- run `generate` and `test` on packages
- build Docker image and tag as `$VERSION`
- ready-to-distribute source code (marked as `tmp/$VERSION/drop-replace`) is tagged as `v$VERSION` in Git
- the git tag is pushed to remote (publish)
- the Docker image is pushed to Docker repository (publish)

## Promote Tagged Source Code

This operation is typically performed on previously released Git Tag.

Because build-info or manifests are built into executables, when promoting one version to another, 
the executables requires re-build.

Command:

```shell
make init
make redist VERSION=4.0.0
make publish VERSION=4.0.0 DOCKER_REPO=dockerhub.cisco.com/vms-platform-dev-docker
```

Variables: 
- `PRIVATE_MODS` not required
- `DOCKER_REPO` required for publish

Do:

- All files are used as-is
- NO `generate` or `test`
- Docker image is rebuilt and tagged as `$VERSION`
- current Git worktree is re-tagged as `v$VERSION` in Git
- the git tag is pushed to remote (publish)
- the Docker image is pushed to Docker repository (publish)

<br>

# Appendix: Original Design Notes (Please Ignore)

When Service is being distributed/built, the general CICD process is below:

1. Create a local temporary branch for build. 

2. Update go.mod file so that the dependency on Go-Lanai points to the latest from the matching Go-Lanai branch or a given Go-Lanai tag.
 
3. Create a commit (commit A). This commit has go.mod updated. But this go.mod may still have the replace directive. This commit can be used
if we want to merge the update back to the working branch.

4. If there's a "replace" directive for go-lanai, drop it from go.mod.

5. Create a commit (commit B). This commit has the go.mod that will be used exactly for build. This commit is going to be used to tag the build.
 
6. Generate and test.

7. Build docker image.

8. If step 6 and 7 are successful. Tag commit B for this build and push the tag.

9. Optionally merge commit A back into main branch.