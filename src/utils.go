package src

import (
  "encoding/json"
)

func PrettyPrintJson(_json interface{}) string {
  s, _ := json.MarshalIndent(_json, "", "\t")
  return string(s)
}
