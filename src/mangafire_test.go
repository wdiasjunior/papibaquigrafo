package src

import (
  "encoding/json"
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
