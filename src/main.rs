#![allow(non_snake_case)]

// use std::fs;
// use std::fs::File;
// use std::{io, thread, time};

mod mangadex;

fn main() {
  let mangaInfo = mangadex::getManga();
  let mangaChapters = mangadex::getMangaChapters(&mangaInfo);
  let mangaTitle = mangaInfo.data.attributes.title.en;
  mangadex::getMangaChapterImages(mangaTitle.to_string(), mangaChapters);

}
