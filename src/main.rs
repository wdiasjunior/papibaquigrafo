#![allow(non_snake_case)]

mod mangadex;

fn main() {
  mangadex::mangadex();
}

/*
  TODO
  - inspo -> https://github.com/metafates/mangal
  - cli -> https://rust-cli.github.io/book/index.html
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
