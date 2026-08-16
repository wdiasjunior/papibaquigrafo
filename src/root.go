package src

import (
  "fmt"
)

// TODO
// - port to bubbletea?
// - add language selection in mangadex
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

func Execute() {
  fmt.Println(`
papibaquigrafo.

Choose an option:
1: Mangadex
2: Weeb Central
3: TCB Scans
4: MangaFire
5: Mangabat
6: Quit
  `)

  loop: for {
    fmt.Printf("-> ")
    var userInput string
    fmt.Scanf("%s", &userInput)
    fmt.Println("\x1B[2J\x1B[1;1H")

    switch userInput {
      case "1":
        fmt.Println("Mangadex")
        notifyDownloadResult("Mangadex", mangadex())
        break loop
      case "2":
        fmt.Println("Weeb Central")
        notifyDownloadResult("Weeb Central", weebcentral())
        break loop
      case "3":
        fmt.Println("TCB Scans\n")
        notifyDownloadResult("TCB Scans", tcbscans())
        break loop
      case "4":
        fmt.Println("MangaFire")
        notifyDownloadResult("MangaFire", mangafire())
        break loop
      case "5":
        fmt.Println("Mangabat")
        notifyDownloadResult("Mangabat", mangabat())
        break loop
      case "6", "quit":
        break loop
      default:
        fmt.Println("Invalid Option")
    }
  }
}
