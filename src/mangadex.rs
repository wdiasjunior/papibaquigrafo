extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

//-------------------------------------------------------------------------------------------------

pub type MangaData = MangaDataJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaDataJSONResponse {
  result: String,
  pub data: Data,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Data {
  id: String,
  pub attributes: MangaAttributes,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaAttributes {
  pub title: Title,
  availableTranslatedLanguages: Vec<String>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Title {
  pub en: String,
}

pub fn getManga() -> MangaData {
  let url = reqwest::Url::parse("https://api.mangadex.org/manga/192aa767-2479-42c1-9780-8d65a2efd36a").unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mangaData: MangaData = serde_json::from_str(&json.to_string()).unwrap();

  // println!("mangaData = {:?}", mangaData.data.id);

  return mangaData;
}

//-------------------------------------------------------------------------------------------------

pub type MangaChapters = MangaChaptersJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaChaptersJSONResponse {
  result: String,
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
  // translatedLanguage: String,
}

pub fn getMangaChapters(_mangaInfo: &MangaData) -> MangaChapters {
  // println!("_mangaInfo.data.id = {:?}", _mangaInfo.data.id);

  let queryLimit: i32 = 500; // the limit is 500 - if manga has more than 500 chapters use the offset parameter
  let selectedLanguage: String = "en".to_string();
  let baseUrl = format!("https://api.mangadex.org/manga/{}/feed?includeFuturePublishAt=0&limit={}&translatedLanguage[]={}", _mangaInfo.data.id, queryLimit, selectedLanguage);

  let url = reqwest::Url::parse(&baseUrl).unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mangaChapters: MangaChapters = serde_json::from_str(&json.to_string()).unwrap();

  // println!("mangaChapters = {:?}", mangaChapters);
  return mangaChapters
}

//-------------------------------------------------------------------------------------------------

pub type MangaImages = MangaImagesJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
pub struct MangaImagesJSONResponse {
  result: String,
  chapter: ChapterImages,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct ChapterImages {
  hash: String,
  data: Vec<String>,
}

pub fn getMangaChapterImages(_mangaTitle: String, _mangaChapters: MangaChapters) {
  let mut i: usize = 0;
  loop {

    let baseUrl = format!("https://api.mangadex.org/at-home/server/{}", _mangaChapters.data[i].id);

    let url = reqwest::Url::parse(&baseUrl).unwrap();

    let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

    let mangaChapterImages: MangaImages = serde_json::from_str(&json.to_string()).unwrap();

    let hash: String = mangaChapterImages.chapter.hash;
    let chapterImagesFileName: Vec<String> = mangaChapterImages.chapter.data;

    let mut j: usize = 0;
    loop {
      let mangaTitleDirectory = format!("downloads/{}", _mangaTitle);
      std::fs::create_dir_all(mangaTitleDirectory).unwrap();
      let directory = match &_mangaChapters.data[i].attributes.title {
        Some(_) => format!("downloads/{}/Ch.{} - {}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter, _mangaChapters.data[i].attributes.title.as_ref().expect("expect title not to be null").to_string()),
        None => format!("downloads/{}/Ch.{}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter),
      };
      std::fs::create_dir_all(&directory).unwrap();
      let fileExtension = if chapterImagesFileName[j].contains(".jpg") {
        "jpg"
      } else if chapterImagesFileName[j].contains(".jpeg") {
        "jpeg"
      } else {
        "png"
      };
      let fileName = format!("{}/0{}.{}", &directory, j + 1, fileExtension);
      let mut file = std::fs::File::create(fileName).unwrap();

      let baseUrl = format!("https://uploads.mangadex.org/data/{}/{}", hash, chapterImagesFileName[j]);

      let url = reqwest::Url::parse(&baseUrl).unwrap();

      let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();

      if j < chapterImagesFileName.len() - 1 {
        j += 1;
      } else {
        break;
      }
    }

    if i < _mangaChapters.data.len() - 1 {
      i += 1;
    } else {
      break;
    }
  }
}
