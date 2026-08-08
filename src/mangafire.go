package src

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "sort"
  "strconv"
  "strings"
  "time"
  "golang.org/x/net/html"
  "github.com/go-rod/rod"
  "github.com/go-rod/rod/lib/launcher"
)

func mangafire() DownloadResult {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  // Fetch English chapter list first to get the title and available languages
  mangaTitle, _, availableLanguages := getMangaMangaFire(mangaID, "en")

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

  // Fetch chapter list for the selected language
  _, chapterList, _ := getMangaMangaFire(mangaID, langCode)

  if len(chapterList) == 0 {
    fmt.Println("\nNo chapters available.")
    return downloadFailed(mangaTitle, "No chapters available.")
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
      getChapterImagesMangaFire(mangaTitle, chapter, langSuffix)
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
  ChapterLink string
  ChapterTitle string
}

type MangaFireLanguage struct {
  Code  string // uppercase from HTML: "EN", "JA", "ES"
  Title string // "English", "Japanese", "Spanish"
}

// AJAX response wrapper — the "result" field contains an HTML string
type mangaFireAjaxResponse struct {
  Result string `json:"result"`
}

// Page list response from /ajax/read/chapter/ — images is [[url, ?, offset], ...]
type mangaFirePageResponse struct {
  Result struct {
    Images [][]json.RawMessage `json:"images"`
  } `json:"result"`
}

func getMangaMangaFire(_mangaID string, _lang string) (string, []ChapterMangaFire, []MangaFireLanguage) {
  mangaURL := fmt.Sprintf("https://mangafire.to/manga/%s", _mangaID)

  client := &http.Client{}

  // 1. Fetch manga page via HTTP to get the title
  req, _ := http.NewRequest("GET", mangaURL, nil)
  req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
  req.Header.Set("Referer", "https://mangafire.to/")
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Could not fetch manga page:", err)
    return "", nil, nil
  }
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)

  // Parse title from <h1 itemprop="name">
  mangaTitle := parseMangaFireTitle(string(bodyBytes))

  // Parse available languages from the page
  availableLanguages := parseMangaFireLanguages(string(bodyBytes))

  // 2. Extract slug ID (e.g. "8wz3" from "kin-no-itoo.8wz3")
  slugParts := strings.Split(_mangaID, ".")
  slugID := slugParts[len(slugParts)-1]

  // 3. Fetch chapter list via AJAX endpoint (same as Tachiyomi: /ajax/manga/{id}/chapter/{lang})
  chapterAjaxURL := fmt.Sprintf("https://mangafire.to/ajax/manga/%s/chapter/%s", slugID, _lang)
  req2, _ := http.NewRequest("GET", chapterAjaxURL, nil)
  req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
  req2.Header.Set("Referer", mangaURL)
  req2.Header.Set("X-Requested-With", "XMLHttpRequest")
  resp2, err := client.Do(req2)
  if err != nil {
    fmt.Println("Could not fetch chapter list:", err)
    return mangaTitle, nil, availableLanguages
  }
  defer resp2.Body.Close()
  chapterBytes, _ := ioutil.ReadAll(resp2.Body)

  // 4. Parse JSON response — result field contains an HTML fragment with the chapter list
  var ajaxResp mangaFireAjaxResponse
  if err := json.Unmarshal(chapterBytes, &ajaxResp); err != nil {
    fmt.Println("Could not parse chapter list response:", err)
    return mangaTitle, nil, availableLanguages
  }

  // 5. Parse chapters from the HTML fragment
  chapterList := parseMangaFireChapters(ajaxResp.Result)

  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")

  // Sort by chapter number numerically (the AJAX response order doesn't match the page order)
  sort.Slice(chapterList, func(i, j int) bool {
    return extractMangaFireChapterNumber(chapterList[i].ChapterLink) < extractMangaFireChapterNumber(chapterList[j].ChapterLink)
  })

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter.ChapterTitle))
  }

  return mangaTitle, chapterList, availableLanguages
}

