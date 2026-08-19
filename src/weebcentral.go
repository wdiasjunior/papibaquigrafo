package src

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strconv"
  "strings"
  "sync"
  "time"
  "golang.org/x/net/html"
)

// browser-like per-host connection count; the site has a history of blocking
// more aggressive clients
const weebcentralConcurrency = 6

var weebcentralClient = &http.Client{
  Timeout: 60 * time.Second,
  Transport: &http.Transport{MaxIdleConnsPerHost: weebcentralConcurrency},
}

func weebcentral() DownloadResult {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaWeebcentral(mangaID)

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
      getChapterImagesWeebcentral(mangaTitle, chapter)
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

func getPageWeebcentral(_url string) (string, error) {
  var lastErr error
  for attempt := 1; attempt <= 3; attempt++ {
    req, err := http.NewRequest("GET", _url, nil)
    if err != nil {
      return "", err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Referer", "https://weebcentral.com/")

    resp, err := weebcentralClient.Do(req)
    if err != nil {
      lastErr = err
      continue
    }
    body, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
      lastErr = err
      continue
    }
    if resp.StatusCode != 200 {
      lastErr = fmt.Errorf("status %d for %s", resp.StatusCode, _url)
      continue
    }
    return string(body), nil
  }
  return "", lastErr
}

func getMangaWeebcentral(_mangaID string) (string, []string) {
  seriesBody, err := getPageWeebcentral(fmt.Sprintf("https://weebcentral.com/series/%s", _mangaID))
  if err != nil {
    fmt.Println("Could not get series page: ", err)
    return "", []string{}
  }

  var mangaTitle string
  var isInsideH1 = false
  tokenizer := html.NewTokenizer(strings.NewReader(seriesBody))
  titleLoop: for {
    switch tokenizer.Next() {
    case html.ErrorToken:
      break titleLoop
    case html.StartTagToken:
      if tokenizer.Token().Data == "h1" {
        isInsideH1 = true
      }
    case html.TextToken:
      if isInsideH1 {
        mangaTitle = strings.TrimSpace(string(tokenizer.Text()))
      }
    case html.EndTagToken:
      if tokenizer.Token().Data == "h1" {
        isInsideH1 = false
      }
    }
  }

  chapterListBody, err := getPageWeebcentral(fmt.Sprintf("https://weebcentral.com/series/%s/full-chapter-list", _mangaID))
  if err != nil {
    fmt.Println("Could not get chapter list: ", err)
    return mangaTitle, []string{}
  }

  targetClass := "bg-base-300"
  var chapterList = []string{}

  tokenizer = html.NewTokenizer(strings.NewReader(chapterListBody))
  loop: for {
    switch tokenizer.Next() {
    case html.ErrorToken:
      break loop
    case html.StartTagToken, html.SelfClosingTagToken:
      token := tokenizer.Token()
      if token.Data == "a" {
        for _, attr := range token.Attr {
          if attr.Key == "class" && strings.Contains(attr.Val, targetClass) {
            for _, attr := range token.Attr {
              if attr.Key == "href" {
                regex := regexp.MustCompile(`-page-\d+\.html$`)
                result := regex.ReplaceAllString(attr.Val, "")
                if !strings.HasPrefix(result, "https://") {
                  result = fmt.Sprintf("https://weebcentral.com%s", result)
                }
                chapterList = append(chapterList, result)
                break
              }
            }
          }
        }
      }
    }
  }

  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")

  reverseStringArray(chapterList)

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter))
  }

  return mangaTitle, chapterList
}

