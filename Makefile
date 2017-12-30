GO := $(shell which go)

test:
	sudo $(GO) test -v ./...
	sudo rm -f ./pmon/core
