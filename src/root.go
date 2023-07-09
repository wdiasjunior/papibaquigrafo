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
    // fmt.Printf("---> %s\n", userInput)
  }
}

// TODO
// - rewrite in go
// - use Bubble Tea for TUI
// - use tachiyomi user agent
// - make a dedicated file to handle file system operations
//     - directory version
//     - don't let program panic if there is no chaper name, leave just the number and version
//     - oneshot
//     - 'asf' (tomo-chan alike. single page per chapter, so it saves everything in one directory)
// - fix bugs in rust version
// - prevent panic and crashes with loops and retries
// - better chapter selection
// -


// 192aa767-2479-42c1-9780-8d65a2efd36a  // Gachiakuta
// 76ee7069-23b4-493c-bc44-34ccbf3051a8  // Tomo-chan - asf
// eb0494de-3b43-4d52-a808-63429c4a4239  // ako to bambi
// ead4b388-cf7f-448c-aec6-bf733968162c  // Hanabi - oneshot
// 239d6260-d71f-43b0-afff-074e3619e3de  // bleach
// d0c88e3b-ea64-4e07-9841-c1d2ac982f4a  // dagashi kashi
// f0a682dd-38dc-4d51-8469-e6ed181766e4  // kirara
// 6fef1f74-a0ad-4f0d-99db-d32a7cd24098  // fire punch
