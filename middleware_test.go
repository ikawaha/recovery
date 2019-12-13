package recovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, r *http.Request) {
			panic("!!!")
		},
	)

	// recovery middleware
	var handler http.Handler = mux
	handler = Recover()(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected OK, got %v", resp.Status)
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("unexpected error, %v", err)
	} else {
		resp.Body.Close()
		if len(body) != 0 {
			t.Errorf("expected empty response body, got %s", body)
		}
	}
}

func TestRecoverMiddlewareWithOptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, r *http.Request) {
			panic("aloha")
		},
	)

	// recovery middleware
	var (
		handler http.Handler = mux
		buf     bytes.Buffer
	)
	handler = Recover(
		ContentType("application/xml"),
		ResponseStatus(http.StatusOK),
		Logger(log.New(&buf, "", 0)),
		StackSize(1), // too small
	)(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected OK, got %v", resp.Status)
	}
	if got, expected := resp.Header.Get("Content-Type"), "application/xml"; got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("unexpected error, %v", err)
	} else {
		resp.Body.Close()
		if len(body) != 0 {
			t.Errorf("expected empty response body, got %s", body)
		}
	}
	if got, expected := buf.String(), "panic: aloha"; !strings.HasPrefix(got, expected) {
		t.Errorf("expected prefix %q, got %q", expected, got)
	}
}

func TestRecoverMiddlewareWithCustomErrorHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, r *http.Request) {
			panic("aloha")
		},
	)

	// recovery middleware
	var (
		handler http.Handler = mux
		buf     bytes.Buffer
	)
	handler = Recover(
		ErrorHandler(func(c *Config, w http.ResponseWriter, msg string, stack []string) {
			w.Header().Set("Content-Type", c.ContentType)
			w.WriteHeader(c.ResponseStatus)
			obj := map[string]string{
				"error": fmt.Sprintf("%s\n%s", msg, strings.Join(stack, "\n")),
			}
			b, err := json.Marshal(obj)
			if err != nil {
				t.Fatal(err)
			}
			w.Write(b)
		}),
	)(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected OK, got %v", resp.Status)
	}
	if got, expected := resp.Header.Get("Content-Type"), "application/json"; got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("unexpected error, %v", err)
	} else {
		resp.Body.Close()
		if got, expected := string(body), `{"error":"panic: aloha`; !strings.HasPrefix(got, expected) {
			t.Errorf("expected prefix %q, got %q", expected, got)
		}
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty response body, got %s", buf.String())
	}
}
