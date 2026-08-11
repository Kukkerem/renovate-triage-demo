# ====================================================================================
# Setup Project
PROJECT_NAME := renovate-triage-demo
PROJECT_REPO := github.com/kukkerem/$(PROJECT_NAME)

# Terraform Related variables — mirrors the real SAP/crossplane-provider-btp
# Makefile. The version below must move in lockstep with the generated
# CRDs/schema; nothing here enforces that automatically.
export TERRAFORM_PROVIDER_SOURCE ?= SAP/btp
export TERRAFORM_PROVIDER_REPO ?= https://github.com/SAP/terraform-provider-btp
export TERRAFORM_PROVIDER_VERSION ?= 1.16.1
export TERRAFORM_PROVIDER_DOWNLOAD_NAME ?= terraform-provider-btp

GO ?= go
GO_PROJECT := $(PROJECT_REPO)

.PHONY: build
build:
	$(GO) build ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: reviewable
reviewable: build vet test
	@echo "reviewable: build, vet and test all passed"

.PHONY: run
run: build
	$(GO) run ./cmd/provider

.PHONY: terraform.buildvars
terraform.buildvars:
	@echo TERRAFORM_PROVIDER_SOURCE=$(TERRAFORM_PROVIDER_SOURCE)
	@echo TERRAFORM_PROVIDER_REPO=$(TERRAFORM_PROVIDER_REPO)
	@echo TERRAFORM_PROVIDER_VERSION=$(TERRAFORM_PROVIDER_VERSION)
	@echo TERRAFORM_PROVIDER_DOWNLOAD_NAME=$(TERRAFORM_PROVIDER_DOWNLOAD_NAME)
