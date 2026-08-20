package src

import (
  "encoding/base64"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "sort"
  "strconv"
  "strings"
  "time"
)

const baseURLMangaFire = "https://mangafire.to"

const chapterPageSizeMangaFire = 200

// The wait for Firefox to be pointed at the site again has no deadline, so
// say so again every now and then rather than looking hung
const clearanceNudgeMangaFire = 1 * time.Minute

var languageNamesMangaFire = map[string]string{
  "en": "English", "ja": "Japanese", "es": "Spanish", "es-la": "Spanish (LATAM)",
  "pt-br": "Portuguese (Br)", "fr": "French", "de": "German", "it": "Italian",
  "ru": "Russian", "pl": "Polish", "ar": "Arabic", "tr": "Turkish",
  "id": "Indonesian", "th": "Thai", "vi": "Vietnamese", "zh": "Chinese",
}

func mangafire() (result DownloadResult) {
  var mangaTitle string

  defer func() {
    if r := recover(); r != nil {
      fmt.Println("\nUnexpected error:", r)
      result = downloadFailed(mangaTitle, fmt.Sprintf("Unexpected error: %v", r))
    }
  }()

  fmt.Printf("\nEnter the Manga ID or URL: ")
  var mangaInput string
  fmt.Scanf("%s", &mangaInput)

  hid := hidMangaFire(mangaInput)
  if hid == "" {
    fmt.Println("\nInvalid manga ID.")
    return downloadFailed("", "Invalid manga ID.")
  }

  client, err := newClientMangaFire()
  if err != nil {
    fmt.Println("\n" + err.Error())
    return downloadFailed("", err.Error())
  }

  err = client.retryUntilCleared(func() error {
    var attemptErr error
    mangaTitle, attemptErr = getMangaTitleMangaFire(client, hid)
    return attemptErr
  })
  if err != nil {
    fmt.Println("\n" + describeFailureMangaFire(err))
    return downloadFailed("", describeFailureMangaFire(err))
  }

  fmt.Println("")
  fmt.Println(mangaTitle)

  var chapterList []ChapterMangaFire
  var availableLanguages []MangaFireLanguage
  err = client.retryUntilCleared(func() error {
    var attemptErr error
    chapterList, availableLanguages, attemptErr = getChaptersMangaFire(client, hid, "")
    return attemptErr
  })
  if err != nil {
    fmt.Println("\n" + describeFailureMangaFire(err))
    return downloadFailed(mangaTitle, describeFailureMangaFire(err))
  }

  if len(availableLanguages) == 0 {
    fmt.Println("\nNo languages available.")
    return downloadFailed(mangaTitle, "No languages available.")
  }

  // Show available languages and let the user pick
  fmt.Println("\nAvailable languages:")
  for i, lang := range availableLanguages {
    fmt.Println(fmt.Sprintf("%d - %s (%s)", i + 1, lang.Title, lang.Code))
  }

  fmt.Printf("\nSelect language: ")
  var langInput string
  fmt.Scanf("%s", &langInput)
  langIndex, _ := strconv.Atoi(langInput)
  if langIndex < 1 || langIndex > len(availableLanguages) {
    fmt.Println("Invalid selection.")
    return downloadFailed(mangaTitle, "Invalid language selection.")
  }
  selectedLang := availableLanguages[langIndex - 1]
  langCode := strings.ToLower(selectedLang.Code)

  // The first call already brought every language back, so only go out again
  // when it could not all fit on one page
  chapterList = filterChaptersByLangMangaFire(chapterList, langCode)
  if len(chapterList) == 0 {
    err = client.retryUntilCleared(func() error {
      var attemptErr error
      chapterList, _, attemptErr = getChaptersMangaFire(client, hid, langCode)
      return attemptErr
    })
    if err != nil {
      fmt.Println("\nCould not get the chapter list:", err)
      return downloadFailed(mangaTitle, describeFailureMangaFire(err))
    }
  }

  if len(chapterList) == 0 {
    fmt.Println("\nNo chapters available.")
    return downloadFailed(mangaTitle, "No chapters available.")
  }

  fmt.Println("")
  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter.ChapterTitle))
  }

  fmt.Println("\nEnter the range of chapters you want to download.")

  fmt.Printf("\nInitial chapter: ")
  var userInputFirstChapter string
  fmt.Scanf("%s", &userInputFirstChapter)
  firstChapter, _ := strconv.Atoi(userInputFirstChapter)

  fmt.Printf("\nLast chapter: ")
  var userInputLastChapter string
  fmt.Scanf("%s", &userInputLastChapter)
  lastChapter, _ := strconv.Atoi(userInputLastChapter)
  fmt.Printf("\n")

  // Use language suffix in folder name only for non-English
  langSuffix := ""
  if selectedLang.Code != "EN" {
    langSuffix = selectedLang.Code
  }

  downloaded := 0

  for i, chapter := range chapterList {
    if i >= firstChapter - 1 && i <= lastChapter - 1 {
      // A long run easily outlives the cookies, and anything else can fail
      // transiently too. Stay on this chapter until it comes down rather than
      // running down the rest of the list leaving holes behind for the user
      // to find and fetch again by hand.
      for {
        err := client.retryUntilCleared(func() error {
          return getChapterImagesMangaFire(client, mangaTitle, chapter, langSuffix)
        })
        if err == nil {
          break
        }
        fmt.Println("Could not download", chapter.ChapterTitle, "- retrying -", describeFailureMangaFire(err))
        time.Sleep(2 * time.Second)
      }
      downloaded++
    }
  }

  if downloaded == 0 {
    fmt.Printf("\nNo chapters in the selected range.\n")
    return downloadFailed(mangaTitle, "No chapters in the selected range.")
  }

  fmt.Printf("\nDownload completed!\n")

  return downloadSuccess(mangaTitle, downloaded)
}

