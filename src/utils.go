package src

import (
  "encoding/json"
  // "errors"
  // "fmt"
  // "log"
  // "net/http"
)

func PrettyPrintJson(_json interface{}) string {
  s, _ := json.MarshalIndent(_json, "", "\t")
  return string(s)
}

func contains(slice []string, element string, _isOneshot bool) bool {
  if _isOneshot {
    return false
  }
  for _, el := range slice {
    if el == element {
      return true
    }
  }
  return false
}

// func urlReq(url string) ([]byte, error) {
//   var responseBody []byte
//   var error = nill
//   for {
//     // retry as many times as necessary
//     // if error/timeout sleep for 10s println and retry
//
//     if err != nil {
//       fmt.Println("Could not get manga chapter")
//       log.Panic(err)
//       // return mangaChapterImages, errors.New("Could not get manga chapter")
//       // break
//     }
//     defer resp.Body.Close()
//     body, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//       log.Panic(err)
//       fmt.Println("Could not parse body - manga chapter")
//       // return mangaChapterImages, errors.New("Could not parse body")
//       // break
//     }
//   }
//   return responseBody, nil
// }
