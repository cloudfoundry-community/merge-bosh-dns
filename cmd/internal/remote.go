package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"runtime"

	"github.com/cloudfoundry-community/merge-bosh-dns/config"
	"github.com/cloudfoundry-community/merge-bosh-dns/merge"
)

func caPool(cfg *config.Config) (*x509.CertPool, error) {
	var caPool *x509.CertPool
	if cfg.RemoteSource.CACerts != "" {
		caPool = x509.NewCertPool()
		foundCerts := caPool.AppendCertsFromPEM([]byte(cfg.RemoteSource.CACerts))
		if !foundCerts {
			return nil, fmt.Errorf("CACerts were specified for remote, but could not parse any certificates")
		}
	} else {
		var err error
		caPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("Could not parse system cert pool: %s", err)
		}
	}

	return caPool, nil
}

func RecordsClientFromConfig(cfg *config.Config) (*merge.RecordsClient, error) {
	caPool, err := caPool(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.RemoteSource.Endpoint == "" {
		return nil, fmt.Errorf("No Endpoint for the remote source was given")
	}

	ret := &merge.RecordsClient{
		URL: cfg.RemoteSource.Endpoint,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.RemoteSource.InsecureSkipVerify,
					RootCAs:            caPool,
				},
				MaxIdleConnsPerHost: runtime.NumCPU(),
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 10 {
					return fmt.Errorf("Stopped after 10 redirects")
				}

				return nil
			},
		},
	}

	if cfg.RemoteSource.BasicAuth.Enabled {
		ret.BasicAuth = &merge.BasicAuthCredentials{
			Username: cfg.RemoteSource.BasicAuth.Username,
			Password: cfg.RemoteSource.BasicAuth.Password,
		}
	}

	return ret, nil
}
