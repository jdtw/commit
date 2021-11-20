package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
)

var port = flag.Int("port", 8080, "listening port")

func main() {
	flag.Parse()
	addr := fmt.Sprint(":", *port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, newHandler()))
}

type commitment struct {
	Message string `yaml:"message"`
	Commit  string `yaml:"commit"`
}

func commit(msg string) (*commitment, error) {
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		return nil, err
	}
	msg = fmt.Sprintf("%s %s", msg, hex.EncodeToString(entropy))
	digest := sha256.Sum256([]byte(msg))
	return &commitment{
		Message: msg,
		Commit:  hex.EncodeToString(digest[:]),
	}, nil
}

func newHandler() http.Handler {
	srv := http.NewServeMux()
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 5)
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !limiter.Allow() {
			http.Error(w, "Are you like, trying to mine bitcoin or something? Or maybe somebody else is...", http.StatusTooManyRequests)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			internalError(w, "failed to read body: %v", err)
			return
		}
		if len(body) == 0 {
			http.Error(w, "No message...", http.StatusBadRequest)
			return
		}
		response, err := commit(string(body))
		w.Header().Set("Content-Type", "text/yaml")
		yaml.NewEncoder(w).Encode(response)
	})
	return srv
}

func internalError(w http.ResponseWriter, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
