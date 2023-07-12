package src

import (
  "fmt"
  "errors"
  "io/ioutil"
  "log"
  "net/http"
  "encoding/json"
  "sort"
  "strconv"
  "strings"
  "bufio"
  "os"
)

func mangadex() {
  fmt.Printf("Enter the Manga ID: ")
  var userInput string
  fmt.Scanf("%s", &userInput)

  // var userInput = "76ee7069-23b4-493c-bc44-34ccbf3051a8" // Tomo-chan

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
  var mangaTitle = mangaInfo.Data.Attributes.Title.EN;
  fmt.Println(mangaTitle)
  fmt.Println("Number of chapters: ", len(mangaChapters.Data))
  fmt.Println("available chapters:")

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
  fmt.Println("Options: 'all', 'asf (all chapters in a single folder)', 'chapter numbers separated by spaces', 'oneshot', 'quit'\n")
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
      case "quit":
        break loop
      default:
        getMangaChapterImages(mangaTitle, mangaChapters, userInput, false)
    }
    break loop
  }
}

type MangaData struct {
  Result string `json:"result"`
  Data struct {
    ID string `json:"id"`
    Attributes struct {
      Title struct {
        EN string `json:"en"`
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
    log.Fatalln(err)
    return mangaData, errors.New("Could not get manga")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
    return mangaData, errors.New("Could not parse body")
  }
  if err := json.Unmarshal(body, &mangaData); err != nil {
    log.Fatalln(err)
    fmt.Println("Could not unmarshal JSON")
    return mangaData, errors.New("Could not unmarshal JSON")
  }
  return mangaData, nil
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

  var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s/feed?includeFuturePublishAt=0&limit=%d&offset=%d&translatedLanguage[]=%s", _mangaInfo.Data.ID, queryLimit, offset, selectedLanguage)
  var mangaChapters MangaChapters

  resp, err := http.Get(url)
  if err != nil {
    log.Fatalln(err)
    return mangaChapters, errors.New("Could get manga chapter")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
    return mangaChapters, errors.New("Could not parse body")
  }
  if err := json.Unmarshal(body, &mangaChapters); err != nil {
    log.Fatalln(err)
    fmt.Println("Could not unmarshal JSON")
    return mangaChapters, errors.New("Could not unmarshal JSON")
  }

  if mangaChapters.Total > queryLimit {
    for offset < queryLimit {
      offset += 500
      var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s/feed?includeFuturePublishAt=0&limit=%d&offset=%d&translatedLanguage[]=%s", _mangaInfo.Data.ID, queryLimit, offset, selectedLanguage)
      var mangaChapters2 MangaChapters

      resp, err := http.Get(url)
      if err != nil {
        log.Fatalln(err)
        return mangaChapters, errors.New("Could get manga chapter")
      }
      defer resp.Body.Close()
      body, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        log.Fatalln(err)
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

      resp, err := http.Get(url)
      if err != nil {
        log.Fatalln(err)
        // return mangaChapterImages, errors.New("Could not get manga chapter")
        break
      }
      defer resp.Body.Close()
      body, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        log.Fatalln(err)
        // return mangaChapterImages, errors.New("Could not parse body")
        break
      }
      if err := json.Unmarshal(body, &mangaChapterImages); err != nil {
        log.Fatalln(err)
        fmt.Println("Could not unmarshal JSON")
        // return mangaChapterImages, errors.New("Could not unmarshal JSON")
        break
      }

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
      fsCreateDir(dir, _singleFolder)
      var j int = 0
      j: for {
        var url string = fmt.Sprintf("https://uploads.mangadex.org/data/%s/%s", mangaChapterImages.Chapter.Hash, mangaChapterImages.Chapter.Data[j])
        resp, err := http.Get(url)
        if err != nil {
          log.Fatalln(err)
          // return mangaChapterImages, errors.New("Could not get chapter image")
          break j
        }
        defer resp.Body.Close()
        chapterImage, err := ioutil.ReadAll(resp.Body)
        if err != nil {
          log.Fatalln(err)
          fmt.Println("Could not parse chapterImage")
          // return mangaChapterImages, errors.New("Could not parse chapterImage")
          break j
        }

        // was trying to make 'asf' work. will I ever need to use this option again?
        // var fileNameNumber string
        // if _userInput == "" && _singleFolder {
        //   fileNameNumber = strconv.Atoi(*_mangaChapters.Data[i].Attributes.Chapter)
        // } else {
        //   fileNameNumber = j + 1
        // }
        // fsCreateFile(mangaChapterImages.Chapter.Data[j], dir, fileNameNumber, _singleFolder, _userInput == "oneshot", chapterImage)
        // fsCreateFile(mangaChapterImages.Chapter.Data[j], dir, j + 1, _singleFolder, _userInput == "oneshot", chapterImage)
        fsCreateFile(mangaChapterImages.Chapter.Data[j], dir, j + 1, chapterImage)

        if j < len(mangaChapterImages.Chapter.Data) - 1 {
          j++
        } else {
          break j
        }
      }
    }
    if i < len(_mangaChapters.Data) - 1 {
      i++
    } else {
      break i
    }
  }
  fmt.Printf("\nDownload completed!\n")
}
