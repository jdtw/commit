package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
)

// GET requests to '/' or '/verify' are redirected to the github repo.
const repo = "https://github.com/jdtw/commit"

var (
	port      = flag.Int("port", 8080, "listening port")
	ratelimit = flag.Duration("rate", 200*time.Millisecond, "global rate limit")
)

func main() {
	flag.Parse()
	addr := fmt.Sprint(":", *port)
	log.Printf("listening on %s", addr)

	limiter := rate.NewLimiter(rate.Every(*ratelimit), 5)
	srv := http.NewServeMux()
	srv.HandleFunc("/verify", postHandler(limiter, verify()))
	srv.HandleFunc("/", postHandler(limiter, commit()))

	log.Fatal(http.ListenAndServe(addr, srv))
}

type commitment struct {
	Message string `yaml:"message"`
	Commit  string `yaml:"commit"`
}

// commit returns a handler that reads the message from the request
// body, appends some entropy, and creates a commitment by taking the
// SHA256 hash of the message+entropy. It returns the message and
// digest as YAML.
func commit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			internalError(w, "failed to read body: %v", err)
			return
		}
		if len(body) == 0 {
			badRequest(w, "")
			return
		}
		entropy := make([]byte, 16)
		if _, err := rand.Read(entropy); err != nil {
			internalError(w, "failed to read entropy: %v", err)
			return
		}
		encoded := base64.RawURLEncoding.EncodeToString(entropy)
		msg := fmt.Sprintf("%s (%s)", string(body), encoded)
		digest := sha256.Sum256([]byte(msg))
		w.Header().Set("Content-Type", "text/yaml")
		yaml.NewEncoder(w).Encode(&commitment{
			Message: msg,
			Commit:  hex.EncodeToString(digest[:]),
		})
	}
}

// verify returns a handler that decodes the commitment YAML in the
// body and verifies that the digest of the message matches the
// commit. If it does, it returns "true" in the body, otherwise
// "false".
func verify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := commitment{}
		if err := yaml.NewDecoder(r.Body).Decode(&c); err != nil {
			badRequest(w, "failed to decody body: %v", err)
			return
		}
		got, err := hex.DecodeString(c.Commit)
		if err != nil {
			badRequest(w, "bad commitment: %v", err)
			return
		}
		want := sha256.Sum256([]byte(c.Message))
		verified := bytes.Equal(got, want[:])
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strconv.FormatBool(verified))
	}
}

// postHandler is middleware that redirects GET requests to the github
// repo, returns method not allowed for anything other than POST, and
// does rate limiting.
func postHandler(l *rate.Limiter, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.Redirect(w, r, repo, http.StatusFound)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !l.Allow() {
			http.Error(w, "Are you like, trying to mine bitcoin or something?", http.StatusTooManyRequests)
			return
		}
		h(w, r)
	}
}

func internalError(w http.ResponseWriter, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Print(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}

func badRequest(w http.ResponseWriter, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	http.Error(w, msg, http.StatusBadRequest)
}
