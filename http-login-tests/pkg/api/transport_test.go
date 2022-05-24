package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type MockRoundTripper struct {
	RoundTripOutput *http.Response
}

func (m MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") != "Bearer 123" {
		return nil, fmt.Errorf("wrong Authorization header: %s", req.Header.Get("Authorization"))
	}
	return m.RoundTripOutput, nil
}

func TestRoundtrip(t *testing.T) {
	loginResponse := LoginResponse{
		Token: "123",
	}
	loginResponseBytes, err := json.Marshal(loginResponse)
	if err != nil {
		t.Errorf("marshal error: %s", err)
	}

	jwtTransport := MyJWTTransport{
		HTTPClient: MockClient{
			PostResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(loginResponseBytes)),
			},
		},
		transport: MockRoundTripper{
			RoundTripOutput: &http.Response{
				StatusCode: 200,
			},
		},
		password: "xyz",
	}
	req := &http.Request{
		Header: make(http.Header),
	}
	res, err := jwtTransport.RoundTrip(req)
	if err != nil {
		t.Errorf("got error: %s", err)
		t.FailNow()
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", res.StatusCode)
	}
}
