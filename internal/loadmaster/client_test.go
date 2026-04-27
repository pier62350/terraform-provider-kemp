// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package loadmaster

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCall_InjectsAPIKeyAndCommand(t *testing.T) {
	t.Parallel()

	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accessv2" {
			t.Fatalf("expected /accessv2, got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, WithAPIKey("k123"), WithInsecureSkipVerify(false))
	if err := c.call(context.Background(), "showvs", map[string]string{"vs": "!1"}, nil); err != nil {
		t.Fatalf("call: %v", err)
	}

	if got["cmd"] != "showvs" {
		t.Errorf("cmd: want showvs, got %v", got["cmd"])
	}
	if got["apikey"] != "k123" {
		t.Errorf("apikey: want k123, got %v", got["apikey"])
	}
	if got["vs"] != "!1" {
		t.Errorf("vs: want !1, got %v", got["vs"])
	}
}

func TestCall_PropagatesFailStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":422,"status":"fail","message":"Unknown VS"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, WithAPIKey("k"))
	err := c.call(context.Background(), "showvs", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound: want true, got false (err=%v)", err)
	}
}

func TestCall_HTTP500BodySurfaced(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`gateway timeout`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, WithBasicAuth("bal", "secret"))
	err := c.call(context.Background(), "showvs", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "gateway timeout") {
		t.Errorf("err should include body, got: %v", err)
	}
}
