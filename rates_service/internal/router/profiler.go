package router

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Profiler(auth func(next http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	if auth != nil {
		r.Use(auth)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	var profileSupportsDelta = map[string]bool{
		"allocs":       true,
		"block":        true,
		"goroutine":    true,
		"heap":         true,
		"mutex":        true,
		"threadcreate": true,
	}

	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.12:src/net/http/pprof/pprof.go
	r.HandleFunc("/pprof/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := r.URL.Path
		ind := strings.Index(cur, "/pprof/")
		suffix := cur[ind+7:]
		if profileSupportsDelta[suffix] {
			r.URL.Path = "/debug/pprof/" + suffix
		}
		pprof.Index(w, r)
	}))
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.HandleFunc("/vars", expVars)
	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
