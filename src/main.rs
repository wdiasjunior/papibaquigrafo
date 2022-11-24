#![allow(non_snake_case)]

mod mangadex;

fn main() {
  // mangaId -> "d09c8abd-24ec-41de-ac8b-b6381a2f3a63"
  let mangaId = "d09c8abd-24ec-41de-ac8b-b6381a2f3a63".to_string();
  mangadex::mangadex(mangaId);
}

/*
  TODO
  - inspo -> https://github.com/metafates/mangal
  - tui -> every tui lib's docs sucks. use simple ass println and read_line? https://github.com/oli-obk/rust-si
    - eye candy -> papibaquigrafo ascii logo and colors?
    - input -> select mangadex or tcbScans
      - mangadex -> user enters manga id (implement search function in the future?)
      - tcbScans -> user selects which manga to download
    - progress bar
    - chapter selection
    - implement tcbScans scraper
      - selection of manga titles
  - generalized String arguments from this video -> https://www.youtube.com/watch?v=b0bgAYJDhhQ
*/
