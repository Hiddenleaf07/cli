package registry // import "github.com/docker/docker/registry"

import (
	"net/url"
	"strings"

	"github.com/docker/go-connections/tlsconfig"
)

func (s *Service) lookupV2Endpoints(hostname string) (endpoints []APIEndpoint, err error) {
	ana := s.config.allowNondistributableArtifacts(hostname)

	if hostname == DefaultNamespace || hostname == IndexHostname {
		for _, mirror := range s.config.Mirrors {
			if !strings.HasPrefix(mirror, "http://") && !strings.HasPrefix(mirror, "https://") {
				mirror = "https://" + mirror
			}
			mirrorURL, err := url.Parse(mirror)
			if err != nil {
				return nil, invalidParam(err)
			}
			mirrorTLSConfig, err := newTLSConfig(mirrorURL.Host, s.config.isSecureIndex(mirrorURL.Host))
			if err != nil {
				return nil, err
			}
			endpoints = append(endpoints, APIEndpoint{
				URL:       mirrorURL,
				Mirror:    true,
				TLSConfig: mirrorTLSConfig,
			})
		}
		endpoints = append(endpoints, APIEndpoint{
			URL:       DefaultV2Registry,
			Official:  true,
			TLSConfig: tlsconfig.ServerDefault(),

			AllowNondistributableArtifacts: ana,
		})

		return endpoints, nil
	}

	tlsConfig, err := newTLSConfig(hostname, s.config.isSecureIndex(hostname))
	if err != nil {
		return nil, err
	}

	endpoints = []APIEndpoint{
		{
			URL: &url.URL{
				Scheme: "https",
				Host:   hostname,
			},
			TLSConfig: tlsConfig,

			AllowNondistributableArtifacts: ana,
		},
	}

	if tlsConfig.InsecureSkipVerify {
		endpoints = append(endpoints, APIEndpoint{
			URL: &url.URL{
				Scheme: "http",
				Host:   hostname,
			},
			// used to check if supposed to be secure via InsecureSkipVerify
			TLSConfig: tlsConfig,

			AllowNondistributableArtifacts: ana,
		})
	}

	return endpoints, nil
}
