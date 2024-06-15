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

func mangasee() {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaMangasee(mangaID)

  fmt.Println("\nEnter the range of chapters you want to download.")

  fmt.Printf("\nInitial chapter: ")
  var userInputFirstChapter string
  fmt.Scanf("%s", &userInputFirstChapter)
  firstChapter, _ := strconv.Atoi(userInputFirstChapter)

  fmt.Printf("\nLast chapter: ")
  var userInputLastChapter string
  fmt.Scanf("%s", &userInputLastChapter)
  lastChapter, _ := strconv.Atoi(userInputLastChapter)
  fmt.Printf("")

  for i, chapter := range chapterList {
    if i >= firstChapter - 1 && i <= lastChapter - 1 {
      getChapterImagesMangasee(mangaTitle, chapter)
    }
  }

  fmt.Printf("\nDownload completed!\n")
}

func getMangaMangasee(_mangaID string) (string, []string) {
  var url string = fmt.Sprintf("https://mangasee123.com/manga/%s", _mangaID)

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(url)
  page.MustWaitLoad().MustWaitIdle()

  if page.MustHas("div.ShowAllChapters") {
    button := page.MustElement("div.ShowAllChapters")
    button.MustClick()
    page.MustWaitLoad().MustWaitIdle()
    time.Sleep(1 * time.Second)
  }

  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter list: ", err)
  }

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  targetClass := "ChapterLink"
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

func getChapterImagesMangasee(_mangaTitle string, _mangaChapter string) {
  var url string = fmt.Sprintf("https://mangasee123.com%s", _mangaChapter)

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(url)
  page.MustWaitLoad().MustWaitIdle()
  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter images: ", err)
  }

  regex, _ := regexp.Compile(`-chapter-([0-9]+(\.[0-9]+)?)`)

  mangaChapterNumber := regex.FindStringSubmatch(_mangaChapter)

  fmt.Println("Downloading chapter: ", mangaChapterNumber[1])

  var chapterImagesList = []string{}

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  targetClass := "HasGap"

  loop: for {
    tokenType := tokenizer.Next()
    switch tokenType {
    case html.ErrorToken:
      break loop
    case html.StartTagToken, html.SelfClosingTagToken:
      token := tokenizer.Token()
      if token.Data == "img" {
        for _, attr := range token.Attr {
          if attr.Key == "class" && strings.Contains(attr.Val, targetClass) {
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

  dir := fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, mangaChapterNumber[1])
  _dir := fsCreateDir(dir, false)
  for i, chapterImageURL := range chapterImagesList {
    var chapterImage []byte
    for {
      resp, err := http.Get(chapterImageURL)
      if err != nil {
        fmt.Println("Request error. Retrying.")
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
  }
}
