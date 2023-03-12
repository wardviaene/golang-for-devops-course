package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type WordsOutput struct {
	Page  string   `json:"page"`
	Input string   `json:"input"`
	Words []string `json:"words"`
}

type OccurrenceOutput struct {
	Page  string         `json:"page"`
	Words map[string]int `json:"words"`
}

type LoginRequest struct {
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

type WordsHandler struct {
	words       []string
	password    string
	tokenSecret []byte
}

func (ct *WordsHandler) wordsHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input != "" {
		ct.words = append(ct.words, input)
	}

	wordsOutput := WordsOutput{
		Page:  "words",
		Input: input,
		Words: ct.words,
	}
	out, err := json.Marshal(wordsOutput)
	if err != nil {
		fmt.Fprintf(w, "marshal error")
		return
	}
	fmt.Fprint(w, string(out))
}

func (ct *WordsHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 Not Found")
		return
	}
	fmt.Fprintf(w, "The server is running!")
	return
}

func (ct *WordsHandler) occurrenceHandler(w http.ResponseWriter, r *http.Request) {
	words := make(map[string]int)
	for _, v := range ct.words {
		if _, ok := words[v]; ok {
			words[v]++
		} else {
			words[v] = 1
		}
	}
	occurrenceOutput := OccurrenceOutput{
		Page:  "occurrence",
		Words: words,
	}
	out, err := json.Marshal(occurrenceOutput)
	if err != nil {
		fmt.Fprintf(w, "marshal error")
		return
	}
	fmt.Fprint(w, string(out))
}

func (ct *WordsHandler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Not a POST request")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Readall error")
		return
	}
	var loginRequest LoginRequest

	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unmarshal error")
		return
	}

	if ct.password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "start the test-server with a password first")
		fmt.Printf("Returned HTTP 400 error to client: server has no password set\n")
		return
	}

	if loginRequest.Password != ct.password {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Password doesn't match")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(ct.tokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Token signing error")
		return
	}

	if err := json.NewEncoder(w).Encode(LoginResponse{Token: tokenString}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Token encoding error")
		return
	}
}

func (ct *WordsHandler) authMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct.password != "" {
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "Authorization header not set")
				return
			}
			tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return ct.tokenSecret, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "Authorization token invalid: %s", err)
				return
			}
		}
		next(w, r)
	})
}

func (wh *WordsHandler) loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wh.password == "" {
			log.Println(r.Method, r.URL.Path)
		} else {
			log.Println(r.Method, r.URL.Path, "Auth:"+r.Header.Get("Authorization"))
		}

		h.ServeHTTP(w, r)
	})
}

func getRandomSecret() []byte {
	b := make([]byte, 30)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func main() {
	port := "8080"
	testListener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't start server on port %q: %s\n", port, err)
		os.Exit(1)
	}
	testListener.Close()

	password := flag.String("password", "", "password protect our API")

	flag.Parse()

	wh := &WordsHandler{
		words:       []string{},
		password:    *password,
		tokenSecret: getRandomSecret(),
	}

	rl := &RateLimit{
		hits: make(map[string]uint64),
	}

	mux := http.NewServeMux()

	mux.Handle("/words", wh.authMiddleware(wh.wordsHandler))
	mux.Handle("/occurrence", wh.authMiddleware(wh.occurrenceHandler))
	mux.HandleFunc("/assignment1", wh.assignment1)
	mux.HandleFunc("/ratelimit", rl.ratelimit)
	mux.HandleFunc("/", wh.indexHandler)
	mux.HandleFunc("/login", wh.login)
	fmt.Printf("Starting server on port %v...\n", port)
	http.ListenAndServe(":"+port, wh.loggingHandler(mux))
}
