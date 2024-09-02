package src

import (
  "fmt"
  beeep "github.com/gen2brain/beeep"
)

// TODO
// - port to bubbletea?
// - add language selection in mangadex
// - add support for bato.to (and language selection)
// - add support for mangafire (and language selection)
// - add notifications?

// BUGS
// - mangadex - if chapter name is null or wahtever, skip chapter and want the the end which chapetrs failed

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
        err := beeep.Notify("Download Finished", "{MangaTitle} has finished downloading.", "")
        if err != nil {
          fmt.Println("Error sending notification.")
        }
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