func parseMangaFireTitle(body string) string {
  reader := strings.NewReader(body)
  tokenizer := html.NewTokenizer(reader)
  isInsideH1 := false

  for {
    tokenType := tokenizer.Next()
    switch tokenType {
    case html.ErrorToken:
      return ""
    case html.StartTagToken:
      token := tokenizer.Token()
      if token.Data == "h1" {
        isInsideH1 = true
      }
    case html.TextToken:
      if isInsideH1 {
        text := strings.TrimSpace(string(tokenizer.Text()))
        if text != "" {
          return text
        }
      }
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "h1" {
        isInsideH1 = false
      }
    }
  }
}

func parseMangaFireLanguages(body string) []MangaFireLanguage {
  reader := strings.NewReader(body)
  tokenizer := html.NewTokenizer(reader)
  var languages []MangaFireLanguage
  seen := map[string]bool{}

  for {
    tokenType := tokenizer.Next()
    switch tokenType {
    case html.ErrorToken:
      return languages
    case html.StartTagToken:
      token := tokenizer.Token()
      if token.Data == "a" {
        isDropdownItem := false
        var code, title string
        for _, attr := range token.Attr {
          if attr.Key == "class" && strings.Contains(attr.Val, "dropdown-item") {
            isDropdownItem = true
          }
          if attr.Key == "data-code" {
            code = attr.Val
          }
          if attr.Key == "data-title" {
            title = attr.Val
          }
        }
        if isDropdownItem && code != "" && !seen[code] {
          seen[code] = true
          languages = append(languages, MangaFireLanguage{Code: code, Title: title})
        }
      }
    }
  }
}

func parseMangaFireChapters(htmlFragment string) []ChapterMangaFire {
  reader := strings.NewReader(htmlFragment)
  tokenizer := html.NewTokenizer(reader)
  var chapterList []ChapterMangaFire
  isInsideChapterLink := false
  isInsideFirstSpan := false
  spanCount := 0

  loop: for {
    tokenType := tokenizer.Next()
    switch tokenType {
    case html.ErrorToken:
      break loop
    case html.StartTagToken, html.SelfClosingTagToken:
      token := tokenizer.Token()
      if token.Data == "a" {
        for _, attr := range token.Attr {
          if attr.Key == "href" && strings.Contains(attr.Val, "/chapter-") {
            chapterList = append(chapterList, ChapterMangaFire{ChapterLink: attr.Val})
            isInsideChapterLink = true
            spanCount = 0
            break
          }
        }
      } else if token.Data == "span" && isInsideChapterLink {
        spanCount++
        if spanCount == 1 {
          isInsideFirstSpan = true
        }
      }
    case html.TextToken:
      text := strings.TrimSpace(string(tokenizer.Text()))
      if isInsideFirstSpan && text != "" {
        chapterList[len(chapterList)-1].ChapterTitle = text
      }
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "a" && isInsideChapterLink {
        isInsideChapterLink = false
        isInsideFirstSpan = false
        spanCount = 0
      } else if token.Data == "span" && isInsideFirstSpan {
        isInsideFirstSpan = false
      }
    }
  }

  return chapterList
}

// Extracts the chapter number as float from a href like "/read/kin-no-itoo.8wz3/en/chapter-10.5"
func extractMangaFireChapterNumber(href string) float64 {
  num, _ := strconv.ParseFloat(extractMangaFireChapterNumberStr(href), 64)
  return num
}

// Extracts the chapter number as string from a href like "/read/kin-no-itoo.8wz3/en/chapter-10.5"
func extractMangaFireChapterNumberStr(href string) string {
  idx := strings.LastIndex(href, "/chapter-")
  if idx == -1 {
    return "0"
  }
  return href[idx+len("/chapter-"):]
}

