package src

import (
  "encoding/json"
)

func PrettyPrintJson(_json interface{}) string {
  s, _ := json.MarshalIndent(_json, "", "\t")
  return string(s)
}

func contains(slice []string, element string) bool {
  for _, el := range slice {
    if el == element {
      return true
    }
  }
  return false
}
