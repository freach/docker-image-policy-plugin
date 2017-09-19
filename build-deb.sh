#!/bin/sh

# Install requirements
apt update
apt install -y ruby-dev build-essential
gem install fpm

# Build plugin
make clean
go get github.com/Sirupsen/logrus
go get github.com/docker/go-plugins-helpers/authorization
go get github.com/docker/docker/api
go get github.com/docker/engine-api/client
make

mkdir -p _tmp/usr/bin _tmp/lib/systemd/system _tmp/etc/docker
cp docker-image-policy _tmp/usr/bin
cp docker-image-policy.service _tmp/lib/systemd/system
chmod 755 _tmp/usr/bin/docker-image-policy
cat <<EOF > _tmp/etc/docker/docker-image-policy.json
{
  "whitelist": [],
  "blacklist": [],
  "defaultAllow": false
}
EOF

cat <<EOF > post-install.sh
#!/bin/sh
systemctl enable docker-image-policy
systemctl start docker-image-policy
EOF

cat <<EOF > post-remove.sh
#!/bin/sh
systemctl stop docker-image-policy
systemctl disable docker-image-policy
EOF


chmod 755 post-install.sh post-remove.sh

fpm \
  -s dir\
  -t deb\
  --description "Docker authentication plugin to enforce a image pull policy"\
  --url https://github.com/freach/docker-image-policy-plugin\
  --license "Apache License, Version 2.0"\
  -m "Simon Pirschel <simon@aboutsimon.com>"\
  -v "$(cat VERSION)"\
  -n docker-image-policy-plugin\
  --category admin\
  --after-install ./post-install.sh\
  --before-remove ./post-remove.sh\
  -C _tmp\
  -p "$(pwd)"

rm -rf _tmp post-install.sh post-remove.sh
