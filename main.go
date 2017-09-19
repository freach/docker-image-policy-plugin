package main

import (
	"flag"
	"net/http"
	"os"
	"regexp"
	"fmt"
	"encoding/json"
	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
	pluginSocket      = "/run/docker/plugins/docker-image-policy.sock"
	defaultConfig  = "/etc/docker/docker-image-policy.json"
	defaultAddr       = "127.0.0.1:5006"
)

type Config struct {
	Whitelist []string `json:"whitelist"`
	Blacklist []string `json:"blacklist"`
	DefaultAllow bool `json:"defaultAllow"`
}

// Globals
var (
	version string
	reWhitelist []*regexp.Regexp
	reBlacklist []*regexp.Regexp
	configuration Config
)

// Command line options
var (
	flDockerHost = flag.String("host", defaultDockerHost, "Docker daemon host")
	flCertPath   = flag.String("cert-path", "", "Path to Docker certificates (cert.pem, key.pem)")
	flTLSVerify  = flag.Bool("tls-verify", false, "Verify certificates")
	flDebug      = flag.Bool("debug", false, "Enable debug logging")
	flVersion    = flag.Bool("version", false, "Print version")
	flAddr       = flag.String("addr", defaultAddr, "Plugin API [HOSTNAME:PORT]")
	flConfig     = flag.String("config", defaultConfig, "Path to plugin config file")
)

func readConfig(configFile string) error {
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode JSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configuration); err != nil {
		return err
	}

	// Build whitelist
	for _, v := range configuration.Whitelist {
		re, err := regexp.Compile(v)
		if err != nil {
			return err
		}
		reWhitelist = append(reWhitelist, re)
	}

	// Build blacklist
	for _, v := range configuration.Blacklist {
		re, err := regexp.Compile(v)
		if err != nil {
			return err
		}
		reBlacklist = append(reBlacklist, re)
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HEALTHY")
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, version)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(configuration)
	if err != nil {
		fmt.Fprint(w, err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "    ")
	out.WriteTo(w)
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	flag.Parse()

	// Print version and exit
	if *flVersion {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	if *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Docker Image policy plugin started (version: %s)", version)

	if err := readConfig(*flConfig); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("%d entries in whitelist.", len(reWhitelist))
	logrus.Infof("%d entries in blacklist.", len(reBlacklist))
	logrus.Infof("Default allow: %t", configuration.DefaultAllow)

	// Add additional handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/version", versionHandler)

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
