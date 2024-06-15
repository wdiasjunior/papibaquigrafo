package src

import (
  "fmt"
  "errors"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "sort"
  "strconv"
  "strings"
  "bufio"
  "os"
)

// TODO - if "all" and name null == oneshot (probably) fix empty chapter number and put 0

func mangadex() {
  fmt.Printf("\nEnter the Manga ID: ")
  var userInput string
  fmt.Scanf("%s", &userInput)

  mangaInfo, err := getManga(userInput)
  if err != nil {
    fmt.Println(err)
    return
  }
  if mangaInfo.Result != "ok" {
    fmt.Println(mangaInfo.Result)
    return
  }

  mangaChapters, err := getMangaChapters(mangaInfo)
  if err != nil {
    fmt.Println(err)
    return
  }
  if len(mangaChapters.Data) == 0 {
    fmt.Println("No chapters available\n")
    return
  }
  var mangaTitle string
  if mangaInfo.Data.Attributes.Title.EN != "" {
    mangaTitle = mangaInfo.Data.Attributes.Title.EN
  } else if mangaInfo.Data.Attributes.Title.JARomaji != "" {
    mangaTitle = mangaInfo.Data.Attributes.Title.JARomaji
  } else if mangaInfo.Data.Attributes.Title.JA != "" {
    mangaTitle = mangaInfo.Data.Attributes.Title.JA
  } else {
    mangaTitle = "manga - unknown title"
  }
  fmt.Println("")
  fmt.Println(mangaTitle)
  fmt.Println("")
  fmt.Println("Number of chapters: ", len(mangaChapters.Data))
  fmt.Println("Available chapters:")

  var arr []string
  if len(mangaChapters.Data) == 1 && mangaChapters.Data[0].Attributes.Chapter == nil {
    fmt.Println("Oneshot")
  } else {
    for _, chapter := range mangaChapters.Data {
      if chapter.Attributes.Chapter != nil {
        arr = append(arr, *chapter.Attributes.Chapter)
      }
    }

    sort.Slice(arr, func(i, j int) bool {
      numA, _ := strconv.ParseFloat(arr[i], 64)
      numB, _ := strconv.ParseFloat(arr[j], 64)
      return numA < numB
    })

    for _, i := range arr {
      fmt.Println(i)
    }
  }

  fmt.Println("\nEnter the chapters you want to download\n")
  fmt.Println("Options: 'all', 'asf (all chapters in a single folder)', 'chapter numbers separated by spaces', 'oneshot', 'covers', 'quit'\n")
  loop: for {
    fmt.Printf("-> ")
    _input := bufio.NewReader(os.Stdin)
    userInput, _ := _input.ReadString('\n')
    userInput = strings.TrimSuffix(userInput, "\n")

    switch userInput {
      case "all":
        getMangaChapterImages(mangaTitle, mangaChapters, "", false)
        break loop
      case "asf":
        getMangaChapterImages(mangaTitle, mangaChapters, "", true)
        break loop
      case "oneshot":
        getMangaChapterImages(mangaTitle, mangaChapters, "oneshot", true)
        break loop
      case "covers":
        getMangaCovers(mangaTitle, mangaInfo.Data.ID)
        break loop
      case "quit":
        break loop
      default:
        getMangaChapterImages(mangaTitle, mangaChapters, userInput, false)
    }
    break loop
  }
  fmt.Printf("\nDownload completed!\n")
}

type MangaData struct {
  Result string `json:"result"`
  Data struct {
    ID string `json:"id"`
    Attributes struct {
      Title struct {
        EN string `json:"en"`
        JARomaji string `json:"ja-ro"`
        JA string `json:"ja"`
      } `json:"title"`
      AvailableTranslatedLanguages []string `json:"availableTranslatedLanguages"`
    } `json:"attributes"`
  } `json:"data"`
}

func getManga(_mangaId string) (MangaData, error) {
  var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s", _mangaId)
  var mangaData MangaData

  resp, err := http.Get(url)
  if err != nil {
    return mangaData, errors.New("Could not get manga info")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return mangaData, errors.New("Could not parse body")
  }
  if err := json.Unmarshal(body, &mangaData); err != nil {
    fmt.Println("Could not unmarshal JSON")
    return mangaData, errors.New("Could not unmarshal JSON")
  }
  return mangaData, nil
}

type MangaCoversData struct {
  Result string `json:"result"`
  Data []struct {
    Attributes struct {
      Volume string `json:"volume"`
      FileName string `json:"fileName"`
      Locale string `json:"locale"`
    } `json:"attributes"`
  } `json:"data"`
  Total int `json:"total"`
}

func getMangaCovers(_mangaTitle string , _mangaId string) {
  // TODO - this breaks on one piece - 108 covers
  var _limit = 100
  var url string = fmt.Sprintf("https://api.mangadex.org/cover?limit=%d&manga[]=%s&order[createdAt]=asc&order[updatedAt]=asc&order[volume]=asc", _limit, _mangaId)
  var mangaCoversData MangaCoversData

  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Could not get manga covers info")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Could not parse body")
  }
  if err := json.Unmarshal(body, &mangaCoversData); err != nil {
    fmt.Println("Could not unmarshal JSON")
  }
  var dir string = fmt.Sprintf("downloads/%s", _mangaTitle)
  _dir := fsCreateDir(dir, true)
  for _, cover := range mangaCoversData.Data {
    // skip covers that are not en or jp
    if cover.Attributes.Locale == "ja" || cover.Attributes.Locale == "en" {
      var url string = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", _mangaId, cover.Attributes.FileName)
      var coverImage []byte
      for {
        resp, err := http.Get(url)
        if err != nil {
          fmt.Println("Request error. Retrying.")
        }
        defer resp.Body.Close()
        res, err := ioutil.ReadAll(resp.Body)
        if err != nil {
          fmt.Println("Request error. Retrying.")
        } else {
          coverImage = res
          break
        }
      }

      var coverFileName = fmt.Sprintf("Cover %s - %s", cover.Attributes.Volume, cover.Attributes.Locale)
      fsCreateFile(cover.Attributes.FileName, _dir, 0, coverImage, true, coverFileName)
    }
  }
}

