package src

import (
  "fmt"
  beeep "github.com/gen2brain/beeep"
)

// TODO
// - port to bubbletea?
// - add language selection in mangadex
// - add support for mangafire (and language selection)
// - add notifications
// - turn this into a server with web ui for remote and queued downloads
//
////////////////////////////////////////////////////////////////////////////////
//
// turn it into a web server
//
// running on docker on mariola
//
// integrate into tsuuchi
//
// rewrite chapter selection of every site connector
// make stuff more modular?
//
// weeb.wdias.dev/papibaquigrafo
//
// {
//   provider: mangadex
//   mangaID: hdvyaibavhauwu
// }
//
// {
//   downloadCovers: true
//   authorSubfolder: true
//   chapterRange: false
//   initialChapter: 1
//   lastChapter: 100
//   chapterSelection: 1 2 3 4, all, oneshot
// }
//
////////////////////////////////////////////////////////////////////////////////
//
// BUGS
// - mangadex - if chapter name is null or whatever, skip chapter and list at the end which chapters failed
//
// different project?
// tool that searches for scanlation groups annoying images and lists them in a ui to select which to delete

func Execute() {
  fmt.Println(`
papibaquigrafo.

Choose an option:
1: Mangadex
2: Weeb Central
3: TCB Scans
4: Bato.to
5: Mangabat
6: Quit
  `)

  loop: for {
    fmt.Printf("-> ")
    var userInput string
    fmt.Scanf("%s", &userInput)

    switch userInput {
      case "1":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Mangadex")
        mangadex()
        // TODO
        // - top level functions should return a result code and the notification is handled here or in another file
        err := beeep.Notify("Download Finished", "{MangaTitle} has finished downloading.", "")
        if err != nil {
          fmt.Println("Error sending notification.")
        }
        break loop
      case "2":
        fmt.Println("\x1B[2J\x1B[1;1H")
        fmt.Println("Weeb Central")
        weebcentral()
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
        if !true { mangabat() }
        fmt.Println("TODO - Mangabat download is currently broken")
        break loop
      case "6", "quit":
        break loop
      default:
        fmt.Println("Invalid Option")
    }
  }
}