type ChapterMangaFire struct {
  ChapterID     int     // /api/chapters/{id} returns the page list
  ChapterNumber float64 // for sorting
  ChapterNumStr string  // "10.5", drives the Ch.<n> folder name
  ChapterTitle  string
  ChapterLang   string
}

type MangaFireLanguage struct {
  Code  string // uppercase for display: "EN", "ES-LA"
  Title string // "English", "Spanish (LATAM)"
}

type mangaFireDetailsResponse struct {
  Data struct {
    Title string `json:"title"`
  } `json:"data"`
}

type mangaFireChaptersResponse struct {
  Items []struct {
    ID       int         `json:"id"`
    Number   json.Number `json:"number"`
    Name     string      `json:"name"`
    Language string      `json:"language"`
  } `json:"items"`
  Meta struct {
    LastPage int `json:"lastPage"`
  } `json:"meta"`
}

type mangaFirePagesResponse struct {
  Data struct {
    Pages []struct {
      URL string `json:"url"`
    } `json:"pages"`
  } `json:"data"`
}

////////////////////////////////////////////////////////////////////////////////
// request signing
////////////////////////////////////////////////////////////////////////////////

// Every /api/ request carries a "vrf" signature of its own path and query.
// Three rounds of a chained substitution cipher, url safe base64, no padding.
const vrfTable1MangaFire = "yINlmUNho8VYJT+ibTIP+9ESiULpVEtMOoD6U6lRE0R/xwXo/Xp9NrUgC4cw/Lmo33vUyjUE40kUoEWIr/fxfNNcq2s79ShQ5NhNrFnJ4hXPwOu/SuXzIbuTQKGFvfm08E9jvCfqAtoDqvQq3dVWPQFmJjgvkISBeXY3BgANR+yVnjGbcxZ47d6kLNfZPIayTq3/YGySb1KuVZodWp/WGNAO5pfMcpaK53Hhs0allBszaMaxuouOwdxbwgxIw6YunSsXjI05Yi0j9j4eHKfSXR8Ifo/Od+8iamRfCXTyvm7NGRGYdcQ0ywcK/u6RXhrbcCm4t2eCtrDgQVecJGkQ+A=="
const vrfKey1MangaFire = "0Ec58JOY3uBzJK9m3zqIOpdlF7UFiax9DmA="
const vrfTable2MangaFire = "IUFltCxD3Oc2cwCgkJffthaOg9cgPUb0LgW6H/VtfcF0kc5F25t+aWj6JH9VOhOaY0rAFdUxlDnl5BLNvwEJvQtP5qcw7vdb/K+chnbwnspSHT8mz5lqwz41TezG0hkO06FTjJZhsyNuFLDpD2ZZxQj/QIRcF90zpmQ7Byu483WsQqUE0C342HL+JXngRB6fRzxRyVTaKu83h7UYTJ0QMt6ixFh6S3F8gqkKwrGTL3jHNBsD45UnifK8+RGtishQV2K3rujLKEkiZxpr2dYcudFW4oFsDKhad3CLBvuyTqsCo4B7mL5IKQ1vXo/MOOvq1I1d8ar9X6Ttu5KF4fZgiA=="
const vrfKey2MangaFire = "AAdjb1iPY8CiDmq9H34tKTBF8a3oDQ=="
const vrfTable3MangaFire = "NQHlu1/wVO5EmkwQymF810qqY2xG1k2obcas4Z9mCsPEIFl9pRIjFxbJ7ybMHbBckT5Ton85E0FOeHezbh/mjlEYpmpnlXOS8dgrqeq2KfxImTh1YK9y0PeMNhzA1OQzSY9brYOJq/l2QnE/hwOeZIhPixVSKIUlDb5vLcH6RWKxkIEMuP0bDwIqQ71AJJaEaMJL7A6YtyIwoRT+L5v4aZzodN/0+3nOGsfblFjgxSfPzVDjNFeNl5P26+kEC/8AHgdrpAbt3hHz3HrRN1Y6e+JHgF7ncFWnoF0y3THL1S71WgWGCa6KtSzTCCG58n68nTyj2T3Sshk7utqCtMi/ZQ=="
const vrfKey3MangaFire = "DELOJgPsVaCcblDtTGMdHzM="

