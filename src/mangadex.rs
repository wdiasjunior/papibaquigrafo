extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

//-------------------------------------------------------------------------------------------------

pub type MangaData = MangaDataJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaDataJSONResponse {
  data: Data,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Data {
  id: String,
  attributes: MangaAttributes,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaAttributes {
  title: Title,
  availableTranslatedLanguages: Vec<String>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Title {
  en: String,
}

pub fn getManga() -> String {
  let url = reqwest::Url::parse("https://api.mangadex.org/manga/192aa767-2479-42c1-9780-8d65a2efd36a").unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mangaData: MangaData = serde_json::from_str(&json.to_string()).unwrap();

  // println!("mangaData = {:?}", mangaData.data.id);

  return mangaData.data.id;
}

//-------------------------------------------------------------------------------------------------

pub type MangaChapters = MangaChaptersJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaChaptersJSONResponse {
  data: Vec<ChapterData>,
  total: i32,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct ChapterData {
  id: String,
  attributes: ChapterAttributes,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct ChapterAttributes {
  chapter: String,
  title: Option<String>,
  translatedLanguage: String,
}

pub fn getMangaChapters(_mangaID: String) -> MangaChapters {
  // println!("_mangaID = {:?}", _mangaID);

  let queryLimit: i32 = 500; // the limit is 500 - if manga has more than 500 chapters use the offset parameter
  let selectedLanguage: String = "en".to_string();
  let baseUrl = format!("https://api.mangadex.org/manga/{}/feed?includeFuturePublishAt=0&limit={}&translatedLanguage[]={}", _mangaID, queryLimit, selectedLanguage);

  let url = reqwest::Url::parse(&baseUrl).unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mangaChapters: MangaChapters = serde_json::from_str(&json.to_string()).unwrap();

  // println!("mangaChapters = {:?}", mangaChapters);
  return mangaChapters
}

//-------------------------------------------------------------------------------------------------

pub fn getMangaChapterImages(_mangaChapters: MangaChapters) {
  let mut i: usize = 0;
  loop {
    println!("{:?}", _mangaChapters.data[i].id);
    // println!("{:?}", _mangaChapters.data[i].attributes.chapter);

    i += 1;
  }
}
