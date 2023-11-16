package proxy

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cleopatrio/proxy/logger"
	"github.com/sirupsen/logrus"
)

const (
	ExactPathType  PathType = "Exact"
	PrefixPathType PathType = "Prefix"
)

type PathType string

type ProxyPath struct {
	Path            string   `yaml:"path" example:"/files"`
	PathType        PathType `yaml:"pathType" example:"Exact"`
	PortNumber      int      `yaml:"portNumber" example:"3001"`
	TLS             bool     `yaml:"tls"`
	EnableReplay    bool     `yaml:"enableReplay"`
	EnableRateLimit bool     `yaml:"enableRateLimit"`
}

func (p *ProxyPath) RequestURL(requestHost, requestPath string) (requestURL *url.URL) {
	if portRegex := regexp.MustCompile(`\:\d+`); portRegex.Match([]byte(requestHost)) {
		// Use `path.Port` if it is present.
		requestHost = portRegex.ReplaceAllStringFunc(requestHost, func(s string) string {
			if p.PortNumber > 0 {
				return fmt.Sprintf(`:%d`, p.PortNumber)
			} else {
				return ""
			}
		})
	} else {
		if p.PortNumber > 0 {
			requestHost = requestHost + fmt.Sprintf(":%d", p.PortNumber)
		}
	}

	var scheme string = func() string {
		if p.TLS {
			return "https://"
		} else {
			return "http://"
		}
	}()

	// Overwrite host scheme based on TLS configuration.
	schemeRegex := regexp.MustCompile(`http[s]?\:\/\/`)
	requestHost = schemeRegex.ReplaceAllLiteralString(requestHost, "")

	switch p.PathType {
	case PrefixPathType:
		if !strings.HasPrefix(requestPath, p.Path) {
			logger.Logger.
				WithFields(logrus.Fields{"host": requestHost, "path": requestPath}).
				Warn("Mismatched route")

			return
		}

		requestURL, _ = url.Parse(scheme + requestHost + requestPath)
		return
	case ExactPathType:
		if requestPath != p.Path {
			logger.Logger.
				WithFields(logrus.Fields{"host": requestHost, "path": requestPath}).
				Warn("Mismatched route")

			return
		}

		requestURL, _ = url.Parse(scheme + requestHost + requestPath)
		return
	}

	logger.Logger.
		WithFields(logrus.Fields{"host": requestHost, "path": requestPath}).
		Warn("Mismatched route")

	return
}
