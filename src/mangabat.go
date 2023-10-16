package src

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strconv"
  "strings"
  "golang.org/x/net/html"
)

func mangabat() {
  fmt.Printf("\nEnter the Manga ID: ")
  var mangaID string
  fmt.Scanf("%s", &mangaID)

  mangaTitle, chapterList := getMangaMangabat(mangaID)

  fmt.Println("\nEnter the range of chapters you want to download\n")

  fmt.Printf("\nInitial chapter: ")
  var userInputFirstChapter string
  fmt.Scanf("%s", &userInputFirstChapter)
  firstChapter, _ := strconv.Atoi(userInputFirstChapter)

  fmt.Printf("\nLast chapter: ")
  var userInputLastChapter string
  fmt.Scanf("%s", &userInputLastChapter)
  lastChapter, _ := strconv.Atoi(userInputLastChapter)

  for i, chapter := range chapterList {
    if i >= firstChapter - 1 && i <= lastChapter - 1 {
      getChapterImagesMangabat(mangaTitle, chapter)
    }
  }

  fmt.Printf("\nDownload completed!\n")
}

func getMangaMangabat(_mangaID string)  (string, []string) {
  var url string = fmt.Sprintf("https://readmangabat.com/read-%s", _mangaID)

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get chapter list")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body chapter list")
  }

  reader := strings.NewReader(string(body))
	tokenizer := html.NewTokenizer(reader)
  targetClass := "chapter-name text-nowrap"
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
                chapterList = append(chapterList, attr.Val)
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

  fmt.Println(mangaTitle)

  reverseStringArray(chapterList)

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter))
  }

  return mangaTitle, chapterList
}

func getChapterImagesMangabat(_mangaTitle string, _mangaChapter string)  {
  var url string = _mangaChapter

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get chapter images")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body chapter images")
  }

  regex,_ := regexp.Compile(`https://[a-zA-Z0-9.-]*mkklcdnv[a-zA-Z0-9.-]*/[^ "\n]+`)
  regex2, _ := regexp.Compile(`-chap-(\d+(?:\.\d+)?)`)

  var chapterImagesListRAW []string = regex.FindAllString(string(body), -1)

  mangaChapterNumber := regex2.FindStringSubmatch(_mangaChapter)

  fmt.Println("Downloading chapter: ", mangaChapterNumber[1])

  var chapterImagesList = []string{}
  for _, chapterImageURL := range chapterImagesListRAW {
    if strings.Contains(chapterImageURL, "chapter") {
      chapterImagesList = append(chapterImagesList, chapterImageURL)
    }
  }

  dir := fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, mangaChapterNumber[1])
  _dir := fsCreateDir(dir, false)
  for i, chapterImageURL := range chapterImagesList {
    var chapterImage []byte
    for {
      client := &http.Client{}
	    req, err := http.NewRequest("GET", chapterImageURL, nil)
      if err != nil {
        fmt.Println("Request error. Retrying.")
      }
      req.Header.Set("Referer", "https://readmangabat.com/")
      // req.Header.Set("referer", "https://h.mangabat.com/")
      // req.Header.Set("Referer", "https://chapmanganato.com/")
      // req.Header.Set("Referer", "https://mangakakalot.com/")
      // req.Header.Set("Referer", "https://manganelo.com/")
      // req.Header.Set("Referer", "https://readmanganato.com/")

      resp, err := client.Do(req)
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
