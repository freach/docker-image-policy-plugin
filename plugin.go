package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	dockerapi "github.com/docker/docker/api"
	dockerclient "github.com/docker/engine-api/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

func newPlugin(dockerHost, certPath string, tlsVerify bool) (*authPlugin, error) {
	var httpClient *http.Client
	if certPath != "" {
		tlsc := &tls.Config{}

		cert, err := tls.LoadX509KeyPair(
			filepath.Join(certPath, "cert.pem"),
			filepath.Join(certPath, "key.pem"),
		)

		if err != nil {
			return nil, fmt.Errorf("Error loading x509 key pair: %s", err)
		}

		tlsc.Certificates = append(tlsc.Certificates, cert)
		tlsc.InsecureSkipVerify = !tlsVerify
		transport := &http.Transport{TLSClientConfig: tlsc}
		httpClient = &http.Client{Transport: transport}
	}

	client, err := dockerclient.NewClient(
		dockerHost, dockerapi.DefaultVersion, httpClient, nil,
	)

	if err != nil {
		return nil, err
	}

	return &authPlugin{client: client}, nil
}

type authPlugin struct {
	client *dockerclient.Client
}

func (p *authPlugin) AuthZReq(req authorization.Request) authorization.Response {
	if req.RequestMethod == "POST" && strings.Contains(req.RequestURI, "/images/create") {
		uri, err := url.ParseRequestURI(req.RequestURI)
		if err != nil {
			errMsg := fmt.Sprintf("Error while parsing request URI: %s", err)
			logrus.Error(errMsg)
			return authorization.Response{Allow: false, Msg: errMsg}
		}

		query, err := url.ParseQuery(uri.RawQuery)
		if err != nil {
			errMsg := fmt.Sprintf("Error while parsing request query string: %s", err)
			logrus.Error(errMsg)
			return authorization.Response{Allow: false, Msg: errMsg}
		}

		var image string
		if _, exists := query["tag"]; exists {
			image = fmt.Sprintf("%s:%s", query["fromImage"][0], query["tag"][0])
		} else {
			// Docker < 17.xx
			image = query["fromImage"][0]
		}

		bImage := []byte(image)

		for _, v := range reWhitelist {
			if v.Match(bImage) {
				return authorization.Response{Allow: true}
			}
		}

		for _, v := range reBlacklist {
			if v.Match(bImage) {
				logrus.Infof("Image %s blocked, because blacklisted.", image)
				return authorization.Response{
					Allow: false,
					Msg: fmt.Sprintf("Image %s is blacklisted on this server.", image),
				}
			}
		}

		if configuration.DefaultAllow == true {
			return authorization.Response{Allow: true}
		}

		logrus.Infof("Image %s blocked, because default is reject.", image)
		return authorization.Response{
			Allow: false,
			Msg: fmt.Sprintf("Image %s is not allowed on this server.", image),
		}
	}

	return authorization.Response{Allow: true}
}

func (p *authPlugin) AuthZRes(req authorization.Request) authorization.Response {
	return authorization.Response{Allow: true}
}
