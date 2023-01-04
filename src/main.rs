#![allow(non_snake_case)]

mod mangadex;

use std::io::{self, BufRead, Write};

fn main() {
  // mangaId -> "d09c8abd-24ec-41de-ac8b-b6381a2f3a63"
  // let mangaId = "d09c8abd-24ec-41de-ac8b-b6381a2f3a63".to_string();
  // mangadex::mangadex(mangaId);

  print!("Choose an option: \n1: Mangadex \n2: TCB Scans\n");
  loop {
    print!("-> ");
    std::io::stdout().flush().expect("failed to flush stdout");

    let mut userInput = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut userInput).expect("Could not read line");

    match userInput.trim() {
      "1" => {
        print!("\x1B[2J\x1B[1;1H"); // clears terminal
        println!("Mangadex");
        print!("Enter the Manga ID: ");
        std::io::stdout().flush().expect("failed to flush stdout");
        let mut userInput = String::new();
        let stdin = io::stdin();
        stdin.read_line(&mut userInput).expect("Could not read line");
        mangadex::mangadex(userInput);
        println!("Download completed!\n");
      },
      "2" => println!("TODO"),
      "quit" => break,
      _ => println!("Invalid option"),
    }
  }
}

/*
  TODO
  - inspo -> https://github.com/metafates/mangal
  - tui -> every tui lib's docs sucks. use simple ass println and read_line? https://github.com/oli-obk/rust-si
    - eye candy -> papibaquigrafo ascii logo and colors?
    - input -> select mangadex or tcbScans
      - mangadex -> user enters manga id (implement search function in the future?)
        - 4komas/single page should download every chapter into the same folder
      - tcbScans -> user selects which manga to download
        - https://github.com/manga-download/hakuneko/blob/dd59107eb47e5c0b6a2a4211e231def4cd6ebde8/src/web/mjs/connectors/TCBScans.mjs
        - https://onepiecechapters.com/mangas/5/one-piece
    - progress bar
    - chapter selection
    - implement tcbScans scraper
      - selection of manga titles
  - generalized String arguments from this video -> https://www.youtube.com/watch?v=b0bgAYJDhhQ
*/
