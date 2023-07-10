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
  // fmt.Printf("Enter the Manga ID: ")
  // var userInput string
  // fmt.Scanf("%s", &userInput)

  // var userInput = "ead4b388-cf7f-448c-aec6-bf733968162c" // Hanabi - oneshot
  // var userInput = "76ee7069-23b4-493c-bc44-34ccbf3051a8" // Tomo-chan
  var userInput = "6fef1f74-a0ad-4f0d-99db-d32a7cd24098" // fire punch
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
  // fmt.Println("mangaChapters.Total", mangaChapters.Total)
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
    // var userInput string
    // fmt.Scanf("%s", &userInput)
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
    fmt.Printf("\nDownload completed!\n")
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
  // fmt.Println(_mangaTitle, _mangaChapters, _userInput, _singleFolder)

  fmt.Println("\nDownloading")

  // var mangaTitleDirectory string = fmt.Sprintf("downloads/%s", _mangaTitle)
  // fsCreateDir(mangaTitleDirectory)



  userInput := strings.Split(_userInput, " ")
  // if !true {
  //   fmt.Println(userInput)
  // }
  var i int = 0
  i: for {
    // fmt.Println(len(_userInput), *_mangaChapters.Data[i].Attributes.Chapter)

    if (contains(userInput, *_mangaChapters.Data[i].Attributes.Chapter)) || (_userInput == "oneshot") || (_userInput == "") {

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
        if len(*_mangaChapters.Data[i].Attributes.Title) > 0 {
          dir = fmt.Sprintf("downloads/%s/Ch.%s - %s", _mangaTitle, *_mangaChapters.Data[i].Attributes.Chapter, &_mangaChapters.Data[i].Attributes.Title)
        } else {
          dir = fmt.Sprintf("downloads/%s/Ch.%s", _mangaTitle, *_mangaChapters.Data[i].Attributes.Chapter)
        }
      }
      // fmt.Println("aqui", _mangaTitle, *_mangaChapters.Data[i].Attributes.Chapter, *_mangaChapters.Data[i].Attributes.Title, _mangaChapters.Data[i].ID)
      fsCreateDir(dir)

      // if condition {
        // var j int = 0
        // j: for {
        //   var url string = fmt.Sprintf("https://uploads.mangadex.org/data/%s/%s", mangaChapterImages.Chapter.Hash, mangaChapterImages.Chapter.Data[j])
        //   resp, err := http.Get(url)
        //   if err != nil {
        //     log.Fatalln(err)
        //     // return mangaChapterImages, errors.New("Could not get chapter image")
        //     break j
        //   }
        //   defer resp.Body.Close()
        //   chapterImages, err := ioutil.ReadAll(resp.Body)
        //   if err != nil {
        //     log.Fatalln(err)
        //     // return mangaChapterImages, errors.New("Could not parse chapterImage")
        //     break j
        //   }
        //   // mangaChapterImages.Chapter.Data[j]
        //   // fsCreateDir("downloads/Fire Punch/Ch.1")
        //   // fsCreateFile("teste.png", "downloads/Fire Punch/Ch.1", j + 1, chapterImage)
        //   // fmt.Println(chapterImages)
        //
        //   if j < len(chapterImages) - 1 {
        //     j++
        //   } else {
        //     break j
        //   }
        // }
      // }
      }
    if i < len(_mangaChapters.Data) - 1 {
      i++
    } else {
      break i
    }
  }

  // fmt.Println(userInput)
  // if _userInput == ""  {
  //   // for _, chapter := range _userInput {
  //     fmt.Println(userInput)
  //   // }
  // }
  // if _userInput == "oneshot"  {
  //   fmt.Println("oneshot")
  // }
}
