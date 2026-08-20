package src

import (
  "bytes"
  "context"
  "fmt"
  "io"
  "net"
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

// the image hosts throttle single connections down to KB/s, so a transfer may
// legitimately take minutes — never cap total time, only abort when no bytes
// arrive for this long
var weebcentralStallTimeout = 30 * time.Second

var weebcentralClient = &http.Client{
  Transport: &http.Transport{
    MaxIdleConnsPerHost: weebcentralConcurrency,
    DialContext: (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
    TLSHandshakeTimeout: 15 * time.Second,
    ResponseHeaderTimeout: 30 * time.Second,
  },
}

func weebcentral() DownloadResult {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList, chapterLabels := getMangaWeebcentral(mangaID)

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
      getChapterImagesWeebcentral(mangaTitle, chapter, chapterLabels[i])
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

// retries until it succeeds — a skipped page leaves a hole in the chapter,
// which is worse than waiting
func fetchWeebcentral(_url string, _accept string) ([]byte, error) {
  rateLimitWait := 5 * time.Second
  for {
    ctx, cancel := context.WithCancel(context.Background())
    req, err := http.NewRequestWithContext(ctx, "GET", _url, nil)
    if err != nil {
      cancel()
      return nil, err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Referer", "https://weebcentral.com/")
    if _accept != "" {
      req.Header.Set("Accept", _accept)
    }

    resp, err := weebcentralClient.Do(req)
    if err != nil {
      cancel()
      fmt.Println("Request error, retrying:", _url, "-", err)
      continue
    }
    // rate-limit waits don't count as attempts; keep going until the limiter clears
    if resp.StatusCode == 429 {
      resp.Body.Close()
      cancel()
      wait := rateLimitWait
      if retryAfter, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
        wait = time.Duration(retryAfter) * time.Second
      }
      fmt.Println("Rate limited, waiting", wait)
      time.Sleep(wait)
      if rateLimitWait < 60 * time.Second {
        rateLimitWait *= 2
      }
      continue
    }
    if resp.StatusCode != 200 {
      resp.Body.Close()
      cancel()
      fmt.Println("Request error, retrying:", _url, "- status", resp.StatusCode)
      continue
    }

    // a slow transfer is fine as long as bytes keep arriving; the watchdog
    // cancels the request only after weebcentralStallTimeout with no progress
    watchdog := time.AfterFunc(weebcentralStallTimeout, cancel)
    var buf bytes.Buffer
    chunk := make([]byte, 32 * 1024)
    var readErr error
    for {
      n, err := resp.Body.Read(chunk)
      if n > 0 {
        watchdog.Reset(weebcentralStallTimeout)
        buf.Write(chunk[:n])
      }
      if err == io.EOF {
        break
      }
      if err != nil {
        readErr = err
        break
      }
    }
    watchdog.Stop()
    resp.Body.Close()
    cancel()
    if readErr != nil {
      fmt.Println("Transfer stalled, retrying:", _url, "-", readErr)
      continue
    }
    return buf.Bytes(), nil
  }
}

func getPageWeebcentral(_url string) (string, error) {
  body, err := fetchWeebcentral(_url, "")
  return string(body), err
}

func getMangaWeebcentral(_mangaID string) (string, []string, []string) {
  seriesBody, err := getPageWeebcentral(fmt.Sprintf("https://weebcentral.com/series/%s", _mangaID))
  if err != nil {
    fmt.Println("Could not get series page: ", err)
    return "", []string{}, []string{}
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
    return mangaTitle, []string{}, []string{}
  }

  targetClass := "bg-base-300"
  var chapterList = []string{}
  var chapterLabels = []string{}
  var isInsideChapterLink = false
  var isInsideLabelSpan = false

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
                chapterLabels = append(chapterLabels, "")
                isInsideChapterLink = true
                break
              }
            }
          }
        }
      } else if token.Data == "span" && isInsideChapterLink {
        // the chapter label ("Chapter 12" / "Episode 48") sits in the only
        // span with an empty class attribute inside the chapter link
        for _, attr := range token.Attr {
          if attr.Key == "class" && attr.Val == "" {
            isInsideLabelSpan = true
          }
        }
      }
    case html.TextToken:
      if isInsideLabelSpan {
        chapterLabels[len(chapterLabels) - 1] = strings.TrimSpace(string(tokenizer.Text()))
      }
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "a" {
        isInsideChapterLink = false
      } else if token.Data == "span" {
        isInsideLabelSpan = false
      }
    }
  }

  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")

  reverseStringArray(chapterList)
  reverseStringArray(chapterLabels)

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s - %s", i + 1, chapterLabels[i], chapter))
  }

  return mangaTitle, chapterList, chapterLabels
}

// the chapter label normally comes from the chapter list; if it was missing
// there, fall back to the chapter page itself and, like the pre-rewrite logic,
// keep trying until a number turns up — never proceed without one
func getChapterNumberWeebcentral(_mangaChapter string, _chapterLabel string) string {
  regex := regexp.MustCompile(`\d+(\.\d+)?$`)
  if chapterNumber := regex.FindString(_chapterLabel); chapterNumber != "" {
    return chapterNumber
  }
  if _chapterLabel != "" {
    // e.g. a oneshot labeled just "Oneshot"
    return _chapterLabel
  }

  for {
    body, err := getPageWeebcentral(_mangaChapter)
    if err != nil {
      fmt.Println("Could not get chapter page: ", err)
      continue
    }

    var innerText string
    buttonDepth := 0
    tokenizer := html.NewTokenizer(strings.NewReader(body))
    loop: for {
      switch tokenizer.Next() {
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
      case html.ErrorToken:
        break loop
      }
    }

    if chapterNumber := regex.FindString(strings.TrimSpace(innerText)); chapterNumber != "" {
      return chapterNumber
    }
    fmt.Println("Could not find chapter number element.")
  }
}

func getChapterImagesWeebcentral(_mangaTitle string, _mangaChapter string, _chapterLabel string) {
  var urlImages string = fmt.Sprintf("%s%s", _mangaChapter, "/images?is_prev=False&current_page=1&reading_style=long_strip")

  chapterNumber := getChapterNumberWeebcentral(_mangaChapter, _chapterLabel)

  fmt.Println("Downloading chapter: ", chapterNumber)

  imgTargetClass := "max-w-full"
  var chapterImagesList []string

  // like everything else: an empty parse is retried, never skipped
  for len(chapterImagesList) == 0 {
    body, err := getPageWeebcentral(urlImages)
    if err != nil {
      fmt.Println("Could not get chapter images: ", err)
      continue
    }

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
      fmt.Println("No images found for chapter, retrying: ", chapterNumber)
    }
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

      chapterImage, err := fetchWeebcentral(chapterImageURL, "image/*")
      if err != nil {
        fmt.Println("Could not download image:", chapterImageURL, "-", err)
        return
      }

      fsCreateFile(chapterImageURL, _dir, i + 1, chapterImage, false, "")
    }(i, chapterImageURL)
  }

  wg.Wait()
}
