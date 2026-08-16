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

func mangadex() DownloadResult {
  fmt.Printf("\nEnter the Manga ID: ")
  var userInput string
  fmt.Scanf("%s", &userInput)

  mangaInfo, err := getMangaMangadex(userInput)
  if err != nil {
    fmt.Println(err)
    return downloadFailed("", err.Error())
  }
  if mangaInfo.Result != "ok" {
    fmt.Println(mangaInfo.Result)
    return downloadFailed("", mangaInfo.Result)
  }

  mangaChapters, err := getMangaChaptersMangadex(mangaInfo)
  if err != nil {
    fmt.Println(err)
    return downloadFailed("", err.Error())
  }
  if len(mangaChapters.Data) == 0 {
    fmt.Println("No chapters available\n")
    // return
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

    for _, i := range arr {
      fmt.Println(i)
    }
  }

  fmt.Println("\nEnter the chapters you want to download\n")
  fmt.Println("Options: 'all', 'asf (all chapters in a single folder)', 'range', 'chapter numbers separated by spaces', 'oneshot', 'covers', 'quit'\n")
  // One reader for the whole prompt - a fresh one per read would buffer ahead
  // and swallow whatever the range prompts are waiting on
  _input := bufio.NewReader(os.Stdin)
  cancelled := false
  // Only the range option knows how many chapters it picked
  selectedChapters := 0
  loop: for {
    fmt.Printf("-> ")
    userInput, _ := _input.ReadString('\n')
    userInput = strings.TrimSuffix(userInput, "\n")

    switch userInput {
      case "all":
        getMangaChapterImagesMangadex(mangaTitle, mangaChapters, "", false)
        break loop
      case "asf":
        getMangaChapterImagesMangadex(mangaTitle, mangaChapters, "", true)
        break loop
      case "range":
        rangeChapters, err := selectChapterRangeMangadex(_input, mangaChapters)
        if err != nil {
          fmt.Println("\n" + err.Error())
          return downloadFailed(mangaTitle, err.Error())
        }
        selectedChapters = len(rangeChapters.Data)
        getMangaChapterImagesMangadex(mangaTitle, rangeChapters, "", false)
        break loop
      case "oneshot":
        getMangaChapterImagesMangadex(mangaTitle, mangaChapters, "oneshot", true)
        break loop
      case "covers":
        getMangaCoversMangadex(mangaTitle, mangaInfo.Data.ID)
        break loop
      case "quit":
        cancelled = true
        break loop
      default:
        getMangaChapterImagesMangadex(mangaTitle, mangaChapters, userInput, false)
    }
    break loop
  }

  if cancelled {
    return downloadCancelled()
  }

  fmt.Printf("\nDownload completed!\n")

  // Mangadex downloads every selected chapter inside one call, so outside of a
  // range there is no per chapter count to report here
  return downloadSuccess(mangaTitle, selectedChapters)
}

// The listing prints chapter numbers instead of positions, so the range is read
// as chapter numbers too - the same values the "chapter numbers" option takes.
// Decimals such as 10.5 are part of the numbering here, hence the float parse.
func selectChapterRangeMangadex(_input *bufio.Reader, _mangaChapters MangaChaptersMangadex) (MangaChaptersMangadex, error) {
  fmt.Println("\nEnter the range of chapters you want to download.")

  fmt.Printf("\nInitial chapter: ")
  userInputFirstChapter, _ := _input.ReadString('\n')
  firstChapter, err := strconv.ParseFloat(strings.TrimSpace(userInputFirstChapter), 64)
  if err != nil {
    return MangaChaptersMangadex{}, errors.New("Invalid initial chapter.")
  }

  fmt.Printf("\nLast chapter: ")
  userInputLastChapter, _ := _input.ReadString('\n')
  lastChapter, err := strconv.ParseFloat(strings.TrimSpace(userInputLastChapter), 64)
  if err != nil {
    return MangaChaptersMangadex{}, errors.New("Invalid last chapter.")
  }

  if lastChapter < firstChapter {
    firstChapter, lastChapter = lastChapter, firstChapter
  }

  // Same struct minus the chapters outside the range, so the download path can
  // treat it as if it were the whole feed
  rangeChapters := _mangaChapters
  rangeChapters.Data = nil
  for _, chapter := range _mangaChapters.Data {
    // Chapters with no number - oneshots and the like - cannot be placed in a
    // range, so they are left to the 'oneshot' option
    if chapter.Attributes.Chapter == nil {
      continue
    }
    chapterNumber, err := strconv.ParseFloat(*chapter.Attributes.Chapter, 64)
    if err != nil {
      continue
    }
    if chapterNumber >= firstChapter && chapterNumber <= lastChapter {
      rangeChapters.Data = append(rangeChapters.Data, chapter)
    }
  }
  rangeChapters.Total = len(rangeChapters.Data)

  if len(rangeChapters.Data) == 0 {
    return MangaChaptersMangadex{}, errors.New("No chapters in the selected range.")
  }

  fmt.Printf("\n")

  return rangeChapters, nil
}

