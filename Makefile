target=./cmd/nocolor

build:
	go build $(target)

install:
	go install $(target)

vet:
	go vet ./...

.PHONY: vet build install
