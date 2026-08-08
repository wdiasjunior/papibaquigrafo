package src

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "sort"
  "strconv"
  "strings"
  "time"
  "golang.org/x/net/html"
)

const baseURLMangabat = "https://www.mangabats.com"
const userAgentMangabat = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func mangabat() DownloadResult {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaMangabat(mangaID)

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

  downloaded := 0
  for i, chapter := range chapterList {
    if i >= firstChapter - 1 && i <= lastChapter - 1 {
      getChapterImagesMangabat(mangaTitle, chapter)
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

type ChapterMangabat struct {
  ChapterName string // "Chapter 525.1" - shown in the picker
  ChapterSlug string // "chapter-525-1" - URL path segment
  ChapterNum  string // "525.1" - display number, drives the Ch.<n> folder
  ChapterURL  string // https://www.mangabats.com/manga/{slug}/chapter-525-1
}

// GET https://www.mangabats.com/api/manga/{slug}/chapters?limit=&offset=
// The manga page no longer server-renders the chapter rows, it fills them in
// from this endpoint on the client. The endpoint pages at 50 by default.
type mangabatChaptersResponse struct {
  Success bool `json:"success"`
  Data    struct {
    Chapters []struct {
      ChapterName string      `json:"chapter_name"`
      ChapterSlug string      `json:"chapter_slug"`
      ChapterNum  json.Number `json:"chapter_num"`
    } `json:"chapters"`
    Pagination struct {
      Total   int  `json:"total"`
      Limit   int  `json:"limit"`
      Offset  int  `json:"offset"`
      HasMore bool `json:"has_more"`
    } `json:"pagination"`
  } `json:"data"`
}

const chapterPageSizeMangabat = 500

func httpGetMangabat(_url string, _referer string) ([]byte, error) {
  client := &http.Client{}

  req, err := http.NewRequest("GET", _url, nil)
  if err != nil {
    return nil, err
  }
  req.Header.Set("User-Agent", userAgentMangabat)
  req.Header.Set("Referer", _referer)

  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, _url)
  }

  return body, nil
}

func getMangaMangabat(_mangaID string) (string, []ChapterMangabat) {
  mangaURL := fmt.Sprintf("%s/manga/%s", baseURLMangabat, _mangaID)

  body, err := httpGetMangabat(mangaURL, fmt.Sprintf("%s/", baseURLMangabat))
  if err != nil {
    fmt.Println("Could not get manga page:", err)
    return "", nil
  }

  mangaTitle := getMangaTitleMangabat(string(body))
  if mangaTitle == "" {
    mangaTitle = _mangaID
  }

  chapterList := getChapterListMangabat(_mangaID, mangaURL)

  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")
  fmt.Println("Available chapters:")

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter.ChapterName))
  }

  return mangaTitle, chapterList
}

