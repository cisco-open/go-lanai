# Tooling

`go-lanai/cmd/lanai-cli` is a command line tool to help with common build/code generation/git tasks.

### Prerequisites for GO Lanai Service

- `Makefile` available 

### To install via GNU Make:

```shell
make init-cli [CLI_TAG=git_brand_or_tag]
```

Where the `CLI_TAG` is optional and supported at Service using 

# CI/CD Operations for Go-Lanai


# CI/CD Operations for Go-Lanai Service

During development cycle, service depends on go-lanai project using branch name.
For example, service is being developed in develop (or feature) branch, and go-lanai is being developed in develop (or feature) branch as well.
Service may have a "replace" directive that points to a local checked out version of Go-Lanai to help facilitate develop.

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