type MangaChapters struct {
  Result string `json:"result"`
  Data []struct {
    ID string `json:"id"`
    Attributes struct {
      Chapter *string `json:"chapter"`
      Title *string `json:"title"`
    } `json:"attributes"`
  } `json:"data"`
  Total int `json:"total"`
}

func getMangaChapters(_mangaInfo MangaData) (MangaChapters, error) {
  var queryLimit int = 500
  var offset int = 0
  var selectedLanguage string = "en"

  var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s/feed?includeFuturePublishAt=0&limit=%d&offset=%d&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic&translatedLanguage[]=%s", _mangaInfo.Data.ID, queryLimit, offset, selectedLanguage)
  var mangaChapters MangaChapters

  resp, err := http.Get(url)
  if err != nil {
    return mangaChapters, errors.New("Could get manga chapter")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return mangaChapters, errors.New("Could not parse body")
  }
  if err := json.Unmarshal(body, &mangaChapters); err != nil {
    fmt.Println("Could not unmarshal JSON")
    return mangaChapters, errors.New("Could not unmarshal JSON")
  }

  if mangaChapters.Total > queryLimit {
    for offset < queryLimit {
      offset += 500
      var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s/feed?includeFuturePublishAt=0&limit=%d&offset=%d&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic&translatedLanguage[]=%s", _mangaInfo.Data.ID, queryLimit, offset, selectedLanguage)
      var mangaChapters2 MangaChapters

      resp, err := http.Get(url)
      if err != nil {
        return mangaChapters, errors.New("Could get manga chapters")
      }
      defer resp.Body.Close()
      body, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        return mangaChapters, errors.New("Could not parse body")
      }
      if err := json.Unmarshal(body, &mangaChapters2); err != nil {
        fmt.Println("Could not unmarshal JSON")
        return mangaChapters, errors.New("Could not unmarshal JSON")
      }
      mangaChapters.Data = append(mangaChapters.Data, mangaChapters2.Data...)
    }
  }

  return mangaChapters, nil
}

type MangaImages struct {
  Result string `json:"result"`
  Chapter struct {
    Hash string `json:"hash"`
    Data []string `json:"data"`
  } `json:"chapter"`
}

func getMangaChapterImages(_mangaTitle string, _mangaChapters MangaChapters, _userInput string, _singleFolder bool) {
  fmt.Println("\nStarting Download")

  userInput := strings.Split(_userInput, " ")

  var i int = 0
  i: for {
    var chapterNameNoNIL string
    if _userInput == "oneshot" {
      chapterNameNoNIL = "Oneshot"
    } else if _mangaChapters.Data[i].Attributes.Chapter == nil {
      chapterNameNoNIL = ""
    } else {
      chapterNameNoNIL = *_mangaChapters.Data[i].Attributes.Chapter
    }
    if (contains(userInput, chapterNameNoNIL, _userInput == "oneshot")) || (_userInput == "oneshot") || (_userInput == "") {
      var url string = fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", _mangaChapters.Data[i].ID)
      var mangaChapterImages MangaImages

      for {
        resp, err := http.Get(url)
        if err != nil {
          fmt.Println("Request error. Retrying.")
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
          fmt.Println("Request error. Retrying.")
        }
        if err := json.Unmarshal(body, &mangaChapterImages); err != nil {
          fmt.Println("Could not unmarshal JSON - manga chapter images")
        } else {
          break
        }
      }

      if len(mangaChapterImages.Chapter.Data) > 0 {
        var dir string
        if _singleFolder {
          dir = fmt.Sprintf("downloads/%s", _mangaTitle)
        } else if _userInput == "oneshot" {
          dir = fmt.Sprintf("downloads/%s/Oneshot", _mangaTitle)
        } else {
          if _mangaChapters.Data[i].Attributes.Title != nil && len(*_mangaChapters.Data[i].Attributes.Title) > 0 {
            dir = fmt.Sprintf("downloads/%s/Ch.%s - %s", _mangaTitle, chapterNameNoNIL, *_mangaChapters.Data[i].Attributes.Title)
          } else {
            dir = fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, chapterNameNoNIL)
          }
        }
        fmt.Println("Downloading chapter: ", chapterNameNoNIL)
        _dir := fsCreateDir(dir, _singleFolder)
        var j int = 0
        j: for {
          var url string = fmt.Sprintf("https://uploads.mangadex.org/data/%s/%s", mangaChapterImages.Chapter.Hash, mangaChapterImages.Chapter.Data[j])
          var chapterImage []byte
          for {
            resp, err := http.Get(url)
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

          fsCreateFile(mangaChapterImages.Chapter.Data[j], _dir, j + 1, chapterImage, false, "")
          if j < len(mangaChapterImages.Chapter.Data) - 1 {
            j++
          } else {
            break j
          }
        }
      }
    }
    if i < len(_mangaChapters.Data) - 1 {
      i++
    } else {
      break i
    }
  }
}
