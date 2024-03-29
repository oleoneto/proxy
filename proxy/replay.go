package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cleopatrio/proxy/logger"
	"github.com/sirupsen/logrus"
)

type ReplayRequest struct {
	RequestPath string
	MatchedPath ProxyPath
	Method      string
	Body        []byte
	Headers     map[string][]string
}

func (xy *Server) ReplayRequest(snapshot ReplayRequest) error {
	if !snapshot.MatchedPath.EnableReplay || !xy.Proxyfile.ReplayEnabled() {
		return nil
	}

	reqTime := time.Now()
	duration := time.Duration(time.Since(reqTime))

	snapshot.Headers["Content-Type"] = []string{"application/json"}

	for _, h := range xy.Proxyfile.ReplayConfig().SuppressedHeaders {
		delete(snapshot.Headers, h.Name)
	}

	host := xy.Proxyfile.ReplayConfig().Host + func() string {
		port := xy.Proxyfile.ReplayConfig().Port
		if port > 0 {
			return fmt.Sprintf(":%d", port)
		}
		return ""
	}()

	reqPath := func() string {
		switch xy.Proxyfile.ReplayConfig().PathRewriteSettings.Strategy {
		case RewritePathStrategy:
			return xy.Proxyfile.ReplayConfig().PathRewriteSettings.Path
		case SuppressPathStrategy:
			return ""
		default:
			return snapshot.RequestPath
		}
	}()

	requestURL, err := url.Parse(xy.Proxyfile.ReplayConfig().Scheme + "://" + host + reqPath)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request.id": snapshot.Headers[PxFile.Annotations.HTTPRequestIdHeader],
			"url":        requestURL,
			"error":      err,
		}).Error("Invalid replay URL ❌")

		return err
	}

	method := func() string {
		switch xy.Proxyfile.ReplayConfig().MethodRewriteSettings.Strategy {
		case RewriteMethodStrategy:
			return xy.Proxyfile.ReplayConfig().MethodRewriteSettings.Method
		default:
			return snapshot.Method
		}
	}()

	data, _ := json.Marshal(map[string]any{
		"body":    snapshot.Body,
		"path":    snapshot.RequestPath,
		"method":  snapshot.Method,
		"headers": snapshot.Headers,
		// "remote_ip": snapshot.Context().RemoteIP(),
	})

	res, err := HTTPClient.Do(&http.Request{
		Method: method,
		Header: snapshot.Headers,
		URL:    requestURL,
		Body:   &RequestBody{Data: data},
	})

	status := -1
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request.id": snapshot.Headers[PxFile.Annotations.HTTPRequestIdHeader],
			"url":        requestURL,
			"method":     method,
			"error":      err,
		}).Error("HTTP replay failed ❌")

		return nil
	}

	status = res.StatusCode
	defer res.Body.Close()

	fmt.Println(string(data))

	logger.Logger.WithFields(logrus.Fields{
		"request.id": snapshot.Headers[PxFile.Annotations.HTTPRequestIdHeader],
		"duration":   duration.Nanoseconds(),
		"url":        requestURL.String(),
		"method":     method,
		"status":     status,
		"error":      err,
	}).Info("Replayed HTTP request ⏪")

	return nil
}