type vrfStageMangaFire struct {
  table []byte
  key   []byte
  iv    int
}

var vrfStagesMangaFire = []vrfStageMangaFire{
  {decodeBase64MangaFire(vrfTable1MangaFire), decodeBase64MangaFire(vrfKey1MangaFire), 0x5A},
  {decodeBase64MangaFire(vrfTable2MangaFire), decodeBase64MangaFire(vrfKey2MangaFire), 0x35},
  {decodeBase64MangaFire(vrfTable3MangaFire), decodeBase64MangaFire(vrfKey3MangaFire), 0xBA},
}

func decodeBase64MangaFire(_encoded string) []byte {
  decoded, err := base64.StdEncoding.DecodeString(_encoded)
  if err != nil {
    return nil
  }
  return decoded
}

func signVrfMangaFire(_signPath string) string {
  data := []byte(_signPath)

  for _, stage := range vrfStagesMangaFire {
    data = vrfEncryptMangaFire(data, stage)
  }

  return base64.RawURLEncoding.EncodeToString(data)
}

func vrfEncryptMangaFire(_data []byte, _stage vrfStageMangaFire) []byte {
  out := make([]byte, len(_data))
  prev := _stage.iv

  for i := 0; i < len(_data); i++ {
    prev = int(_stage.table[(int(_data[i]) ^ int(_stage.key[i % len(_stage.key)]) ^ prev) & 0xFF])
    out[i] = byte(prev)
  }

  return out
}

// The signature covers the path without the /api prefix and the query sorted by
// key, so the request has to send the params back in that same order
func apiURLMangaFire(_path string, _params [][2]string) string {
  params := make([][2]string, len(_params))
  copy(params, _params)
  sort.SliceStable(params, func(i, j int) bool {
    return params[i][0] < params[j][0]
  })

  var pairs []string
  for _, param := range params {
    pairs = append(pairs, fmt.Sprintf("%s=%s", param[0], param[1]))
  }
  query := strings.Join(pairs, "&")

  signPath := _path
  if query != "" {
    signPath = fmt.Sprintf("%s?%s", _path, query)
  }

  signature := signVrfMangaFire(signPath)

  if query != "" {
    return fmt.Sprintf("%s/api%s?%s&vrf=%s", baseURLMangaFire, _path, query, signature)
  }
  return fmt.Sprintf("%s/api%s?vrf=%s", baseURLMangaFire, _path, signature)
}

func vrfReadyMangaFire() bool {
  for _, stage := range vrfStagesMangaFire {
    if len(stage.table) != 256 || len(stage.key) == 0 {
      return false
    }
  }
  return true
}

////////////////////////////////////////////////////////////////////////////////
// http
////////////////////////////////////////////////////////////////////////////////

// mangafire.to sits behind a Cloudflare managed challenge. It cannot be passed
// from here - Cloudflare rejects the devtools protocol a headless browser is
// driven with, so even a real Chrome sits on "Just a moment..." forever. What
// does work is the clearance Firefox already earned by browsing the site
// normally: borrow those cookies and the api answers plain http requests.
//
// Two cookies are needed, not one. cf_clearance is the long lived proof the
// challenge was passed, but the site also puts Cloudflare in front of anything
// missing its own waf_pass cookie, so cf_clearance on its own gets a "Just a
// moment..." page back. waf_pass is short lived (~30 minutes) and Firefox is
// handed a new one whenever the site is opened.
type mangaFireClient struct {
  http      *http.Client
  clearance string
  wafPass   string
  userAgent string
}