// The chapter-select button on the chapter page holds the chapter label,
// e.g. "Chapter 12" or "Episode 48".
func getChapterNumberWeebcentral(_chapterURL string) string {
  chapterBody, err := getPageWeebcentral(_chapterURL)
  if err != nil {
    fmt.Println("Could not get chapter page: ", err)
    return ""
  }

  var innerText string
  buttonDepth := 0
  tokenizer := html.NewTokenizer(strings.NewReader(chapterBody))
  loop: for {
    switch tokenizer.Next() {
    case html.ErrorToken:
      break loop
    case html.StartTagToken:
      token := tokenizer.Token()
      if buttonDepth > 0 {
        buttonDepth++
      } else if token.Data == "button" {
        for _, attr := range token.Attr {
          if attr.Key == "class" && strings.Contains(attr.Val, "col-span-3") {
            buttonDepth = 1
            break
          }
        }
      }
    case html.TextToken:
      if buttonDepth > 0 {
        innerText += string(tokenizer.Text())
      }
    case html.EndTagToken:
      if buttonDepth > 0 {
        buttonDepth--
        if buttonDepth == 0 {
          break loop
        }
      }
    }
  }

  regex := regexp.MustCompile(`\d+(\.\d+)?$`)
  return regex.FindString(strings.TrimSpace(innerText))
}

func getChapterImagesWeebcentral(_mangaTitle string, _mangaChapter string) {
  var urlImages string = fmt.Sprintf("%s%s", _mangaChapter, "/images?is_prev=False&current_page=1&reading_style=long_strip")

  body, err := getPageWeebcentral(urlImages)
  if err != nil {
    fmt.Println("Could not get chapter images: ", err)
    return
  }

  chapterNumber := getChapterNumberWeebcentral(_mangaChapter)
  if chapterNumber == "" {
    fmt.Println("Could not find chapter number, skipping: ", _mangaChapter)
    return
  }

  fmt.Println("Downloading chapter: ", chapterNumber)

  var chapterImagesList = []string{}
  imgTargetClass := "max-w-full"

  tokenizer := html.NewTokenizer(strings.NewReader(body))
  loop: for {
    switch tokenizer.Next() {
    case html.ErrorToken:
      break loop
    case html.StartTagToken, html.SelfClosingTagToken:
      token := tokenizer.Token()
      if token.Data == "img" {
        for _, attr := range token.Attr {
          if attr.Key == "class" && strings.Contains(attr.Val, imgTargetClass) {
            for _, attr := range token.Attr {
              if attr.Key == "src" {
                chapterImagesList = append(chapterImagesList, attr.Val)
                break
              }
            }
          }
        }
      }
    }
  }

  if len(chapterImagesList) == 0 {
    fmt.Println("No images found for chapter: ", chapterNumber)
    return
  }

  dir := fmt.Sprintf("%s%s/%s/Ch.%s", downloadsRoot, authorSubDir, _mangaTitle, chapterNumber)
  _dir := fsCreateDir(dir, false)

  var wg sync.WaitGroup
  semaphore := make(chan struct{}, weebcentralConcurrency)

  for i, chapterImageURL := range chapterImagesList {
    wg.Add(1)
    go func(i int, chapterImageURL string) {
      defer wg.Done()
      semaphore <- struct{}{}
      defer func() { <-semaphore }()

      var chapterImage []byte
      for attempt := 1; attempt <= 3; attempt++ {
        req, err := http.NewRequest("GET", chapterImageURL, nil)
        if err != nil {
          fmt.Println("Error creating request:", err)
          break
        }
        req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
        req.Header.Set("Referer", "https://weebcentral.com/")
        req.Header.Set("Accept", "image/*")

        resp, err := weebcentralClient.Do(req)
        if err != nil {
          fmt.Println("Request error:", err)
          continue
        }
        res, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
          fmt.Println("Request error. Retrying.")
          continue
        }
        if resp.StatusCode != 200 {
          fmt.Println("Request error:", fmt.Sprintf("status %d", resp.StatusCode))
          continue
        }
        chapterImage = res
        break
      }

      if chapterImage == nil {
        fmt.Println("Could not download image:", chapterImageURL)
        return
      }

      fsCreateFile(chapterImageURL, _dir, i + 1, chapterImage, false, "")
    }(i, chapterImageURL)
  }

  wg.Wait()
}
