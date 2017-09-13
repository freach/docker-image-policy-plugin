package main

import (
	"flag"
	"net/http"
	"bufio"
	"os"
	"regexp"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
	pluginSocket      = "/run/docker/plugins/docker-image-policy.sock"
	defaultWhitelist  = "/etc/docker/image-whitelist"
	defaultAddr       = "0.0.0.0:8080"
)

var (
	flDockerHost = flag.String("host", defaultDockerHost, "Docker daemon host")
	flCertPath   = flag.String("cert-path", "", "Path to Docker certificates (cert.pem, key.pem)")
	flTLSVerify  = flag.Bool("tls-verify", false, "Verify certificates")
	flDebug      = flag.Bool("debug", false, "Enable debug logging")
	flAddr       = flag.String("addr", defaultAddr, "Webserver [HOSTNAME:PORT]")
	flWhitelist  = flag.String("whitelist-path", defaultWhitelist, "Path to Docker Image whitelist file")
	reWhitelist []*regexp.Regexp
)

func readWhiteList(whitelist string) error {
	file, err := os.Open(whitelist)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(scanner.Bytes()) > 0 {
			re, err := regexp.Compile(scanner.Text())
			if err != nil {
				return err
			}
			reWhitelist = append(reWhitelist, re)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HEALTHY")
}

func main() {
	flag.Parse()

	if *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Info("Docker Image policy plugin started")

	if err := readWhiteList(*flWhitelist); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("%d entries in whitelist.", len(reWhitelist))

	http.HandleFunc("/health", healthHandler)

	go func() {
		logrus.Debugf("Server running on %s", *flAddr)
		if err := http.ListenAndServe(*flAddr, nil); err != nil {
			logrus.Fatal(err)
		}
	}()

	plugin, err := newPlugin(*flDockerHost, *flCertPath, *flTLSVerify)
	if err != nil {
		logrus.Fatal(err)
	}

	h := authorization.NewHandler(plugin)

	logrus.Debugf("Plugin running on %s", pluginSocket)
	if err := h.ServeUnix(pluginSocket, 0); err != nil {
		logrus.Fatal(err)
	}
}
