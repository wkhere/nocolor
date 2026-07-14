target=./cmd/nocolor

build:
	go test .
	go build $(target)

install:
	go install -ldflags=-s $(target)

vet:
	go vet ./...

fuzz:
	go test -fuzz=.

sel=.
cnt=5
bench:
	go test -bench=$(sel) -benchmem -count=$(cnt)


.PHONY: vet build install fuzz bench
