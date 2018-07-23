package httptools

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CORSHandler struct {
	Handler      http.Handler
	AllowOrigin  string
	AllowMethods []string
	AllowHeaders []string
	MaxAge       time.Duration
}

func (c CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		if c.AllowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", c.AllowOrigin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		if len(c.AllowHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowHeaders, ", "))
		}
		if len(c.AllowMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowMethods, ", "))
		}
		if c.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.MaxAge*time.Second))
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	c.Handler.ServeHTTP(w, r)
}
