// Package http is a pgSCV http helper
package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cherts/pgscv/internal/log"
)

// AuthConfig defines configuration settings for authentication.
type AuthConfig struct {
	EnableAuth bool   // flag tells about authentication should be enabled
	Username   string `yaml:"username"` // username used for basic authentication
	Password   string `yaml:"password"` // password used for basic authentication
	EnableTLS  bool   // flag tells about TLS should be enabled
	Keyfile    string `yaml:"keyfile"`  // path to key file
	Certfile   string `yaml:"certfile"` // path to certificate file
}

// Validate check authentication options of AuthConfig and returns toggle flags.
func (cfg AuthConfig) Validate() (bool, bool, error) {
	var enableAuth, enableTLS bool

	if (cfg.Username == "" && cfg.Password != "") || (cfg.Username != "" && cfg.Password == "") {
		return false, false, fmt.Errorf("authentication settings invalid")
	}

	if (cfg.Keyfile == "" && cfg.Certfile != "") || (cfg.Keyfile != "" && cfg.Certfile == "") {
		return false, false, fmt.Errorf("TLS settings invalid")
	}

	if cfg.Username != "" && cfg.Password != "" {
		enableAuth = true
	}

	if cfg.Keyfile != "" && cfg.Certfile != "" {
		enableTLS = true
	}

	return enableAuth, enableTLS, nil
}

// ServerConfig defines HTTP server configuration.
type ServerConfig struct {
	Addr string
	AuthConfig
}

// Server defines HTTP server.
type Server struct {
	config ServerConfig
	server *http.Server
}

// NewServer creates new HTTP server instance.
func NewServer(cfg ServerConfig,
	handlerMetrics func(http.ResponseWriter, *http.Request),
	targetsMetrics func(http.ResponseWriter, *http.Request),
	flushServiceConfig func(http.ResponseWriter, *http.Request),
) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot())
	if cfg.EnableAuth {
		mux.HandleFunc("/metrics", basicAuth(cfg.AuthConfig, handlerMetrics))
	} else {
		mux.HandleFunc("/metrics", handlerMetrics)
	}
	mux.HandleFunc("/targets", targetsMetrics)
	if cfg.EnableAuth {
		mux.HandleFunc("/flush-services-config", basicAuth(cfg.AuthConfig, flushServiceConfig))
	} else {
		mux.HandleFunc("/flush-services-config", flushServiceConfig)
	}

	return &Server{
		config: cfg,
		server: &http.Server{
			Addr:         cfg.Addr,
			Handler:      mux,
			IdleTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

// Serve method starts listening and serving requests.
func (s *Server) Serve() error {
	if s.config.EnableTLS {
		log.Infof("listen on https://%s", s.server.Addr)
		return s.server.ListenAndServeTLS(s.config.Certfile, s.config.Keyfile)
	}

	log.Infof("listen on http://%s", s.server.Addr)
	return s.server.ListenAndServe()
}

// handleRoot defines handler for '/' endpoint.
func handleRoot() http.HandlerFunc {
	const htmlTemplate = `<html>
<head><title>pgSCV / PostgreSQL metrics collector</title></head>
<body>
pgSCV / PostgreSQL metrics collector, for more info visit <a href="https://github.com/cherts/pgscv">Github</a> page.
<p><a href="/metrics">Metrics</a> (add ?target=service_id, to get metrics for one service)</p>
<p><a href="/targets">Targets</a></p>
<p><a href="/flush-services-config">Reload service config</a></p>
</body>
</html>
`

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(htmlTemplate))
		if err != nil {
			log.Warnln("response write failed: ", err)
		}
	})
}

func basicAuth(cfg AuthConfig, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			if username == cfg.Username && password == cfg.Password {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", StatusUnauthorized)
	})
}
