target=./cmd/nocolor

build:
	go test .
	go build $(target)

install:
	go install $(target)

vet:
	go vet ./...

fuzz:
	go test -fuzz=.

sel=.
cnt=5
bench:
	go test -bench=$(sel) -benchmem -count=$(cnt)


.PHONY: vet build install fuzz bench