func newClientMangaFire() (*mangaFireClient, error) {
  if !vrfReadyMangaFire() {
    return nil, fmt.Errorf("Request signing is broken - the built in vrf tables did not decode.")
  }

  clearance, wafPass := firefoxCookiesMangaFire()
  if override := os.Getenv("PAPIBAQUIGRAFO_MF_CLEARANCE"); override != "" {
    clearance = override
  }
  if override := os.Getenv("PAPIBAQUIGRAFO_MF_WAF_PASS"); override != "" {
    wafPass = override
  }
  if clearance == "" || wafPass == "" {
    return nil, fmt.Errorf("No Cloudflare clearance found. Open %s in Firefox, let the check pass, then run this again.", baseURLMangaFire)
  }

  userAgent := os.Getenv("PAPIBAQUIGRAFO_MF_UA")
  if userAgent == "" {
    userAgent = firefoxUserAgentMangaFire()
  }

  return &mangaFireClient{
    http:      &http.Client{Timeout: 60 * time.Second},
    clearance: clearance,
    wafPass:   wafPass,
    userAgent: userAgent,
  }, nil
}

// waf_pass only stays valid for half an hour or so, and a download can easily
// outlive it, so a 403 is worth one look for fresher cookies before giving up
func (c *mangaFireClient) apiGet(_path string, _params [][2]string) ([]byte, int, error) {
  body, status, err := c.get(_path, _params)
  if status == 403 && c.refreshClearance() {
    return c.get(_path, _params)
  }

  return body, status, err
}

// Either cookie moving on is worth another try - cf_clearance is usually the
// one that stays put for months while waf_pass turns over every visit
func (c *mangaFireClient) refreshClearance() bool {
  clearance, wafPass := firefoxCookiesMangaFire()
  if clearance == "" || wafPass == "" {
    return false
  }
  if clearance == c.clearance && wafPass == c.wafPass {
    return false
  }

  fmt.Println("Picked up newer Cloudflare cookies from Firefox.")
  c.clearance = clearance
  c.wafPass = wafPass

  return true
}

// waitForClearance parks until Firefox has been pointed at the site again and
// a different cookie shows up, so the run can carry on where it left off.
// There is deliberately no deadline: giving up only moves the problem onto the
// user, who then has to work out which chapters were skipped and fetch them by
// hand. Sitting on the same chapter until the cookies come back is cheaper.
func (c *mangaFireClient) waitForClearance() {
  fmt.Printf("\nThe Cloudflare cookies have gone stale.\n")
  fmt.Printf("Open %s in Firefox and let the page finish loading - this picks the new cookies up on its own.\n", baseURLMangaFire)
  fmt.Printf("Waiting - nothing is skipped, this carries on where it left off. Ctrl-C to stop.\n")

  nudge := time.Now().Add(clearanceNudgeMangaFire)
  for {
    time.Sleep(3 * time.Second)
    if c.refreshClearance() {
      return
    }

    if time.Now().After(nudge) {
      fmt.Printf("Still waiting on %s in Firefox...\n", baseURLMangaFire)
      nudge = time.Now().Add(clearanceNudgeMangaFire)
    }
  }
}

// retryUntilCleared runs _attempt over and over, parking for fresh
// cookies between tries, until it comes back with something other than a
// challenge. Anything that is not a challenge is handed straight back.
func (c *mangaFireClient) retryUntilCleared(_attempt func() error) error {
  err := _attempt()
  for isChallengedMangaFire(err) {
    c.waitForClearance()
    err = _attempt()
  }

  return err
}

