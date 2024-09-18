package utils

import "net/http"

type ClickhouseClient struct {
	URL      string
	Username string
	Password string
	Client   *http.Client
}
