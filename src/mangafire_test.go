package src

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "testing"
)

// The vrf tables are opaque constants copied from the site's signer. If they
// ever rotate, every /api/ call starts coming back 403 - these vectors say so
// immediately instead of leaving it looking like a Cloudflare problem.
func TestSignVrf(t *testing.T) {
  if !vrfReadyMangaFire() {
    t.Fatal("vrf tables did not decode")
  }

  cases := map[string]string{
    "/titles/kwyvw": "8sK3xtqdFdsnTwfb6Q",
    "/titles/kwyvw/chapters?language=en&limit=200&order=desc&page=1&sort=number": "8sK3xtqdFdsnTwfb6bZjox2VOlpOzAhCLmMkJCBZWqtdIv4xA-dQBdeXci9JDrAlMLdmUukZS7-oLlXnlm54gnliam65xKGvT5M",
  }

  for path, want := range cases {
    if got := signVrfMangaFire(path); got != want {
      t.Errorf("signVrfMangaFire(%q)\n got %q\nwant %q", path, got, want)
    }
  }
}

func TestAPIURL(t *testing.T) {
  // params must go out sorted by key, with vrf last
  got := apiURLMangaFire("/titles/kwyvw/chapters", [][2]string{
    {"sort", "number"},
    {"language", "en"},
    {"page", "1"},
    {"limit", "200"},
    {"order", "desc"},
  })

  want := "https://mangafire.to/api/titles/kwyvw/chapters?language=en&limit=200&order=desc&page=1&sort=number&vrf=8sK3xtqdFdsnTwfb6bZjox2VOlpOzAhCLmMkJCBZWqtdIv4xA-dQBdeXci9JDrAlMLdmUukZS7-oLlXnlm54gnliam65xKGvT5M"
  if got != want {
    t.Errorf("apiURLMangaFire()\n got %q\nwant %q", got, want)
  }

  gotNoParams := apiURLMangaFire("/titles/kwyvw", nil)
  wantNoParams := "https://mangafire.to/api/titles/kwyvw?vrf=8sK3xtqdFdsnTwfb6Q"
  if gotNoParams != wantNoParams {
    t.Errorf("apiURLMangaFire(no params)\n got %q\nwant %q", gotNoParams, wantNoParams)
  }
}

func TestHid(t *testing.T) {
  cases := map[string]string{
    "https://mangafire.to/title/kwyvw-nikaidou-kou-tanpenshuu-arigatou-tte-itte":            "kwyvw",
    "kwyvw-nikaidou-kou-tanpenshuu-arigatou-tte-itte":                                       "kwyvw",
    "https://mangafire.to/title/kwyvw-nikaidou-kou-tanpenshuu-arigatou-tte-itte/":           "kwyvw",
    "https://mangafire.to/title/kwyvw-nikaidou-kou-tanpenshuu-arigatou-tte-itte/chapter/302": "302",
    "https://mangafire.to/manga/kin-no-itoo.8wz3":                                           "8wz3",
    "kin-no-itoo.8wz3":                                                                      "8wz3",
    "kwyvw":                                                                                 "kwyvw",
    "":                                                                                      "",
  }

  for input, want := range cases {
    if got := hidMangaFire(input); got != want {
      t.Errorf("hidMangaFire(%q) = %q, want %q", input, got, want)
    }
  }
}

func TestTrimChapterNumber(t *testing.T) {
  cases := map[string]string{"7.0": "7", "10.5": "10.5", "7": "7"}

  for input, want := range cases {
    if got := trimChapterNumberMangaFire(input); got != want {
      t.Errorf("trimChapterNumberMangaFire(%q) = %q, want %q", input, got, want)
    }
  }
}

// The real page list, as returned by /api/chapters/{id}
func TestParsePages(t *testing.T) {
  body := `{"data":{"pages":[{"url":"https://nw8.mfcdn1.xyz/mf/aaa/h/p.jpg"},{"url":"https://nw8.mfcdn1.xyz/mf/bbb/h/p.jpg"}]}}`

  var pages mangaFirePagesResponse
  if err := json.Unmarshal([]byte(body), &pages); err != nil {
    t.Fatal(err)
  }
  if len(pages.Data.Pages) != 2 {
    t.Fatalf("got %d pages, want 2", len(pages.Data.Pages))
  }
  if pages.Data.Pages[0].URL != "https://nw8.mfcdn1.xyz/mf/aaa/h/p.jpg" {
    t.Errorf("unexpected url %q", pages.Data.Pages[0].URL)
  }
}

