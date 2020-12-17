package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/stretchr/testify/assert"
)

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		if body == "" {
			// mirror back the headers
			b, _ := json.Marshal(r.Header)
			fmt.Fprintln(w, string(b))
		} else {
			fmt.Fprintln(w, body)
		}
	}))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}

	return server, client
}

func assertJSONEqual(t *testing.T, expected, actual interface{}) {
	e, err := json.Marshal(expected)
	assert.NoError(t, err)
	a, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(e), string(a))
}

func TestHTTPFile(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"hello": "world"}`)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.WithHTTPClient(ctx, client)

	data := &Data{
		ds: map[string]config.DataSource{
			"foo": {URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/foo",
			}},
		},
		ctx: ctx,
	}

	expected := map[string]interface{}{
		"hello": "world",
	}

	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	actual, err = data.Datasource(server.URL)
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}

func TestHTTPFileWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, jsonMimetype, "")
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.WithHTTPClient(ctx, client)

	sources := make(map[string]config.DataSource)
	sources["foo"] = config.DataSource{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/foo",
		},
		Header: http.Header{
			"Foo":             {"bar"},
			"foo":             {"baz"},
			"User-Agent":      {},
			"Accept-Encoding": {"test"},
		},
	}
	data := &Data{
		ds:  sources,
		ctx: ctx,
	}
	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}
	data = &Data{
		ds:           sources,
		extraHeaders: map[string]http.Header{server.URL: expected},
		ctx:          ctx,
	}
	actual, err = data.Datasource(server.URL)
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}
