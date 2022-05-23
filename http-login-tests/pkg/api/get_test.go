package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type MockClient struct {
	GetResponse *http.Response
}

func (m MockClient) Get(url string) (resp *http.Response, err error) {
	return m.GetResponse, nil
}

func TestDoGetRequest(t *testing.T) {
	words := WordsPage{
		Page: Page{"words"},
		Words: Words{
			Input: "abc",
			Words: []string{"a", "b"},
		},
	}
	wordsBytes, err := json.Marshal(words)
	if err != nil {
		t.Errorf("marshal error: %s", err)
	}
	apiInstance := api{
		Options: Options{},
		Client: MockClient{
			GetResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(wordsBytes)),
			},
		},
	}

	response, err := apiInstance.DoGetRequest("http://localhost/words")
	if err != nil {
		t.Errorf("DoGetRequest error: %s", err)
	}
	if response == nil {
		t.Errorf("Response is nil")
	}
	if response.GetResponse() != `Words: a, b` {
		t.Errorf("Got wrong output: %s", response.GetResponse())
	}
}
