package src

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strconv"
  "strings"
  "time"
  "golang.org/x/net/html"
  "github.com/go-rod/rod"
  "github.com/go-rod/rod/lib/launcher"
)

func weebcentral() {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaWeebcentral(mangaID)

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

  for i, chapter := range chapterList {
    if i >= firstChapter - 1 && i <= lastChapter - 1 {
      getChapterImagesWeebcentral(mangaTitle, chapter)
    }
  }

  fmt.Printf("\nDownload completed!\n")
}

func getMangaWeebcentral(_mangaID string) (string, []string) {
  var url string = fmt.Sprintf("https://weebcentral.com/series/%s", _mangaID)

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(url)
  page.MustWaitLoad().MustWaitIdle()

  if page.MustHas(`button.p-2`) {
    button := page.MustElement(`button.p-2`)
    button.MustClick()
    page.MustWaitLoad().MustWaitIdle()
    time.Sleep(2 * time.Second)
  }

  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter list: ", err)
  }

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  targetClass := "bg-base-300"
  var isInsideH1 = false

  var mangaTitle string
  var chapterList = []string{}

  loop: for {
    tokenType := tokenizer.Next()
    switch tokenType {
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
                chapterList = append(chapterList, result)
                break
              }
            }
          }
        }
      } else if token.Data == "h1" {
        isInsideH1 = true
      }
    case html.TextToken:
      if isInsideH1 {
        mangaTitle = string(tokenizer.Text())
      }
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "a" {
      } else if token.Data == "h1" {
        isInsideH1 = false
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

func getChapterImagesWeebcentral(_mangaTitle string, _mangaChapter string) {
  var urlChapterNumber string = fmt.Sprintf("%s", _mangaChapter)
  var urlImages string = fmt.Sprintf("%s%s", _mangaChapter, "/images?is_prev=False&current_page=1&reading_style=long_strip")

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(urlImages)
  page.MustWaitLoad().MustWaitIdle()
  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter images: ", err)
  }

  chapterTargetClass := "button.col-span-4"
  var mangaChapterNumber = []string{}

  {
    for {
      browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
      defer browser.MustClose()
      page := browser.MustPage(urlChapterNumber)
      page.MustWaitLoad().MustWaitIdle()

      var innerText string = ""

      _, err := page.HTML()
      if err != nil {
        fmt.Println("Could not get chapter images: ", err)
      }

      if page.MustHas(chapterTargetClass) {
        elements := page.MustElements(chapterTargetClass)

        if len(elements) > 0 {
          innerText = elements[0].MustText()
        }
      } else {
        fmt.Println("Could not find chapter number element.")
      }

      regex := regexp.MustCompile(`\d+(\.\d+)?$`)
      mangaChapterNumber = regex.FindStringSubmatch(innerText)

      if len(mangaChapterNumber) > 0 {
        break
      }
    }
  }

  fmt.Println("Downloading chapter: ", mangaChapterNumber[0])

  var chapterImagesList = []string{}

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  imgTargetClass := "maw-w-full"

  loop: for {
    tokenType := tokenizer.Next()
    switch tokenType {
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
    case html.TextToken:
      if true {}
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "img" {}
    }
  }

  dir := fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, mangaChapterNumber[0])
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
      req.Header.Set("Referer", "https://weebcentral.com/")
      req.Header.Set("x-referer", "https://weebcentral.com/")
      req.Header.Set("Accept", "image/*")
      req.Header.Set("Accept-Encoding", "gzip, deflate, br")

      resp, err := client.Do(req)
      if err != nil {
        fmt.Println("Request error:", err)
        continue
      }
      defer resp.Body.Close()

      res, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        fmt.Println("Request error. Retrying.")
      } else {
        chapterImage = res
        break
      }
    }

    fsCreateFile(chapterImageURL, _dir, i + 1, chapterImage, false, "")

    // time.Sleep(1 * time.Second) // in case I need to debugg this goddamn website blocking requests again
  }
}
