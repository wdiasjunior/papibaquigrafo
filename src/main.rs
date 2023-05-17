#![allow(non_snake_case)]
#[allow(dead_code)]

mod mangadex;
mod tcbscans;
mod cubari;
mod tonarinoyj;
mod comicojp;

// use std::io::{self, BufRead, Write};
use std::io::{self, Write};

fn main() {
  // mangaId -> "d09c8abd-24ec-41de-ac8b-b6381a2f3a63"
  // let mangaId = "d09c8abd-24ec-41de-ac8b-b6381a2f3a63".to_string();
  // mangadex::mangadex(mangaId);

  print!("Choose an option: \n1: Mangadex \n2: TCB Scans \n3: tonarinoyj.jp (One Punch Man) \n4: Comico.jp \n5: Cubari\n");
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

        // let userInput = "6a1d1cb1-ecd5-40d9-89ff-9d88e40b136b".to_string(); // tokyo ghoul
        // let userInput = "192aa767-2479-42c1-9780-8d65a2efd36a".to_string(); // Gachiakuta
        // let userInput = "eb0494de-3b43-4d52-a808-63429c4a4239".to_string(); // ako to bambi
        // let userInput = "76ee7069-23b4-493c-bc44-34ccbf3051a8".to_string(); // tomo chan // asf
        // let userInput = "ead4b388-cf7f-448c-aec6-bf733968162c".to_string(); // oneshot
        mangadex::mangadex(userInput);
        break;
      },
      "2" => {
        print!("\x1B[2J\x1B[1;1H"); // clears terminal
        println!("TCB Scans");
        tcbscans::tcbscans();
        break;
      },
      "3" => {
        print!("\x1B[2J\x1B[1;1H"); // clears terminal
        println!("tonarinoyj.jp");
        tonarinoyj::tonarinoyj();
        break;
      },
      "4" => {
        print!("\x1B[2J\x1B[1;1H"); // clears terminal
        println!("comico.jp");
        comicojp::comicojp();
        break;
      },
      "5" => {
        print!("\x1B[2J\x1B[1;1H"); // clears terminal
        println!("Cubari");
        print!("Enter the gist link: ");
        std::io::stdout().flush().expect("failed to flush stdout");
        let mut userInput = String::new();
        let stdin = io::stdin();
        stdin.read_line(&mut userInput).expect("Could not read line");
        // one punch man gist link
        // let userInput = "https://gist.githubusercontent.com/funkyhippo/1d40bd5dae11e03a6af20e5a9a030d81/raw/opm.json".to_string();
        cubari::cubari(userInput);
        break;
      },
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
    - progress bar - tcb scans and tonarinoyj
    - chapter selection
  - generalized String arguments from this video -> https://www.youtube.com/watch?v=b0bgAYJDhhQ
*/
