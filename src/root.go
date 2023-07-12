package src

import (
  "fmt"
)

func Execute() {
  fmt.Println("papibaquigrafo go.")
  fmt.Println("Choose an option: \n1: Mangadex \n2: TCB Scans \nquit")

  loop: for {
    fmt.Printf("-> ")
    var userInput string
    fmt.Scanf("%s", &userInput)

    switch userInput {
      case "1":
        fmt.Println("Mangadex")
        fmt.Println("\x1B[2J\x1B[1;1H") // clears terminal
        mangadex()
        break loop
      case "2":
        fmt.Println("TCB Scans")
        fmt.Println("\x1B[2J\x1B[1;1H") // clears terminal
        tcbscans()
        break loop
      case "quit":
        break loop
      default:
        fmt.Println("Invalid Option")
    }
  }
}

// TODO
// - use Bubble Tea for TUI
// - use tachiyomi user agent
// - 'asf' (tomo-chan alike. single page per chapter, so it saves everything in one directory)
// - prevent panic and crashes with loops and retries
// - better chapter selection
// -