func getMangaTitleMangabat(_body string) string {
  reader := strings.NewReader(_body)
  tokenizer := html.NewTokenizer(reader)
  isInsideH1 := false

  loop: for {
    tokenType := tokenizer.Next()
    switch tokenType {
    case html.ErrorToken:
      break loop
    case html.StartTagToken:
      token := tokenizer.Token()
      if token.Data == "h1" {
        isInsideH1 = true
      }
    case html.TextToken:
      if isInsideH1 {
        text := strings.TrimSpace(string(tokenizer.Text()))
        if text != "" {
          // The title ends up as a directory name, so it can't carry separators
          text = strings.ReplaceAll(text, ":", "-")
          text = strings.ReplaceAll(text, "/", "-")
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

  return ""
}

func getChapterListMangabat(_mangaID string, _mangaURL string) []ChapterMangabat {
  var chapterList = []ChapterMangabat{}
  seen := map[string]bool{}
  offset := 0

  // The endpoint returns 50 chapters at a time unless a limit is given, so
  // follow has_more until the list is exhausted
  for {
    apiURL := fmt.Sprintf("%s/api/manga/%s/chapters?limit=%d&offset=%d", baseURLMangabat, _mangaID, chapterPageSizeMangabat, offset)

    body, err := httpGetMangabat(apiURL, _mangaURL)
    if err != nil {
      fmt.Println("Could not get chapter list:", err)
      break
    }

    var apiResp mangabatChaptersResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
      fmt.Println("Could not parse chapter list:", err)
      break
    }

    if !apiResp.Success || len(apiResp.Data.Chapters) == 0 {
      break
    }

    for _, chapter := range apiResp.Data.Chapters {
      if chapter.ChapterSlug == "" || seen[chapter.ChapterSlug] {
        continue
      }
      seen[chapter.ChapterSlug] = true

      // chapter_num is a JSON number decoded as json.Number, so the literal is
      // preserved verbatim - 40 stays "40" and 525.1 stays "525.1"
      chapterNum := chapter.ChapterNum.String()
      if chapterNum == "" {
        chapterNum = strings.TrimSpace(strings.TrimPrefix(chapter.ChapterName, "Chapter "))
      }
      if chapterNum == "" {
        // Last resort - "chapter-525-1" is ambiguous with "chapter-525-part-1"
        chapterNum = strings.ReplaceAll(strings.TrimPrefix(chapter.ChapterSlug, "chapter-"), "-", ".")
      }

      chapterList = append(chapterList, ChapterMangabat{
        ChapterName: chapter.ChapterName,
        ChapterSlug: chapter.ChapterSlug,
        ChapterNum:  chapterNum,
        ChapterURL:  fmt.Sprintf("%s/manga/%s/%s", baseURLMangabat, _mangaID, chapter.ChapterSlug),
      })
    }

    if !apiResp.Data.Pagination.HasMore {
      break
    }

    // Advance by what the server actually returned, not by what was asked for
    offset += len(apiResp.Data.Chapters)
  }

  if len(chapterList) == 0 {
    fmt.Println("No chapters found for:", _mangaID)
    return nil
  }

  // The API returns newest first, but the ordering is not guaranteed to be by
  // chapter number - it diverges from it as soon as an old chapter is backfilled
  sort.Slice(chapterList, func(i, j int) bool {
    a, _ := strconv.ParseFloat(chapterList[i].ChapterNum, 64)
    b, _ := strconv.ParseFloat(chapterList[j].ChapterNum, 64)
    return a < b
  })

  return chapterList
}

func getChapterImagesMangabat(_mangaTitle string, _mangaChapter ChapterMangabat) {
  body, err := httpGetMangabat(_mangaChapter.ChapterURL, fmt.Sprintf("%s/", baseURLMangabat))
  if err != nil {
    fmt.Println("Could not get chapter images:", err)
    return
  }

  imagePaths, cdnList := extractChapterImagesMangabat(string(body))

  if len(imagePaths) == 0 || len(cdnList) == 0 {
    fmt.Println("No images found for chapter:", _mangaChapter.ChapterName)
    return
  }

  chapterFolder := fmt.Sprintf("Ch.%s", _mangaChapter.ChapterNum)

  fmt.Println("Downloading chapter:", chapterFolder)

  dir := fmt.Sprintf("%s/%s/%s", downloadsRoot, _mangaTitle, chapterFolder)
  _dir := fsCreateDir(dir, false)

  client := &http.Client{}

  for i, imagePath := range imagePaths {
    var chapterImage []byte
    var lastURL string

    for attempt := 0; attempt < len(cdnList) * 3; attempt++ {
      chapterImageURL := fmt.Sprintf("%s%s", cdnList[attempt % len(cdnList)], imagePath)
      lastURL = chapterImageURL

      req, err := http.NewRequest("GET", chapterImageURL, nil)
      if err != nil {
        fmt.Println("Error creating request:", err)
        time.Sleep(2 * time.Second)
        continue
      }
      req.Header.Set("User-Agent", userAgentMangabat)
      // The CDN answers 403 without this
      req.Header.Set("Referer", fmt.Sprintf("%s/", baseURLMangabat))

      resp, err := client.Do(req)
      if err != nil {
        fmt.Println("Request error. Retrying.")
        time.Sleep(2 * time.Second)
        continue
      }
      res, readErr := ioutil.ReadAll(resp.Body)
      resp.Body.Close()

      if readErr != nil {
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

    if len(chapterImage) == 0 {
      fmt.Println("Skipping image after repeated failures:", lastURL)
      continue
    }

    fsCreateFile(imagePath, _dir, i + 1, chapterImage, false, "")
  }
}

// The chapter page carries the page list and the CDN hosts in a script block:
//   chapterImages = ["slug\/525.1\/0.webp", ...];
//   cdns = ["https:\/\/img-r1.2xstorage.com\/"];
// Returns the image paths and the CDN hosts separately so the download loop can
// rotate hosts.
func extractChapterImagesMangabat(_body string) ([]string, []string) {
  imagePaths := parseJSArrayMangabat(_body, "chapterImages")
  cdnList := parseJSArrayMangabat(_body, "cdns")

  var cdnHosts = []string{}
  for _, cdn := range cdnList {
    if cdn != "" {
      cdnHosts = append(cdnHosts, strings.TrimSuffix(cdn, "/") + "/")
    }
  }

  if len(imagePaths) > 0 && len(cdnHosts) > 0 {
    return imagePaths, cdnHosts
  }

  // Fallback - scrape the raw body. The reader img tags are written by
  // JavaScript with single quoted attributes, so the tokenizer never sees them.
  regex := regexp.MustCompile(`https://img-r[0-9]+\.2xstorage\.com/[^"'\s>\\]+\.(?:webp|jpg|jpeg|png)`)

  var fallbackPaths = []string{}
  var fallbackHosts = []string{}
  seenPaths := map[string]bool{}
  seenHosts := map[string]bool{}

  for _, imageURL := range regex.FindAllString(_body, -1) {
    // The sidebar carries ~20 cover thumbnails on the same hosts
    if strings.Contains(imageURL, "/thumb/") {
      continue
    }

    host := getHostAndReferer(imageURL)
    if host == "" {
      continue
    }
    path := strings.TrimPrefix(imageURL, host)

    if !seenPaths[path] {
      seenPaths[path] = true
      fallbackPaths = append(fallbackPaths, path)
    }
    if !seenHosts[host] {
      seenHosts[host] = true
      fallbackHosts = append(fallbackHosts, host)
    }
  }

  return fallbackPaths, fallbackHosts
}

// The arrays are valid JSON - "\/" is a legal escape - so let encoding/json
// handle the unescaping instead of doing it by hand.
func parseJSArrayMangabat(_body string, _varName string) []string {
  regex := regexp.MustCompile(_varName + `\s*=\s*(\[[^\]]*\])`)

  match := regex.FindStringSubmatch(_body)
  if len(match) < 2 {
    return nil
  }

  var values []string
  if err := json.Unmarshal([]byte(match[1]), &values); err != nil {
    fmt.Println("Could not parse", _varName + ":", err)
    return nil
  }

  return values
}