func (c *mangaFireClient) get(_path string, _params [][2]string) ([]byte, int, error) {
  requestURL := apiURLMangaFire(_path, _params)

  req, err := http.NewRequest("GET", requestURL, nil)
  if err != nil {
    return nil, 0, err
  }
  // The clearance is tied to the user agent that earned it, so these two have
  // to travel together, and waf_pass has to come along or Cloudflare answers
  // with a challenge page no matter how fresh the clearance is
  req.Header.Set("User-Agent", c.userAgent)
  req.Header.Set("Accept", "application/json")
  req.Header.Set("Referer", fmt.Sprintf("%s/", baseURLMangaFire))
  req.Header.Set("Cookie", fmt.Sprintf("cf_clearance=%s; waf_pass=%s", c.clearance, c.wafPass))

  resp, err := c.http.Do(req)
  if err != nil {
    return nil, 0, err
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, resp.StatusCode, err
  }

  if resp.StatusCode != 200 {
    return nil, resp.StatusCode, fmt.Errorf("HTTP %d for %s: %s", resp.StatusCode, _path, truncateMangaFire(string(body), 200))
  }
  if !strings.HasPrefix(strings.TrimSpace(string(body)), "{") {
    return nil, resp.StatusCode, fmt.Errorf("unexpected non-JSON response for %s: %s", _path, truncateMangaFire(string(body), 200))
  }

  return body, resp.StatusCode, nil
}

func isChallengedMangaFire(_err error) bool {
  return _err != nil && strings.Contains(_err.Error(), "HTTP 403")
}

// A 403 here always means the same thing, so say so instead of dumping the
// whole challenge page at the user
func describeFailureMangaFire(_err error) string {
  if isChallengedMangaFire(_err) {
    return fmt.Sprintf("Cloudflare clearance expired. Open %s in Firefox, let the page load, then run this again.", baseURLMangaFire)
  }

  return fmt.Sprintf("%v", _err)
}

////////////////////////////////////////////////////////////////////////////////
// firefox clearance
////////////////////////////////////////////////////////////////////////////////

// Firefox keeps its cookies in plain text inside a sqlite file. Rather than
// pull in a sqlite driver for one query, shell out to the sqlite3 binary.
// Both cookies have to come from the same profile - pairing a clearance with
// another profile's waf_pass gets the challenge page back.
func firefoxCookiesMangaFire() (string, string) {
  if _, err := exec.LookPath("sqlite3"); err != nil {
    fmt.Println("sqlite3 is not installed, so the Firefox cookies cannot be read.")
    return "", ""
  }

  home, err := os.UserHomeDir()
  if err != nil {
    return "", ""
  }

  profiles, _ := filepath.Glob(filepath.Join(home, ".mozilla", "firefox", "*", "cookies.sqlite"))
  if len(profiles) == 0 {
    return "", ""
  }

  // Newest profile first - the one being browsed with is the one with the
  // freshest cookies
  sort.Slice(profiles, func(i, j int) bool {
    return modTimeMangaFire(profiles[i]).After(modTimeMangaFire(profiles[j]))
  })

  for _, profile := range profiles {
    if clearance, wafPass := readFirefoxCookiesMangaFire(profile); clearance != "" && wafPass != "" {
      return clearance, wafPass
    }
  }

  return "", ""
}

func modTimeMangaFire(_path string) time.Time {
  info, err := os.Stat(_path)
  if err != nil {
    return time.Time{}
  }
  return info.ModTime()
}

func readFirefoxCookiesMangaFire(_cookiesPath string) (string, string) {
  // Firefox journals in wal mode and holds the newest cookies there, so the
  // sidecar has to come along or the values read back empty
  tempDir, err := ioutil.TempDir("", "papibaquigrafo")
  if err != nil {
    return "", ""
  }
  defer os.RemoveAll(tempDir)

  copyPath := filepath.Join(tempDir, "cookies.sqlite")
  if err := copyFileMangaFire(_cookiesPath, copyPath); err != nil {
    return "", ""
  }
  copyFileMangaFire(_cookiesPath + "-wal", copyPath + "-wal")

  // Oldest first so a later row for the same name overwrites an earlier one,
  // leaving the most recently used value of each
  query := "select name, value from moz_cookies where host like '%mangafire%' and name in ('cf_clearance', 'waf_pass') and value != '' order by lastAccessed asc;"
  out, err := exec.Command("sqlite3", copyPath, query).Output()
  if err != nil {
    return "", ""
  }

  var clearance, wafPass string
  for _, line := range strings.Split(string(out), "\n") {
    parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
    if len(parts) != 2 {
      continue
    }

    switch parts[0] {
    case "cf_clearance":
      clearance = parts[1]
    case "waf_pass":
      wafPass = parts[1]
    }
  }

  return clearance, wafPass
}

