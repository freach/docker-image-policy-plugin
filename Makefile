prefix = /usr/local
VERSION = $(shell cat VERSION)

all: build

build:
	go build -o docker-image-policy -ldflags "-X main.version=$(VERSION)" .

install: build
	install -D docker-image-policy $(DESTDIR)$(prefix)/bin/docker-image-policy

clean:
	-rm -f docker-image-policy
	go clean

distclean: clean

uninstall:
	-rm -f $(DESTDIR)$(prefix)/bin/docker-image-policy

.PHONY: all install clean distclean uninstall
