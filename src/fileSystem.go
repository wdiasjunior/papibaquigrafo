package src

import (
  "fmt"
  "os"
  "strings"
  "io/ioutil"
)

func fsCreateDir(_dir string, _singleFolder bool) string {
  if _singleFolder {
    if _, err := os.Stat(_dir); os.IsNotExist(err) {
      err := os.MkdirAll(_dir, 0755)
      if err != nil {
        fmt.Println("Error creating directory:", err)
      }
    }
    return _dir
  }
  var dirVersion int = 2
  var stringDir string = _dir + " - V"
  for {
    if _, err := os.Stat(_dir); !os.IsNotExist(err) {
      _stringDir := fmt.Sprint(stringDir, dirVersion)
      if _, err := os.Stat(_stringDir); os.IsNotExist(err) {
        err := os.MkdirAll(_stringDir, 0755)
        if err != nil {
          fmt.Println("Error creating directory:", err)
        }
        return _stringDir
      } else {
        dirVersion += 1
      }
    } else {
      err := os.MkdirAll(_dir, 0755)
      if err != nil {
        fmt.Println("Error creating directory:", err)
      }
      return _dir
    }
  }
}

func fsCreateFile(_fileName string, _dir string, _fileNameNumber int, _fileContents []byte, _isCover bool, _coverName string) {
  fileExtension := fsFileExtension(_fileName)
  var fileName string
  if _fileNameNumber < 10 {
    if !_isCover {
      fileName = fmt.Sprintf("%s/00%d.%s", _dir, _fileNameNumber, fileExtension)
    } else {
      fileName = fmt.Sprintf("%s/%s.%s", _dir, _coverName, fileExtension)
    }
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
  } else if strings.Contains(_fileName, ".gif") {
    return "gif"
  } else {
    return "webp"
  }
}
