package src

import (
  "encoding/json"
  "net/url"
  "fmt"
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

func reverseStringArray(arr []string) {
  n := len(arr)

  for i := 0; i < n/2; i++ {
    arr[i], arr[n-1-i] = arr[n-1-i], arr[i]
  }
}

func reverseStructStringArray(arr []ChapterBatoto) {
  n := len(arr)

  for i := 0; i < n/2; i++ {
    arr[i], arr[n-1-i] = arr[n-1-i], arr[i]
  }
}

func getHostAndReferer(imageURL string) string {
  parsedURL, err := url.Parse(imageURL)

  if err != nil {
    fmt.Println("Error parsing URL:", err)
    return ""
  }

  return fmt.Sprintf("https://%s/", parsedURL.Host)
}
