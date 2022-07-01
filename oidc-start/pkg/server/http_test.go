package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

/* Global code for all tests */
var privkeyPem []byte

var testConfig Config

func TestMain(m *testing.M) {
	err := testSetup()
	if err != nil {
		log.Fatalf("test setup failed: %s", err)
	}
	code := m.Run()
	os.Exit(code)
}

func testSetup() error {
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	privkeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privkeyPem = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)
	// populate testConfig
	testConfig = Config{
		Apps: map[string]AppConfig{
			"app1": {
				ClientID:     "1-2-3-4",
				ClientSecret: "secret",
				Issuer:       "http://localhost:8080",
				RedirectURIs: []string{"http://localhost:8082/callback"},
			},
		},
	}
	return nil
}

func TestStart(t *testing.T) {
	httpServer := &http.Server{Addr: ":8080"}

	go func() {
		err := Start(httpServer, privkeyPem, testConfig)
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Start error: %s\n", err)
		}
	}()

	time.Sleep(1 * time.Second) // give time for the http server to start

	endpoints := []string{"/authorization", "token", "login", "jwks.json", "/.well-known/openid-configuration", "userinfo"}
	for _, endpoint := range endpoints {
		addr := httpServer.Addr
		if strings.HasPrefix(addr, ":") {
			addr = "http://localhost" + addr
		} else {
			addr = "http://" + addr
		}
		res, err := http.Get(addr + "/" + endpoint)
		if err != nil {
			t.Fatalf("http get error: %s", err)
		}
		if res.StatusCode == 404 || res.StatusCode >= 500 {
			t.Errorf("Endpoint %s not available. Statuscode: %d", endpoint, res.StatusCode)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		t.Fatalf("Could not shut down http server")
	}

}
