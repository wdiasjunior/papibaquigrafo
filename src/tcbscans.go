package src

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strconv"
  "strings"
)

func tcbscans() {
  mangaList := getMangaList()

  fmt.Printf("\nEnter the Manga ID: ")
  var userInput string
  fmt.Scanf("%s", &userInput)
  mangaID, _ := strconv.Atoi(userInput)

  chapterList := getChapterList(mangaList[mangaID - 1])

  fmt.Println("\nEnter the range of chapters you want to download\n")

  fmt.Printf("\nInitial chapter: ")
  var userInputFirstChapter string
  fmt.Scanf("%s", &userInputFirstChapter)
  firstChapter, _ := strconv.Atoi(userInputFirstChapter)

  fmt.Printf("\nLast chapter: ")
  var userInputLastChapter string
  fmt.Scanf("%s", &userInputLastChapter)
  lastChapter, _ := strconv.Atoi(userInputLastChapter)

  regex, _ := regexp.Compile(`[^/]*$`)
  mangaTitle := regex.FindAllString(mangaList[mangaID - 1], -1)[0]
  mangaTitleCapitalized := strings.Title(strings.Replace(mangaTitle, "-", " ", -1))
  fmt.Println("\n")
  fmt.Println(mangaTitleCapitalized)

  for i, chapter := range chapterList {
    if i >= firstChapter && i <= lastChapter {
      getChapterImages(mangaTitleCapitalized, chapter)
    }
  }

  fmt.Printf("\nDownload completed!\n")
}

func getMangaList() []string {
  var url string = "https://tcbscans.com/projects"

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get manga list")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body manga list")
  }

  regex, _ := regexp.Compile(`/mangas/(.*?)[^"]+`)

  var mangaListRaw []string = regex.FindAllString(string(body), -1)
  var mangaList = []string{}

  regex2, _ := regexp.Compile(`/mangas/[^/]+/([^/]+)`)
  for i, manga := range mangaListRaw {
    if i % 2 == 0 {
      mangaList = append(mangaList, manga)
      mangaTitle := regex2.FindStringSubmatch(manga)
      fmt.Println(fmt.Sprintf("%d - %s", (i / 2 + 1), mangaTitle[1]))
    }
  }

  return mangaList
}

func getChapterList(_mangaURL string) []string {
  var url string = fmt.Sprintf("https://tcbscans.com%s", _mangaURL)

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get chapter list")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body chapter list")
  }

  regex, _ := regexp.Compile(`/chapters/(.*?)[^"]+`)
  var chapterList []string = regex.FindAllString(string(body), -1)
  // sort.Slice(chapterList, func(i, j int) bool { return chapterList[i] < chapterList[j] })
  reverseStringArray(chapterList)

  // fmt.Println("Could not parse body manga list")

  for i, chapter := range chapterList {
    fmt.Println(fmt.Sprintf("%d - %s", i + 1, chapter))
  }

  return chapterList
}

func getChapterImages(_mangaTitle string, _mangaChapter string)  {
  var url string = fmt.Sprintf("https://tcbscans.com%s", _mangaChapter)

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get chapter images")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body chapter images")
  }

  regex, _ := regexp.Compile(`https://cdn.onepiecechapters.com/file/(.*?)[^"]+`)
  regex2, _ := regexp.Compile(`/chapters/\d+/[A-Za-z0-9-]+-chapter-(\d+)(?:-review-\d+)?`)

  var chapterImagesList []string = regex.FindAllString(string(body), -1)

  fmt.Println(chapterImagesList)

  mangaChapterNumber := regex2.FindStringSubmatch(_mangaChapter)

  fmt.Println("Downloading chapter: ", mangaChapterNumber[1])

  dir := fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, mangaChapterNumber[1])
  _dir := fsCreateDir(dir, false)

  for i, chapterImageURL := range chapterImagesList {
    if !strings.Contains(chapterImageURL, "dragonball.png") &&
       !strings.Contains(chapterImageURL, "siberianhusky.jpg") &&
       !strings.Contains(chapterImageURL, "ign.png") {
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
}
