package recovery

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// Recover is a middleware that recovers panics and maps them to errors.
// Default configuration:
//    ContentType:   application/json
//    ResponseStatus: 500 (http.StatusInternalServerError)
//    StackSize:     MinimumStackSize (4KB)
//    Logger:        standard log
//    ErrHandler:    the stack traces are output to the log and not to the HTTP response.
func Recover(options ...Option) func(http.Handler) http.Handler {
	c := NewConfig()
	for _, v := range options {
		v(c)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					var msg string
					switch x := v.(type) {
					case string:
						msg = fmt.Sprintf("panic: %s", x)
					case error:
						msg = fmt.Sprintf("panic: %s", x)
					default:
						msg = "unknown panic"
					}
					buf := make([]byte, c.StackSize)
					buf = buf[:runtime.Stack(buf, false)]
					stack := strings.Split(string(buf), "\n")
					if len(stack) > 3 {
						stack = stack[3:]
					}
					c.ErrHandler(c, w, msg, stack)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}

// MinimumStackSize represents minimum stack size.
const MinimumStackSize = 4 << 10

// DefaultErrorHandler is a default error handler that outputs stack trace to the log,
// but does not output it to the HTTP response.
func DefaultErrorHandler(c *Config, w http.ResponseWriter, msg string, stack []string) {
	c.Logger.Printf("%s\n%s", msg, strings.Join(stack, "\n"))
	w.Header().Set("Content-Type", c.ContentType)
	w.WriteHeader(c.ResponseStatus)
}

// Config represents the configuration of recovery.
type Config struct {
	ContentType    string
	ResponseStatus int
	StackSize      int
	Logger         *log.Logger
	ErrHandler     func(c *Config, w http.ResponseWriter, msg string, stack []string)
}

// Option represents options for recovery.
type Option func(c *Config)

// NewConfig creates the recovery config with default options.
// ContentType:   application/json
// ResponseStatus: 500 (http.StatusInternalServerError)
// StackSize:     MinimumStackSize (4KB)
// Logger:        default log
// ErrHandler:    the stack traces are output to the log and not to the HTTP response.
func NewConfig() *Config {
	return &Config{
		ContentType:    "application/json",
		ResponseStatus: http.StatusInternalServerError,
		StackSize:      MinimumStackSize,
		Logger:         log.New(os.Stderr, "", log.LstdFlags),
		ErrHandler:     DefaultErrorHandler,
	}
}

// ContentType specify the content-type of the HTTP response.
func ContentType(val string) Option {
	return func(c *Config) {
		c.ContentType = val
	}
}

// ResponseStatus sets the HTTP response status after recovery.
func ResponseStatus(val int) Option {
	return func(c *Config) {
		c.ResponseStatus = val
	}
}

// StackSize sets the amount to load runtime stack. (> MinimumStackSize)
func StackSize(val int) Option {
	return func(c *Config) {
		if val <= MinimumStackSize {
			return
		}
		c.StackSize = val
	}
}

// Logger sets the logger to be used for the error handler.
func Logger(logger *log.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// ErrorHandler sets the error handler for outputting HTTP responses and logs.
func ErrorHandler(h func(c *Config, w http.ResponseWriter, msg string, stack []string)) Option {
	return func(c *Config) {
		c.ErrHandler = h
	}
}