func getChapterImagesMangaFire(_mangaTitle string, _mangaChapter ChapterMangaFire, _langSuffix string) {
  chapterURL := fmt.Sprintf("https://mangafire.to%s", _mangaChapter.ChapterLink)

  l := launcher.New().
    Headless(true).
    Set("disable-blink-features", "AutomationControlled").
    MustLaunch()

  browser := rod.New().ControlURL(l).MustConnect()
  defer browser.MustClose()

  page := browser.MustPage("")

  // Hide headless detection signals and inject XHR interceptor before the page loads
  // (mirrors Tachiyomi's WebViewHelper that intercepts /ajax/read/chapter/ and /ajax/read/volume/)
  page.MustEvalOnNewDocument(`
    Object.defineProperty(navigator, 'webdriver', { get: () => false });
    window.__mf_data = '';
    (function() {
      var origOpen = XMLHttpRequest.prototype.open;
      XMLHttpRequest.prototype.open = function(method, url) {
        if (url && url.indexOf('/ajax/read/') !== -1) {
          this.addEventListener('load', function() {
            var text = this.responseText;
            // Only capture the page images response (has "images" field),
            // skip the chapter navigation panel response (has "html" field)
            if (text && text.indexOf('"images"') !== -1) {
              window.__mf_data = text;
            }
          });
        }
        return origOpen.apply(this, arguments);
      };
    })();
  `)

  page.MustNavigate(chapterURL)

  // Poll for the captured AJAX response (up to 30 seconds)
  var imagesJSON string
  for i := 0; i < 60; i++ {
    time.Sleep(500 * time.Millisecond)
    result := page.MustEval(`() => window.__mf_data || ''`)
    imagesJSON = result.String()
    if imagesJSON != "" {
      break
    }
  }

  if imagesJSON == "" {
    fmt.Println("Timed out waiting for image data for chapter:", _mangaChapter.ChapterTitle)
    return
  }

  // Parse the JSON response (same format as Tachiyomi's ResponseDto<PageListDto>)
  var pageResp mangaFirePageResponse
  if err := json.Unmarshal([]byte(imagesJSON), &pageResp); err != nil {
    fmt.Println("Error parsing image data:", err)
    return
  }

  var chapterImagesList []string
  for _, image := range pageResp.Result.Images {
    if len(image) > 0 {
      var imageURL string
      json.Unmarshal(image[0], &imageURL)
      if imageURL != "" {
        chapterImagesList = append(chapterImagesList, imageURL)
      }
    }
  }

  if len(chapterImagesList) == 0 {
    fmt.Println("No images found for chapter:", _mangaChapter.ChapterTitle)
    return
  }

  chapterNumber := extractMangaFireChapterNumberStr(_mangaChapter.ChapterLink)
  chapterFolder := fmt.Sprintf("Ch.%s", chapterNumber)

  fmt.Println("Downloading chapter:", chapterFolder)

  normalizedTitle := strings.ReplaceAll(_mangaTitle, ":", "-")
  var dir string
  if _langSuffix != "" {
    dir = fmt.Sprintf("%s/%s [%s]/%s", downloadsRoot, normalizedTitle, _langSuffix, chapterFolder)
  } else {
    dir = fmt.Sprintf("%s/%s/%s", downloadsRoot, normalizedTitle, chapterFolder)
  }
  _dir := fsCreateDir(dir, false)

  client := &http.Client{}

  for i, chapterImageURL := range chapterImagesList {
    var chapterImage []byte
    for {
      req, err := http.NewRequest("GET", chapterImageURL, nil)
      if err != nil {
        fmt.Println("Error creating request:", err)
        continue
      }
      req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
      req.Header.Set("Referer", "https://mangafire.to/")

      resp, err := client.Do(req)
      if err != nil {
        fmt.Println("Request error. Retrying.")
        time.Sleep(2 * time.Second)
        continue
      }
      defer resp.Body.Close()
      res, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        fmt.Println("Read error. Retrying.")
        time.Sleep(2 * time.Second)
        continue
      }
      if resp.StatusCode != 200 {
        fmt.Printf("HTTP %d. Retrying.\n", resp.StatusCode)
        time.Sleep(2 * time.Second)
        continue
      }
      if len(res) == 0 {
        fmt.Println("Empty response. Retrying.")
        time.Sleep(2 * time.Second)
        continue
      }
      chapterImage = res
      break
    }
    fsCreateFile(chapterImageURL, _dir, i + 1, chapterImage, false, "")
  }
}
