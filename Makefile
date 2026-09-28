BINARY := updatego
BUILDDIR := builds

# Static binaries (no cgo).
export CGO_ENABLED := 0
LDFLAGS := -s -w

.PHONY: default build all clean

default: build

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

all: $(BUILDDIR)/updatego-linux-amd64 \
	$(BUILDDIR)/updatego-windows-amd64.exe \
	$(BUILDDIR)/updatego-darwin-arm64

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

$(BUILDDIR)/updatego-linux-amd64: $(BUILDDIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $@ .

$(BUILDDIR)/updatego-windows-amd64.exe: $(BUILDDIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $@ .

$(BUILDDIR)/updatego-darwin-arm64: $(BUILDDIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $@ .

clean:
	rm -f $(BINARY)
	rm -rf $(BUILDDIR)
