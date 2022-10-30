#![allow(non_snake_case)]

// use std::fs;
// use std::fs::File;
// use std::{io, thread, time};

mod mangadex;

fn main() {
  let mangaID = mangadex::getManga();
  let mangaChapters = mangadex::getMangaChapters(mangaID);
  // println!("mangaChapters = {:?}", mangaChapters);
  mangadex::getMangaChapterImages(mangaChapters);

}
