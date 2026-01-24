package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/utils/logger"
)

var HTTPClient *http.Client

func Init() {
	var transport http.RoundTripper
	if constants.GetRelayProxy() != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", constants.GetRelayProxy()))
		proxyURL, err := url.Parse(constants.GetRelayProxy())
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", constants.GetRelayProxy()))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if constants.GetRelayTimeout() == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(constants.GetRelayTimeout()) * time.Second,
			Transport: transport,
		}
	}
}
