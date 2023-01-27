extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

// use std::io::{self, Write};
use std::collections::HashMap;

// use indicatif::ProgressBar;

// Currently only working for One Punch Man

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
          mangaChapters.push(ChapterParsed { title: &i.title, chapterImages: k.to_vec() });
        }
      }
    }
  }

  // return mangaChapters;
  getMangaChapterImages(mangaChapters);
}

fn getMangaChapterImages(_mangaChapters: Vec<ChapterParsed>) {
  for chapter in _mangaChapters {
    println!("_mangaChapters = {:?}\n", chapter.title);
  }
}

pub fn cubari(_gistURL: String) {
  getManga(_gistURL);
  // let mangaInfo = getManga(_gistURL);
  // getMangaChapterImages(mangaInfo);
}