type MangaDataMangadex struct {
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

func getMangaMangadex(_mangaId string) (MangaDataMangadex, error) {
  var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s", _mangaId)
  var mangaData MangaDataMangadex

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

type MangaCoversDataMangadex struct {
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

func getMangaCoversMangadex(_mangaTitle string, _mangaId string) {
  // TODO
  // - this breaks on one piece - 100+ covers
  // - automate saving covers in the first chapter of their respective volume
  var _limit = 100
  var url string = fmt.Sprintf("https://api.mangadex.org/cover?limit=%d&manga[]=%s&order[createdAt]=asc&order[updatedAt]=asc&order[volume]=asc", _limit, _mangaId)
  var mangaCoversData MangaCoversDataMangadex

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
  var dir string = fmt.Sprintf("%s%s/%s", downloadsRoot, authorSubDir, _mangaTitle)
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
          continue
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

type MangaChaptersMangadex struct {
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

func getMangaChaptersMangadex(_mangaInfo MangaDataMangadex) (MangaChaptersMangadex, error) {
  var queryLimit int = 500
  var offset int = 0
  var selectedLanguage string = "en"

  var url string = fmt.Sprintf("https://api.mangadex.org/manga/%s/feed?includeFuturePublishAt=0&limit=%d&offset=%d&contentRating[]=safe&contentRating[]=suggestive&contentRating[]=erotica&contentRating[]=pornographic&translatedLanguage[]=%s", _mangaInfo.Data.ID, queryLimit, offset, selectedLanguage)
  var mangaChapters MangaChaptersMangadex

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
      var mangaChapters2 MangaChaptersMangadex

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

  sort.Slice(mangaChapters.Data, func(i, j int) bool {
    var chapterA, chapterB string

    if mangaChapters.Data[i].Attributes.Chapter != nil {
      chapterA = *mangaChapters.Data[i].Attributes.Chapter
    }
    if mangaChapters.Data[j].Attributes.Chapter != nil {
      chapterB = *mangaChapters.Data[j].Attributes.Chapter
    }

    numA, _ := strconv.ParseFloat(chapterA, 64)
    numB, _ := strconv.ParseFloat(chapterB, 64)

    return numA < numB
  })

  return mangaChapters, nil
}

type MangaImagesMangadex struct {
  Result string `json:"result"`
  Chapter struct {
    Hash string `json:"hash"`
    Data []string `json:"data"`
  } `json:"chapter"`
}

func getMangaChapterImagesMangadex(_mangaTitle string, _mangaChapters MangaChaptersMangadex, _userInput string, _singleFolder bool) {
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
      var mangaChapterImages MangaImagesMangadex

      for {
        resp, err := http.Get(url)
        if err != nil {
          fmt.Println("Request error. Retrying.")
          continue
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
        // TODO - if "all" option, and name == nil then it is probably a oneshot. fix empty chapter number and set it to 0?
        var dir string
        if _singleFolder {
          dir = fmt.Sprintf("%s%s/%s", downloadsRoot, authorSubDir, _mangaTitle)
        } else if _userInput == "oneshot" {
          dir = fmt.Sprintf("%s%s/%s/Oneshot", downloadsRoot, authorSubDir, _mangaTitle)
        } else {
          if _mangaChapters.Data[i].Attributes.Title != nil && len(*_mangaChapters.Data[i].Attributes.Title) > 0 {
            dir = fmt.Sprintf("%s%s/%s/Ch.%s - %s", downloadsRoot, authorSubDir, _mangaTitle, chapterNameNoNIL, *_mangaChapters.Data[i].Attributes.Title)
          } else {
            dir = fmt.Sprintf("%s%s/%s/Ch.%s", downloadsRoot, authorSubDir, _mangaTitle, chapterNameNoNIL)
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
