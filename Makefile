PLUGIN_NAME=terraform-provider-active24
PROVIDER_NS=petrskyva
PROVIDER_NAME=active24
VERSION?=0.0.1

OS?=$(shell uname | tr '[:upper:]' '[:lower:]')
ARCH?=$(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

.PHONY: build
build:
	go build -o bin/$(PLUGIN_NAME) .

.PHONY: install
install: build
	mkdir -p $$HOME/.terraform.d/plugins/registry.terraform.io/$(PROVIDER_NS)/$(PROVIDER_NAME)/$(VERSION)/$(OS)_$(ARCH)
	cp bin/$(PLUGIN_NAME) $$HOME/.terraform.d/plugins/registry.terraform.io/$(PROVIDER_NS)/$(PROVIDER_NAME)/$(VERSION)/$(OS)_$(ARCH)/

.PHONY: tidy
tidy:
	go mod tidy