func TestParseChapters(t *testing.T) {
  body := `{"items":[{"id":3023040,"number":7,"name":"Love Swing-by"},{"id":3022890,"number":1.5,"name":""}],"meta":{"lastPage":1}}`

  var chapters mangaFireChaptersResponse
  if err := json.Unmarshal([]byte(body), &chapters); err != nil {
    t.Fatal(err)
  }
  if len(chapters.Items) != 2 {
    t.Fatalf("got %d chapters, want 2", len(chapters.Items))
  }
  if chapters.Items[0].ID != 3023040 || chapters.Items[0].Number.String() != "7" {
    t.Errorf("unexpected chapter %+v", chapters.Items[0])
  }
  if chapters.Items[1].Number.String() != "1.5" {
    t.Errorf("unexpected number %q", chapters.Items[1].Number.String())
  }
}

////////////////////////////////////////////////////////////////////////////////
// route selection
////////////////////////////////////////////////////////////////////////////////

type routeTestServerMangaFire struct {
  *httptest.Server
  requests []*http.Request
}

// A stand in for mangafire.to: _accept decides, per request, whether it is
// answered or refused with a Cloudflare style 403
func newRouteTestServerMangaFire(t *testing.T, _accept func(*http.Request) bool) *routeTestServerMangaFire {
  t.Helper()
  server := &routeTestServerMangaFire{}
  server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    server.requests = append(server.requests, r)
    if !_accept(r) {
      w.Header().Set("cf-mitigated", "challenge")
      w.WriteHeader(403)
      w.Write([]byte("<html><title>Just a moment...</title></html>"))
      return
    }
    w.Write([]byte(`{"data":{"title":"ok"}}`))
  }))
  t.Cleanup(server.Close)

  return server
}

func newRouteTestClientMangaFire(t *testing.T, _server *routeTestServerMangaFire, _cookies func() (string, string)) *mangaFireClient {
  t.Helper()
  t.Setenv("PAPIBAQUIGRAFO_MF_CLEARANCE", "")
  t.Setenv("PAPIBAQUIGRAFO_MF_WAF_PASS", "")

  return &mangaFireClient{
    http:             _server.Client(),
    baseURL:          _server.URL,
    firefoxUserAgent: "firefox-ua",
    cookieSource:     _cookies,
  }
}

func hasCookieMangaFire(_r *http.Request, _clearance string) bool {
  cookie, err := _r.Cookie("cf_clearance")
  return err == nil && cookie.Value == _clearance
}

// The plain route goes first: when it answers, Firefox is never even looked at
func TestPlainRouteFirst(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(*http.Request) bool { return true })
  cookieReads := 0
  client := newRouteTestClientMangaFire(t, server, func() (string, string) {
    cookieReads++
    return "abc", "1.ff"
  })

  if _, _, err := client.apiGet("/titles/kwyvw", nil); err != nil {
    t.Fatal(err)
  }
  if cookieReads != 0 {
    t.Errorf("read the Firefox cookies %d times before the plain route had failed", cookieReads)
  }
  if len(server.requests) != 1 {
    t.Fatalf("got %d requests, want 1", len(server.requests))
  }
  if _, err := server.requests[0].Cookie("cf_clearance"); err == nil {
    t.Error("the plain route sent a cookie")
  }
  if ua := server.requests[0].Header.Get("User-Agent"); ua != plainUserAgentMangaFire {
    t.Errorf("plain route user agent %q", ua)
  }
  if client.useCookies {
    t.Error("switched to the cookie route for no reason")
  }
}

