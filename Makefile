prefix = /usr/local

all: docker-image-policy

docker-image-policy:
	go build -o docker-image-policy .

install: docker-image-policy
	install -D docker-image-policy $(DESTDIR)$(prefix)/bin/docker-image-policy

clean:
	-rm -f docker-image-policy
	go clean

distclean: clean

uninstall:
	-rm -f $(DESTDIR)$(prefix)/bin/docker-image-policy

.PHONY: all install clean distclean uninstall
