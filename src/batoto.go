package src

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "strconv"
  "strings"
  "golang.org/x/net/html"
  "github.com/go-rod/rod"
  "github.com/go-rod/rod/lib/launcher"
)

func batoto() {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaBatoto(mangaID)

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
      getChapterImagesBatoto(mangaTitle, chapter)
    }
  }

  fmt.Printf("\nDownload completed!\n")
}

type ChapterBatoto struct {
  ChapterLink string
  ChapterTitle string
}

func getMangaBatoto(_mangaID string) (string, []ChapterBatoto) {
  var url string = fmt.Sprintf("https://bato.to/series/%s", _mangaID)

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(url)
  page.MustWaitLoad().MustWaitIdle()

  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter list: ", err)
  }

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  targetClass := "chapt"
  var isInsideH3 = false
  var isInsideChapterLink = false

  var mangaTitle string
  var chapterList = []ChapterBatoto{}
  var currentText strings.Builder

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
                chapterList = append(chapterList, ChapterBatoto{ChapterLink: attr.Val})
                isInsideChapterLink = true
                currentText.Reset()
              }
            }
          }
        }
      } else if token.Data == "h3" {
        isInsideH3 = true
      }
    case html.TextToken:
      text := strings.TrimSpace(string(tokenizer.Text()))
      text = strings.ReplaceAll(text, "\u00A0", "")
      if isInsideH3 {
        mangaTitle = text
      } else if isInsideChapterLink && text != "" {
        if currentText.Len() > 0 {
          lastChar := currentText.String()[currentText.Len()-1]
          if lastChar != ':' && lastChar != ' ' {
            currentText.WriteString(" ")
          }
        }
        currentText.WriteString(text)
      }
    case html.EndTagToken:
      token := tokenizer.Token()
      if token.Data == "a" && isInsideChapterLink {
        formattedTitle := strings.TrimSpace(currentText.String())
        if strings.Contains(formattedTitle, ":") {
          parts := strings.Split(formattedTitle, ":")
          if len(parts) > 1 {
            formattedTitle = strings.TrimSpace(parts[0]) + ": " + strings.TrimSpace(parts[1])
          }
        }
        chapterList[len(chapterList)-1].ChapterTitle = formattedTitle
        isInsideChapterLink = false
      }
      if token.Data == "h3" {
        isInsideH3 = false
      }
    }
  }

  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")

  reverseStructStringArray(chapterList)

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter.ChapterLink))
  }

  return mangaTitle, chapterList
}

func getChapterImagesBatoto(_mangaTitle string, _mangaChapter ChapterBatoto) {
  var url string = fmt.Sprintf("https://bato.to%s", _mangaChapter.ChapterLink)

  browser := rod.New().ControlURL(launcher.New().Headless(true).MustLaunch()).MustConnect()
  defer browser.MustClose()
  page := browser.MustPage(url)
  page.MustWaitLoad().MustWaitIdle()
  imgs, err := page.Elements("img.page-img")
  if err != nil {
    fmt.Println("Error fetching image elements:", err)
    return
  }
  if len(imgs) == 0 {
    fmt.Println("No images found with class 'page-img'.")
    return
  }
  body, err := page.HTML()
  if err != nil {
    fmt.Println("Could not get chapter images: ", err)
    return
  }

  fmt.Println("Downloading chapter: ", _mangaChapter.ChapterTitle)

  var chapterImagesList = []string{}

  reader := strings.NewReader(string(body))
  tokenizer := html.NewTokenizer(reader)
  targetClass := "page-img"

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

  dir := fmt.Sprintf("downloads/%s/%s", _mangaTitle, _mangaChapter.ChapterTitle)
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