// A refused plain request falls back on Firefox's cookies, paired with its user agent
func TestFallsBackOnFirefoxCookies(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(r *http.Request) bool { return hasCookieMangaFire(r, "abc") })
  client := newRouteTestClientMangaFire(t, server, func() (string, string) { return "abc", "1.ff" })

  body, _, err := client.apiGet("/titles/kwyvw", nil)
  if err != nil {
    t.Fatal(err)
  }
  if string(body) != `{"data":{"title":"ok"}}` {
    t.Errorf("unexpected body %q", body)
  }
  if len(server.requests) != 2 {
    t.Fatalf("got %d requests, want 2 (plain then cookies)", len(server.requests))
  }
  if _, err := server.requests[0].Cookie("cf_clearance"); err == nil {
    t.Error("the first request already carried a cookie")
  }
  if !hasCookieMangaFire(server.requests[1], "abc") {
    t.Error("the fallback request did not carry cf_clearance")
  }
  if cookie, err := server.requests[1].Cookie("waf_pass"); err != nil || cookie.Value != "1.ff" {
    t.Error("the fallback request did not carry waf_pass")
  }
  if ua := server.requests[1].Header.Get("User-Agent"); ua != "firefox-ua" {
    t.Errorf("fallback user agent %q, want Firefox's", ua)
  }
  if !client.useCookies {
    t.Error("did not stay on the cookie route")
  }
  if client.userAgent() != "firefox-ua" {
    t.Errorf("image downloads would use %q", client.userAgent())
  }
}

// No cookies to fall back on: the challenge is handed back so the caller parks on it
func TestChallengeWithoutCookiesIsReported(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(*http.Request) bool { return false })
  client := newRouteTestClientMangaFire(t, server, func() (string, string) { return "", "" })

  _, status, err := client.apiGet("/titles/kwyvw", nil)
  if !isChallengedMangaFire(err) {
    t.Fatalf("got %v (status %d), want a challenge", err, status)
  }
  if len(server.requests) != 1 {
    t.Errorf("got %d requests, want 1", len(server.requests))
  }
  if client.useCookies {
    t.Error("switched to the cookie route with no cookies")
  }
}

// On the cookie route a 403 means waf_pass turned over: the newer pair wins
func TestStaleCookiesRefreshed(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(r *http.Request) bool { return hasCookieMangaFire(r, "new") })
  client := newRouteTestClientMangaFire(t, server, func() (string, string) { return "new", "2.ff" })
  client.useCookies = true
  client.clearance = "old"
  client.wafPass = "1.ff"

  if _, _, err := client.apiGet("/titles/kwyvw", nil); err != nil {
    t.Fatal(err)
  }
  if len(server.requests) != 2 {
    t.Fatalf("got %d requests, want 2", len(server.requests))
  }
  if !hasCookieMangaFire(server.requests[1], "new") {
    t.Error("the retry did not carry the fresher cf_clearance")
  }
  if client.clearance != "new" || client.wafPass != "2.ff" || !client.useCookies {
    t.Errorf("client did not keep the fresher cookies: %+v", client)
  }
}

// Stale cookies, nothing fresher in Firefox, but the plain route answers again:
// go back to it rather than parking
func TestReturnsToPlainRouteWhenCookiesStale(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(r *http.Request) bool {
    _, err := r.Cookie("cf_clearance")
    return err != nil
  })
  client := newRouteTestClientMangaFire(t, server, func() (string, string) { return "old", "1.ff" })
  client.useCookies = true
  client.clearance = "old"
  client.wafPass = "1.ff"

  if _, _, err := client.apiGet("/titles/kwyvw", nil); err != nil {
    t.Fatal(err)
  }
  if len(server.requests) != 2 {
    t.Fatalf("got %d requests, want 2 (cookies then plain)", len(server.requests))
  }
  if _, err := server.requests[1].Cookie("cf_clearance"); err == nil {
    t.Error("the plain retry still carried a cookie")
  }
  if client.useCookies {
    t.Error("did not move back onto the plain route")
  }
  if client.userAgent() != plainUserAgentMangaFire {
    t.Errorf("image downloads would use %q", client.userAgent())
  }
}

// Stale cookies and the plain route refused too: the challenge is handed back unchanged
func TestBothRoutesRefusedIsReported(t *testing.T) {
  server := newRouteTestServerMangaFire(t, func(*http.Request) bool { return false })
  client := newRouteTestClientMangaFire(t, server, func() (string, string) { return "old", "1.ff" })
  client.useCookies = true
  client.clearance = "old"
  client.wafPass = "1.ff"

  _, _, err := client.apiGet("/titles/kwyvw", nil)
  if !isChallengedMangaFire(err) {
    t.Fatalf("got %v, want a challenge", err)
  }
  if len(server.requests) != 2 {
    t.Errorf("got %d requests, want 2 (cookies then plain)", len(server.requests))
  }
  if !client.useCookies {
    t.Error("left the cookie route even though the plain route was refused too")
  }
}
