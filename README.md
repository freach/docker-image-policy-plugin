# Docker Image policy plugin

`docker-image-policy` is a Docker *Access authorization plugin* written Go to control
which Images are allowed to be pulled by your Docker daemon. The plugin is using
the [AuthZPlugin API](https://docs.docker.com/engine/extend/plugins_authorization/)
by Docker. Black and Whitelistings are expressed through regular expression.
A default policy if no listing matched can be defined also.

**Supported: Docker Engine >= 1.11**

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

## Get started

### Plugin configuration

Add a config file (default: /etc/docker/docker-image-policy.json), and configure the plugin like so:

```json
{
  "whitelist": [
    "^alpine:",
    "^docker\\.elastic\\.co/beats/filebeat:",
    "^gcr\\.io/google_containers",
    "^mysql:",
    "^nginx:",
    "^php:",
    "^apache:",
    "^quay\\.io/calico/cni",
    "^quay\\.io/calico/node",
    "^quay\\.io/coreos/flannel"
  ],
  "blacklist": [
    "^docker:"
  ],
  "defaultAllow": false
}
```
The *whitelist* and *blacklist* array expect strings in regex format. Image pull requests will be checked by applying the compiled regular expressions on the full image, *<repository>:<tag>*.
**Certain characters in a regular expression like "." have special meaning and need to be escaped. The JSON format requires you to double escape**.

Image pull request will be handled in the following order:

1. Whitelist: Allow explicitly white listed images
1. Blacklist: Reject explicitly black listed images
1. defaultAllow: If true allow, if false reject

If one of the steps matched, the plugin will return accordingly.

### Docker configuration

Edit your `/etc/docker/daemon.json`

```
{
  "authorization-plugins": ["docker-image-policy"]
}
```

### Running

Start `docker-image-policy` and restart Docker daemon.

```sh
$ docker-image-policy &
$ curl localhost:5006/health
$ service docker restart
```

**Please consider using the systemd service file for running docker-image-policy**

## API Endpoints

* `/health` -> Health check
* `/config` -> Current config
* `/version` -> Current version

```
$ curl localhost:5006/health
HEALTHY
```

## Author

* Simon Pirschel
