package httpserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/gorilla/mux"
	"github.com/phuslu/log"
	"github.com/rs/cors"
)

// LoggingMiddlewareConfig configuração para o middleware de logging
type LoggingMiddlewareConfig struct {
	// IgnorePaths lista de prefixos de paths que NÃO devem gerar logs
	// Exemplo: []string{"/assets/", "/static/", "/favicon.ico"}
	IgnorePaths []string
	// Enabled habilita ou desabilita completamente o logging (padrão: true)
	Enabled bool
}

func New(r *mux.Router, conf *config.Config, opts *cors.Options) *http.Server {
	return NewWithLogConfig(r, conf, opts, nil)
}

// NewWithLogConfig cria um servidor HTTP com configuração customizada de logging
func NewWithLogConfig(r *mux.Router, conf *config.Config, opts *cors.Options, logCfg *LoggingMiddlewareConfig) *http.Server {

	if conf.HttpConfig.Logger == nil {
		conf.HttpConfig.Logger = conf.GetGlobalLogger()
	}

	r.Use(LoggingMiddlewareWithConfig(logCfg))

	r.MethodNotAllowedHandler = defaultMethodNotAllowedHandler()
	r.NotFoundHandler = defaultNotFoundHandler()

	var handler http.Handler = r

	if opts != nil {
		handler = cors.New(*opts).Handler(r)
	}

	SRV_HTTP_PORT := os.Getenv("SRV_HTTP_PORT")
	if SRV_HTTP_PORT != "" {
		conf.PORT = SRV_HTTP_PORT
	} else {
		conf.PORT = "3000"
	}

	srv := &http.Server{
		ReadTimeout:  10 * time.Second, // Aguarda 10 segundos
		WriteTimeout: 10 * time.Second, // Responde em 10 segundos
		Addr:         ":" + conf.PORT,
		Handler:      handler,
		// ErrorLog:     log.New(os.Stderr, "logger: ", log.Lshortfile),
		ErrorLog: log.DefaultLogger.Std("", 0),
	}

	return srv
}

// LoggingMiddlewareWithConfig retorna um middleware de logging configurável
func LoggingMiddlewareWithConfig(cfg *LoggingMiddlewareConfig) func(http.Handler) http.Handler {
	// Se config for nil, usar padrão (tudo habilitado)
	if cfg == nil {
		cfg = &LoggingMiddlewareConfig{
			Enabled:     true,
			IgnorePaths: []string{},
		}
	}

	return func(next http.Handler) http.Handler {
		conf := config.NewDefaultConf()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			srw := &statusResponseWriter{ResponseWriter: w}
			next.ServeHTTP(srw, r)

			// Se logging está desabilitado, retornar sem fazer log
			if !cfg.Enabled {
				return
			}

			// Verificar se o path deve ser ignorado
			shouldIgnore := false
			for _, ignorePath := range cfg.IgnorePaths {
				if strings.HasPrefix(r.URL.Path, ignorePath) {
					shouldIgnore = true
					break
				}
			}

			// Se deve ignorar, retornar sem fazer log
			if shouldIgnore {
				return
			}

			conf.HttpConfig.Logger.Info().
				Str("AppName", conf.AppName).
				Str("AppVersion", conf.AppVersion).
				Str("AppCommitShortSha", conf.AppCommitShortSha).
				Str("UserAgent", r.UserAgent()).
				Str("HttpVersion", r.Proto).
				Str("Method", r.Method).
				Str("Host", r.Host).
				Str("RemoteAddr", r.RemoteAddr).
				Str("UserRealRemoteAddr", userIP(r)).
				Str("Path", r.URL.Path).
				Str("Duration", fmt.Sprintf("%v", time.Since(start))).
				Str("StatusCode", fmt.Sprintf("%v", srw.status)).
				// Str("RawQuery", r.URL.RawQuery).
				Msg(http.StatusText(srw.status))

			if conf.AppLogLevel == log.TraceLevel.String() {
				trac := conf.HttpConfig.Logger.Trace()
				for k, v := range r.Header {
					trac.Str(k, fmt.Sprintf("%v", v))
				}
				trac.Msg("Log Tracer")
			}
		})
	}
}

// LoggingMiddleware versão sem configuração (mantida para compatibilidade)
func LoggingMiddleware(next http.Handler) http.Handler {
	return LoggingMiddlewareWithConfig(nil)(next)
}

func userIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func ContentTypeJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (srw *statusResponseWriter) WriteHeader(status int) {
	if srw.status == 0 {
		srw.status = status
		srw.ResponseWriter.WriteHeader(status)
	}
}

// Hijack implementa http.Hijacker para suportar WebSocket upgrades
func (srw *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := srw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("responsewriter does not support hijacking")
	}
	return hijacker.Hijack()
}

// Flush implementa http.Flusher para suportar streaming
func (srw *statusResponseWriter) Flush() {
	if flusher, ok := srw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type HttpMsg struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

func (m *HttpMsg) toBytes() []byte {
	data, err := json.Marshal(m)

	if err != nil {
		log.Error().Str("FuntionName", "toBytes").Msg(err.Error())
	}

	return data
}

func (m *HttpMsg) Write(w http.ResponseWriter) {
	w.WriteHeader(m.Code)
	w.Write(m.toBytes())
}

var ErroHttpMsgPageNotFound HttpMsg = HttpMsg{
	Msg:  "Erro Page Not Found",
	Code: http.StatusNotFound,
}

var ErroHttpMsgMethodNotAllowed HttpMsg = HttpMsg{
	Msg:  "Erro Method Not Allowed",
	Code: http.StatusMethodNotAllowed,
}

func defaultMethodNotAllowedHandler() http.Handler {
	return LoggingMiddlewareWithConfig(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		ErroHttpMsgMethodNotAllowed.Write(w)
	}))
}

func defaultNotFoundHandler() http.Handler {
	return LoggingMiddlewareWithConfig(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		ErroHttpMsgPageNotFound.Write(w)
	}))
}
