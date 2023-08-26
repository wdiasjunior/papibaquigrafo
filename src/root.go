package src

import (
  "fmt"
)

func Execute() {
  fmt.Println("papibaquigrafo go.")
  fmt.Println("Choose an option: \n1: Mangadex \n2: TCB Scans \n3: Mangabat \nquit")

  loop: for {
    fmt.Printf("-> ")
    var userInput string
    fmt.Scanf("%s", &userInput)

    switch userInput {
      case "1":
        fmt.Println("\x1B[2J\x1B[1;1H") // clears terminal
        fmt.Println("Mangadex")
        mangadex()
        break loop
      case "2":
        fmt.Println("\x1B[2J\x1B[1;1H") // clears terminal
        fmt.Println("TCB Scans")
        tcbscans()
        break loop
      case "3":
        fmt.Println("\x1B[2J\x1B[1;1H") // clears terminal
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
