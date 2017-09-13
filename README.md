# Docker Image policy plugin

`docker-image-policy` is a Docker *Access authorization plugin* written Go to control
which Images are allowed to be pulled by your Docker daemon. The plugin is using
the [AuthZPlugin API](https://docs.docker.com/engine/extend/plugins_authorization/)
by Docker. Whitelistings are expressed through Regular expression.

## Building

To build this plugin Go >= 1.7 and proper GOPATH setup is required.

```sh
$ make
```

## Build Debian Package

**PACKAGES WILL BE INSTALLED**

*Please consider using a Docker container for building the Debian package.*

```sh
$ sudo sh build-deb.sh
```

Example with Docker container:

```sh
$ git clone git@github.com:freach/docker-image-policy-plugin.git ~/docker-image-policy-plugin
$ docker run -it --rm -v ~/docker-image-policy-plugin:/go golang bash
$ sh build-deb.sh
$ ls *.deb
```

## Running

Edit your `/etc/docker/daemon.json`

```
{
  "authorization-plugins": ["docker-image-policy"]
}
```

Add a whitelist config file (default: */etc/docker/image-whitelist*), one
whitelist regex entry per line, like so:

```
^alpine:
^docker\.elastic\.co\/beats\/filebeat:
^gcr\.io\/google_containers
^mysql:
^nginx:
^php:
^apache:
^quay\.io\/calico\/cni
^quay\.io\/calico\/node
^quay\.io\/coreos\/flannel
```

Start `docker-image-policy` and restart Docker daemon.

```sh
$ docker-image-policy &
$ curl localhost:8080/health
$ service docker restart
```

**Please consider using the systemd service file for running docker-image-policy**

## Author

* Simon Pirschel
