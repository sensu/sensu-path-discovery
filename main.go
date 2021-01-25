package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/utahta/go-openuri"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	SubPrefix          string
	PathsFiles         []string
	InsecureSkipVerify bool
	TrustedCAFile      string
}

type Paths struct {
	Path string   `json:"path"`
	Subs []string `json:"subs"`
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-path-discovery",
			Short:    "Discover file system paths and output a list of agent subscriptions.",
			Keyspace: "sensu.io/plugins/sensu-path-discovery/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "subscription-prefix",
			Env:       "SUBSCRIPTION_PREFIX",
			Argument:  "subscription-prefix",
			Shorthand: "p",
			Default:   "",
			Usage:     "The agent subscription name prefix",
			Value:     &plugin.SubPrefix,
		},
		{
			Path:      "paths-file",
			Env:       "PATHS_FILE",
			Argument:  "paths-file",
			Shorthand: "f",
			Default:   []string{},
			Usage:     "The file location(s) for the mapping file (file path(s) or URL(s))",
			Value:     &plugin.PathsFiles,
		},
		{
			Path:      "insecure-skip-verify",
			Env:       "",
			Argument:  "insecure-skip-verify",
			Shorthand: "i",
			Default:   false,
			Usage:     "Skip TLS certificate verification (not recommended!)",
			Value:     &plugin.InsecureSkipVerify,
		},
		{
			Path:      "trusted-ca-file",
			Env:       "",
			Argument:  "trusted-ca-file",
			Shorthand: "t",
			Default:   "",
			Usage:     "TLS CA certificate bundle in PEM format",
			Value:     &plugin.TrustedCAFile,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	if len(plugin.PathsFiles) == 0 {
		return sensu.CheckStateCritical, fmt.Errorf("--paths-file is required")
	}
	return sensu.CheckStateOK, nil
}

func processSubs() ([]string, error) {
	subs := []string{}
	subsSet := make(map[string]bool)

	for _, path := range plugin.PathsFiles {
		pathsBytes, err := readPath(string(path))
		if err != nil {
			return []string{}, err
		}
		pathsList := []Paths{}
		err = json.Unmarshal(pathsBytes, &pathsList)
		if err != nil {
			return []string{}, fmt.Errorf("Failed to unmarshal JSON from paths file %s: %v", path, err)
		}

		for _, v := range pathsList {
			if _, err := os.Stat(v.Path); !os.IsNotExist(err) {
				for _, s := range v.Subs {
					if _, e := subsSet[s]; !e {
						subs = append(subs, plugin.SubPrefix+s)
						subsSet[s] = true
					}
				}
			}
		}
	}

	return subs, nil
}

func readPath(path string) ([]byte, error) {
	var paths io.ReadCloser
	var err error

	// If provided SSL options, setup the client manually
	if plugin.InsecureSkipVerify || len(plugin.TrustedCAFile) > 0 {
		certs, err := loadCACerts(plugin.TrustedCAFile)
		if err != nil {
			return []byte{}, err
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: plugin.InsecureSkipVerify,
			RootCAs:            certs,
		}
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		client := &http.Client{
			Transport: transport,
		}
		paths, err = openuri.Open(path, openuri.WithHTTPClient(client))
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to open paths file %s: %v", path, err)
		}
	} else {
		paths, err = openuri.Open(path)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to open paths file %s: %v", path, err)
		}
	}

	defer paths.Close()

	pathsBytes, err := ioutil.ReadAll(paths)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to read from paths file %s: %v", path, err)
	}

	return pathsBytes, nil
}

func loadCACerts(path string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("Failed to get system cert pool: %v", err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if len(path) > 0 {
		certs, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("Failed to read CA file %s: %s", path, err)
		} else {
			rootCAs.AppendCertsFromPEM(certs)
		}
	}
	return rootCAs, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	subs, err := processSubs()

	if len(subs) > 0 {
		fmt.Println(strings.Join(subs, "\n"))
	}

	if err != nil {
		return sensu.CheckStateWarning, err
	}

	return sensu.CheckStateOK, nil
}
