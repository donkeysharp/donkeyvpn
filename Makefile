.PHONY: build check test

build:
	bash deploy/scripts/install.sh build

check:
	bash deploy/scripts/install.sh check

test:
	go test ./...
