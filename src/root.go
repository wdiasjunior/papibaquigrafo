package src

import (
  "fmt"
)

// TODO
// - bubbletea
// - add language selection in mangadex
// - add support for mangasee
// - add support for bato.to
// - add support for mangafire (ana language selection)

// different project?
// tool that searches for scanlation groups annoying images and lists them in a ui to select which to delete

func Execute() {
  fmt.Println("papibaquigrafo.")
  fmt.Println("Choose an option: \n1: Mangadex \n2: Mangasee \n3: TCB Scans \n4: Bato.to \n5. Mangabat \nquit")

  loop: for {
    fmt.Printf("-> ")
    var userInput string
    fmt.Scanf("%s", &userInput)

    switch userInput {
      case "1":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Mangadex")
        mangadex()
        break loop
      case "2":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Mangasee")
        mangasee()
        break loop
      case "3":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("TCB Scans\n")
        tcbscans()
        break loop
      case "4":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Bato.to")
        batoto()
        break loop
      case "5":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Mangabat")
        mangabat()
        break loop
      case "quit":
        break loop
      default:
        fmt.Println("Invalid Option")
    }
  }
}
