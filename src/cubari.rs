extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

// use std::io::{self, Write};
use std::{thread, time};
use std::collections::HashMap;

// use indicatif::ProgressBar;

// Currently only working for github gists with one translation group

//-------------------------------------------------------------------------------------------------

type MangaData = MangaDataJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
struct MangaDataJSONResponse {
  title: String,
  description: String,
  cover: String,
  #[serde(flatten)]
  chapters: HashMap<String, Chapters>,
}

#[derive(Debug, Deserialize, Serialize)]
struct Chapters {
  #[serde(flatten)]
  chapter: HashMap<String, Chapter>,
}

#[derive(Debug, Deserialize, Serialize)]
struct Chapter {
  last_updated: usize,
  title: String,
  volume: String,
  #[serde(flatten)]
  groups: HashMap<String, Groups>,
}

#[derive(Debug, Deserialize, Serialize)]
struct Groups {
  #[serde(flatten)]
  group: HashMap<String, Vec<String>>,
}

struct ChapterParsed<'a> {
  title: &'a String,
  chapterImages: Vec<String>,
}

// fn getManga(_gistURL: String) -> Vec<ChapterParsed<'static>> {
fn getManga(_gistURL: String) {
  let url = reqwest::Url::parse(&_gistURL).unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");
  // println!("json = {:?}", json);

  let mangaData: MangaData = serde_json::from_str(&json.to_string()).unwrap();
  // println!("mangaData = {:?}", mangaData.chapters[0].chapter.title);

  let mut mangaChapters: Vec<ChapterParsed> = Vec::new();

  // println!("mangaData = {:?}", mangaData.chapters);
  for chapter in mangaData.chapters.values() {
    for i in chapter.chapter.values() { // get chapter object
      // println!("title = {:?}\n", i.title); // get chapter title



      for j in i.groups.values() { // get group key
        for k in j.group.values() { // array of links for images
          // println!("groups = {:?}\n\n", k);
          // println!("chapter number = {:?}\n\n", &i);
          // mangaChapters.push(ChapterParsed { title: &i.title, chapterNumber: &i.key(), chapterImages: k.to_vec() });
          mangaChapters.push(ChapterParsed { title: &i.title, chapterImages: k.to_vec() });
        }
      }
    }
  }

  // return mangaChapters;
  getMangaChapterImages(mangaData.title, mangaChapters);
}

fn getMangaChapterImages(_mangaTitle: String, _mangaChapters: Vec<ChapterParsed>) {
  let mangaTitleDirectory = format!("downloads/{}", _mangaTitle);
  std::fs::create_dir_all(mangaTitleDirectory).unwrap();

  for chapter in _mangaChapters {
    let directory = format!("downloads/{}/{}/", _mangaTitle, chapter.title);
    std::fs::create_dir_all(&directory).unwrap();

    for (i, page) in chapter.chapterImages.iter().enumerate() {
      let fileExtension = if page.contains(".jpg") {
          "jpg"
        } else if page.contains(".jpeg") {
          "jpeg"
        } else {
          "png"
        };
      let fileName = if i + 1 < 10 {
          format!("{}/00{}.{}", &directory, i + 1, fileExtension)
        } else {
          format!("{}/0{}.{}", &directory, i + 1, fileExtension)
        };
      let mut file = std::fs::File::create(fileName).unwrap();

      let url = reqwest::Url::parse(&page.to_string()).unwrap();

      // let ten_millis = time::Duration::from_millis(2000);
      // thread::sleep(ten_millis);

      // let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();
      let _mangaImage = reqwest::blocking::get(url.clone());
      if let Err(e) = _mangaImage {
        // print!("erro {:?}\n\n", e);
        if e.is_timeout() {
          print!("timed out - chapter {} - page {}\n", chapter.title, i + 1);
          loop {
            thread::sleep(time::Duration::from_millis(15000));
            print!("retrying\n\n");
            let _mangaImage = reqwest::blocking::get(url.clone());
            if let Err(e) = _mangaImage {
              if e.is_timeout() {
                print!("timed out - chapter {} - page {}\n", chapter.title, i + 1);
              }
            } else {
              _mangaImage.unwrap().copy_to(&mut file).unwrap();
              break;
            }
          }
        }
      } else {
        _mangaImage.unwrap().copy_to(&mut file).unwrap();
      }
    }
  }
}

pub fn cubari(_gistURL: String) {
  getManga(_gistURL);
  // let mangaInfo = getManga(_gistURL);
  // getMangaChapterImages(mangaInfo);
}
