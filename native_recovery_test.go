//go:build integration

package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestNativeSessionExpiryKeepsBrowserAndExistingConfig(t *testing.T) {
	engine := os.Getenv("ARIA_TEST_ENGINE_DIR")
	if engine == "" {
		t.Skip("set ARIA_TEST_ENGINE_DIR to the pinned 3x-ui engine directory")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0700); err != nil {
		t.Fatal(err)
	}
	bin, err := prepareRuntimeBin(engine, dir)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	cfg := Config{Port: 8080, DataDir: dir, EngineDir: engine, PublicURL: "https://native-recovery.example", AdminUser: "admin", AdminPassword: "Test-only-password-123", Mode: "cloud", Bind: "127.0.0.1"}
	env := append(os.Environ(), "XUI_DB_FOLDER="+dir, "XUI_LOG_FOLDER="+filepath.Join(dir, "logs"), "XUI_BIN_FOLDER="+bin, "XRAY_LOCATION_ASSET="+filepath.Join(engine, "bin"), "XUI_ENABLE_FAIL2BAN=false")
	setup := exec.CommandContext(ctx, filepath.Join(engine, "x-ui"), "setting", "-username", cfg.AdminUser, "-password", cfg.AdminPassword, "-port", strconv.Itoa(port), "-listenIP", "127.0.0.1", "-webBasePath", "/")
	setup.Env, setup.Dir = env, dir
	if err := setup.Run(); err != nil {
		t.Fatal("native initialization failed", err)
	}
	log, err := os.Create(filepath.Join(dir, "native-private.log"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.Close() })
	cmd := exec.CommandContext(ctx, filepath.Join(engine, "x-ui"))
	cmd.Env, cmd.Dir, cmd.Stdout, cmd.Stderr = env, dir, log, log
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	})
	a, err := NewApp(cfg, "http://127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 60; i++ {
		if a.refresh(ctx) == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !a.snapshot().Ready {
		t.Fatal("native core did not become ready")
	}
	cookie, csrf := autoLogin(t, a, cfg.PublicURL)
	request := func(path, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := browserRequest("POST", path, cfg.PublicURL, body).WithContext(ctx)
		r.AddCookie(cookie)
		r.Header.Set("X-ATIA-CSRF", csrf)
		w := httptest.NewRecorder()
		a.ServeHTTP(w, r)
		return w
	}
	create := `{"name":"native recovery test","count":1,"quotaGB":1,"days":7,"profiles":["vless-ws"]}`
	w := request("/api/users", create)
	if w.Code != http.StatusOK || len(a.snapshot().Users) != 1 {
		t.Fatal("initial config creation failed", w.Code, w.Body.String())
	}
	old := a.snapshot().Users[0]
	e := a.service
	u, _ := url.Parse(a.base)
	expire := func() {
		// Expiring the native cookie accelerates the six-hour default boundary.
		// The wrapper's browser session and the running Xray process stay intact.
		e.http.Jar.SetCookies(u, []*http.Cookie{{Name: "3x-ui", Path: "/", MaxAge: -1}})
	}
	expire()
	bare, _ := http.NewRequestWithContext(ctx, "GET", a.base+"/panel/api/inbounds/list", nil)
	bare.Header.Set("Accept", "application/json")
	resp, err := e.http.Do(bare)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("pinned 3x-ui must hide an anonymous non-XHR API request with 404", resp.StatusCode)
	}
	if err := a.refresh(ctx); err != nil {
		t.Fatal("background sync failed to recover the expired native session", err)
	}
	if !a.snapshot().Ready || a.coreStatus().Recoveries != 1 {
		t.Fatal("native session was not recovered exactly once")
	}
	expire()
	a.coreFailed()
	w = request("/api/core/reconnect", `{}`)
	if w.Code != http.StatusOK || !a.snapshot().Ready || a.coreStatus().Recoveries != 2 {
		t.Fatal("original browser could not reconnect after a second expiration", w.Code, w.Body.String())
	}
	w = request("/api/users", create)
	var created struct {
		Created []string `json:"created"`
		Warning string   `json:"warning"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || w.Code != http.StatusOK || len(created.Created) != 1 || created.Warning != "" || len(a.snapshot().Users) != 2 {
		t.Fatal("original browser could not create a new config after recovery", w.Code, w.Body.String(), err)
	}
	current, ok := a.findUser(old.Email)
	if !ok || current.Subscription != old.Subscription || marshalString(current.Links) != marshalString(old.Links) || current.Quota != old.Quota || current.Expiry != old.Expiry {
		t.Fatal("recovery changed or removed an existing config")
	}
}
