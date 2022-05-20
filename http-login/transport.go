package main

import "net/http"

type MyJWTTransport struct {
	transport http.RoundTripper
	token     string
}

func (t MyJWTTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.token)
	return t.transport.RoundTrip(req)
}
