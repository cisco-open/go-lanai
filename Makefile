### Global Variables
WORK_DIR = $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
TMP_DIR = $(WORK_DIR).tmp

# patterns
null  =
space = $(null) #
comma = ,

### Main
.PHONY: init-once init-cli init print

default: init-cli

-include Makefile-Generated

## Required Variables from command-line

## Required Variables by Local Targets
GO ?= go
GIT ?= git
CLI ?= lanai-cli

INIT_TMP_DIR = $(TMP_DIR)/init

## Local Targets

# init-once:
#	Used to setup local dev environment, and should be used only once per environment.
# 	This target assumes local environment has proper access to $PRIVATE_REPOS
#	Required Vars:
#		- PRIVATE_REPOS space delimited private git servers. e.g. PRIVATE_REPOS="private-github.my-domain.org"
ifneq ($(strip $(value PRIVATE_REPOS)),)
init-once:
	$(GO) env -w GOPRIVATE="$(subst $(space),$(comma),$(strip $(value PRIVATE_REPOS)))"
	$(foreach repo,$(PRIVATE_REPOS),\
		$(GIT) config --global url."ssh://git@$(repo)/".insteadOf "https://$(repo)/";\
	)
else
init-once:
	@echo "Nothing to be done. Did you forget to set PRIVATE_REPOS?"
endif

# init-cli:
#	Used to bootstrap any targets other than init-once
init-cli: CLI_PKG = ./cmd/lanai-cli
init-cli: CLI_VERSION = 0.0.0-$(shell date +"%Y%m%d%H%M%S")-$(shell $(GIT) rev-parse --short=12 HEAD || echo "SNAPSHOT")
init-cli: print
	@echo Installing $(CLI_PKG)@$(CLI_VERSION) ...
	@$(GO) install -ldflags="-X main.BuildVersion=$(CLI_VERSION)" $(CLI_PKG)

# init:
#	Used to bootstrap any targets other than init-once and init-cli
init: init-cli
	$(CLI) init libs -o ./

print:
	@echo "CLI_VERSION:  $(CLI_VERSION)"