func copyFileMangaFire(_from string, _to string) error {
  contents, err := ioutil.ReadFile(_from)
  if err != nil {
    return err
  }

  return ioutil.WriteFile(_to, contents, 0600)
}

func firefoxUserAgentMangaFire() string {
  version := "153"

  if out, err := exec.Command("firefox", "--version").Output(); err == nil {
    fields := strings.Fields(string(out))
    if len(fields) > 0 {
      major := strings.SplitN(fields[len(fields)-1], ".", 2)[0]
      if _, err := strconv.Atoi(major); err == nil {
        version = major
      }
    }
  }

  return fmt.Sprintf("Mozilla/5.0 (X11; Linux x86_64; rv:%s.0) Gecko/20100101 Firefox/%s.0", version, version)
}

////////////////////////////////////////////////////////////////////////////////
// manga
////////////////////////////////////////////////////////////////////////////////

// hidMangaFire pulls the short id out of a pasted url or a bare id. It used to
// trail the slug as "kin-no-itoo.8wz3" and now leads it as "kwyvw-nikaidou-kou",
// so handle both rather than betting on either.
func hidMangaFire(_input string) string {
  input := strings.TrimSpace(_input)
  if input == "" {
    return ""
  }

  input = strings.TrimSuffix(input, "/")
  if idx := strings.LastIndex(input, "/"); idx != -1 {
    input = input[idx + 1:]
  }
  // a pasted url can still carry a query or fragment
  if idx := strings.IndexAny(input, "?#"); idx != -1 {
    input = input[:idx]
  }

  if idx := strings.LastIndex(input, "."); idx != -1 {
    return input[idx + 1:]
  }
  if idx := strings.Index(input, "-"); idx != -1 {
    return input[:idx]
  }

  return input
}

func getMangaTitleMangaFire(_client *mangaFireClient, _hid string) (string, error) {
  body, status, err := _client.apiGet(fmt.Sprintf("/titles/%s", _hid), nil)
  if err != nil {
    if status == 404 {
      return "", fmt.Errorf("no manga found with id %s", _hid)
    }
    return "", err
  }

  var details mangaFireDetailsResponse
  if err := json.Unmarshal(body, &details); err != nil {
    return "", err
  }
  if details.Data.Title == "" {
    return _hid, nil
  }

  return details.Data.Title, nil
}

func newLanguageMangaFire(_code string) MangaFireLanguage {
  title := languageNamesMangaFire[_code]
  if title == "" {
    title = strings.ToUpper(_code)
  }

  return MangaFireLanguage{Code: strings.ToUpper(_code), Title: title}
}

////////////////////////////////////////////////////////////////////////////////
// chapters
////////////////////////////////////////////////////////////////////////////////

// Leaving the language out returns every language at once, which is where the
// list of available ones comes from - nothing else in the api reports them
func getChaptersMangaFire(_client *mangaFireClient, _hid string, _lang string) ([]ChapterMangaFire, []MangaFireLanguage, error) {
  var chapterList []ChapterMangaFire
  var languages []MangaFireLanguage
  seen := map[string]bool{}

  lastPage := 1
  for page := 1; page <= lastPage; page++ {
    if page > 1 {
      time.Sleep(500 * time.Millisecond)
    }

    params := [][2]string{
      {"limit", strconv.Itoa(chapterPageSizeMangaFire)},
      {"order", "desc"},
      {"page", strconv.Itoa(page)},
      {"sort", "number"},
    }
    if _lang != "" {
      params = append(params, [2]string{"language", _lang})
    }

    body, _, err := _client.apiGet(fmt.Sprintf("/titles/%s/chapters", _hid), params)
    if err != nil {
      return nil, nil, err
    }

    var chapters mangaFireChaptersResponse
    if err := json.Unmarshal(body, &chapters); err != nil {
      return nil, nil, err
    }

    for _, item := range chapters.Items {
      numStr := trimChapterNumberMangaFire(item.Number.String())
      number, _ := strconv.ParseFloat(numStr, 64)

      title := fmt.Sprintf("Ch. %s", numStr)
      if item.Name != "" {
        title = fmt.Sprintf("Ch. %s - %s", numStr, item.Name)
      }

      language := strings.ToLower(item.Language)
      if language != "" && !seen[language] {
        seen[language] = true
        languages = append(languages, newLanguageMangaFire(language))
      }

      chapterList = append(chapterList, ChapterMangaFire{
        ChapterID:     item.ID,
        ChapterNumber: number,
        ChapterNumStr: numStr,
        ChapterTitle:  title,
        ChapterLang:   language,
      })
    }

    if page == 1 && chapters.Meta.LastPage > lastPage {
      lastPage = chapters.Meta.LastPage
    }
  }

  sort.Slice(languages, func(i, j int) bool {
    return languages[i].Code < languages[j].Code
  })

  return sortChaptersMangaFire(chapterList), languages, nil
}

