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
	if constants.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", constants.RelayProxy))
		proxyURL, err := url.Parse(constants.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", constants.RelayProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if constants.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(constants.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}
}
