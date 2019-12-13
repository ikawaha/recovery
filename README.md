Panic recovery middleware for Go HTTP handlers
---

This is a customizable middleware that recovers panic for HTTP handlers.

## Options

| option| default|
|:---|:---|
| ContentType|    application/json|
| ResponseStatus| 500 (http.StatusInternalServerError)|
| StackSize|      MinimumStackSize (4KB)|
| Logger|         standard log|
| ErrorHandler|   the stack traces are output to the log and not to the HTTP response.|


### DefaultErrorHandler
```go
// DefaultErrorHandler is a default error handler that outputs stack trace to the log,
// but does not output it to the HTTP response.
func DefaultErrorHandler(c *Config, w http.ResponseWriter, msg string, stack []string) {
	c.Logger.Printf("%s\n%s", msg, strings.Join(stack, "\n"))
	w.Header().Set("Content-Type", c.ContentType)
	w.WriteHeader(c.ResponseStatus)
}
```

---

License MIT