// The api hands them back newest first
func sortChaptersMangaFire(_chapterList []ChapterMangaFire) []ChapterMangaFire {
  sort.Slice(_chapterList, func(i, j int) bool {
    return _chapterList[i].ChapterNumber < _chapterList[j].ChapterNumber
  })

  return _chapterList
}

func filterChaptersByLangMangaFire(_chapterList []ChapterMangaFire, _lang string) []ChapterMangaFire {
  var filtered []ChapterMangaFire

  for _, chapter := range _chapterList {
    if chapter.ChapterLang == _lang {
      filtered = append(filtered, chapter)
    }
  }

  return filtered
}

func trimChapterNumberMangaFire(_number string) string {
  number := strings.TrimSuffix(_number, ".0")
  if number == "" {
    return "0"
  }

  return number
}

////////////////////////////////////////////////////////////////////////////////
// images
////////////////////////////////////////////////////////////////////////////////

func getChapterImagesMangaFire(_client *mangaFireClient, _mangaTitle string, _mangaChapter ChapterMangaFire, _langSuffix string) error {
  // One request returns every page - the reader itself lazy loads them, which
  // is why scraping the page only ever yielded the handful already on screen
  body, _, err := _client.apiGet(fmt.Sprintf("/chapters/%d", _mangaChapter.ChapterID), nil)
  if err != nil {
    return err
  }

  var pages mangaFirePagesResponse
  if err := json.Unmarshal(body, &pages); err != nil {
    return err
  }

  var chapterImagesList []string
  for _, page := range pages.Data.Pages {
    if page.URL != "" {
      chapterImagesList = append(chapterImagesList, page.URL)
    }
  }

  if len(chapterImagesList) == 0 {
    return fmt.Errorf("no images found")
  }

  chapterFolder := fmt.Sprintf("Ch.%s", _mangaChapter.ChapterNumStr)

  fmt.Println("Downloading chapter:", chapterFolder)

  normalizedTitle := strings.ReplaceAll(_mangaTitle, ":", "-")
  var dir string
  if _langSuffix != "" {
    dir = fmt.Sprintf("%s%s/%s [%s]/%s", downloadsRoot, authorSubDir, normalizedTitle, _langSuffix, chapterFolder)
  } else {
    dir = fmt.Sprintf("%s%s/%s/%s", downloadsRoot, authorSubDir, normalizedTitle, chapterFolder)
  }
  _dir := fsCreateDir(dir, false)

  for i, chapterImageURL := range chapterImagesList {
    chapterImage := downloadImageMangaFire(_client, chapterImageURL)
    fsCreateFile(chapterImageURL, _dir, i + 1, chapterImage, false, "")
  }

  return nil
}

// The images sit on a cdn that is not behind the challenge. Retries have no
// cap - a skipped page leaves a hole in the chapter, which is worse than
// waiting for the cdn to come back
func downloadImageMangaFire(_client *mangaFireClient, _imageURL string) []byte {
  for attempt := 0; ; attempt++ {
    if attempt > 0 {
      time.Sleep(2 * time.Second)
    }

    req, err := http.NewRequest("GET", _imageURL, nil)
    if err != nil {
      fmt.Println("Error creating request:", err)
      continue
    }
    req.Header.Set("User-Agent", _client.userAgent)
    req.Header.Set("Referer", fmt.Sprintf("%s/", baseURLMangaFire))

    resp, err := _client.http.Do(req)
    if err != nil {
      fmt.Println("Request error. Retrying.")
      continue
    }

    res, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
      fmt.Println("Read error. Retrying.")
      continue
    }
    if resp.StatusCode != 200 {
      fmt.Printf("HTTP %d. Retrying.\n", resp.StatusCode)
      continue
    }
    if len(res) == 0 {
      fmt.Println("Empty response. Retrying.")
      continue
    }

    return res
  }
}

func truncateMangaFire(_body string, _max int) string {
  body := strings.Join(strings.Fields(_body), " ")
  if len(body) <= _max {
    return body
  }

  return body[:_max] + "..."
}
