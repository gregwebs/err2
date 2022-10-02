# Original from github.com/pkg/errors

PKG1 := github.com/gregwebs/try/err3
PKG2 := github.com/gregwebs/try/assert
PKG3 := github.com/gregwebs/try/try
PKG4 := github.com/gregwebs/try/stackprint
PKGS := $(PKG1) $(PKG2) $(PKG3) $(PKG4)

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

GO := go
# GO := go1.18beta2

check: test vet gofmt

test1:
	$(GO) test $(PKG1)

test2:
	$(GO) test $(PKG2)

test3:
	$(GO) test $(PKG3)

test4:
	$(GO) test $(PKG4)

test:
	$(GO) test $(PKGS)

bench:
	$(GO) test -bench=. $(PKGS)

bench1:
	$(GO) test -bench=. $(PKG1)

bench2:
	$(GO) test -bench=. $(PKG2)

vet: | test
	$(GO) vet $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

godoc:
	@GO111MODULE=off godoc -http=0.0.0.0:6060

build:
	cat err3/err3.go | sed 's|package err3|package try|' > err3.go
