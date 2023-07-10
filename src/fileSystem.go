package src

import (
  "fmt"
  // "errors"
  "os"
  "strings"
  "io/ioutil"
)

func fsCreateDir(_dir string) {
  var dirVersion int = 2
  var stringDir string = _dir + " - V"
  for {
    // if dirVersion >= 4 { // todo remove
    //   break
    // }

    if _, err := os.Stat(_dir); !os.IsNotExist(err) {
      _stringDir := fmt.Sprint(stringDir, dirVersion)
      if _, err := os.Stat(_stringDir); os.IsNotExist(err) {
        err := os.MkdirAll(_stringDir, 0755)
        if err != nil {
          fmt.Println("Error creating directory:", err)
        }
        break
      } else {
        dirVersion += 1
      }
    } else {
      err := os.MkdirAll(_dir, 0755)
      if err != nil {
        fmt.Println("Error creating directory:", err)
      }
      break
    }
  }
  return
}

func fsCreateFile(_fileName string, _dir string, _fileNameNumber int, _fileContents []byte) {
  fileExtension := fsFileExtension(_fileName)
  var fileName string
  if _fileNameNumber < 10 {
    fileName = fmt.Sprintf("%s/00%d.%s", _dir, _fileNameNumber, fileExtension)
  } else {
    fileName = fmt.Sprintf("%s/0%d.%s", _dir, _fileNameNumber, fileExtension)
  }
  err := ioutil.WriteFile(fileName, _fileContents, 0644)
  if err != nil {
    fmt.Println("Error writing to file:", err)
    return
  }
}

func fsFileExtension(_fileName string) string {
  if strings.Contains(_fileName, ".jpg") {
    return "jpg"
  } else if strings.Contains(_fileName, ".jpeg") {
    return "jpeg"
  } else if strings.Contains(_fileName, ".png") {
    return "png"
  } else {
    return "webp"
  }
}
