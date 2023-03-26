SHELL:=/usr/bin/env sh

OSFLAG :=
UNAME := $(shell uname -m)
ifeq ($(UNAME), arm64)
  OSFLAG = DOCKER_DEFAULT_PLATFORM=linux/amd64
endif

# Build & start simple validator cluster
.PHONY:build
build:
	$(OSFLAG) docker compose -f docker-compose.bsc.yml build
	$(OSFLAG) docker compose -f docker-compose.simple.yml build

# run only bootstrap
.PHONY:bootstrap
bootstrap:
	$(OSFLAG) docker compose -f docker-compose.simple.yml run --rm bootstrap-simple

# run bsc-rpc, bsc-validator1
.PHONY:network
network:
	$(OSFLAG) docker compose -f docker-compose.simple.yml up -d bsc-rpc bsc-validator1

.PHONY:network-down
network-down:
	$(OSFLAG) docker compose -f docker-compose.simple.yml down -v

# run bootstrap and bsc-rpc, bsc-validator1
.PHONY:network-all
network-all:
	$(OSFLAG) docker compose -f docker-compose.simple.yml up -d

.PHONY:reset-network
reset-network: network-down network-all
#reset-network: network-down bootstrap network

#------------------------------------------------------------------------------
# Util
#------------------------------------------------------------------------------
.PHONY:shfmt
shfmt:
	shfmt -l -w ./scripts/*.sh

.PHONY:shchk
shchk:
	shellcheck ./scripts/*.sh
