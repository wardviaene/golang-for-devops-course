package api

import "net/http"

type Options struct {
	Password string
	LoginURL string
}

type ClientIface interface {
	Get(url string) (resp *http.Response, err error)
}

type APIIface interface {
	DoGetRequest(requestURL string) (Response, error)
}

type api struct {
	Options Options
	Client  ClientIface
}

func New(options Options) APIIface {
	return api{
		Options: options,
		Client: &http.Client{
			Transport: MyJWTTransport{
				transport: http.DefaultTransport,
				password:  options.Password,
				loginURL:  options.LoginURL,
			},
		},
	}
}
