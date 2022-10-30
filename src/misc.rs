extern crate reqwest;
extern crate serde;

use serde_json::Value as JsonValue;

// use serde::Deserialize;
// use serde::Serialize;
// use std::collections::HashMap;
// use serde_json::{Result, Value};

// #[derive(Debug, Deserialize)]
// #[serde(rename_all = "lowercase")]
// pub struct Manga {
//   pub title: String,
// }
//
// #[derive(Debug, Deserialize)]
// pub struct MangaData {
//   pub manga: Manga,
//   pub chapter: HashMap<String, Chapter>,
// }
//
// #[derive(Debug, Deserialize)]
// pub struct Chapter {
//   pub volume: String,
//   pub chapter: String,
//   pub lang_code: String,
//   pub group_name: String,
// }

// #[derive(Debug, Deserialize)]
// pub struct ChapterData {
//   pub id: MangaDatau32,
//   pub hash: String,
//   pub manga_id: u32,
//   pub server: String,
//   pub page_array: Vec<String>
// }

// #[derive(Serialize, Deserialize, Debug)]
// struct JSONResponse {
//   json: HashMap<String, String>,
// }

pub fn getManga() {
// pub async fn getManga() -> Result<()> {
// pub fn getManga(client: &reqwest::blocking::Client) {
  // let client = reqwest::Client::new();
  // println!("papibaquigrafo");
  // String::from("");
  // let selectedLanguage = "en".to_string();
  //
  // let mockMangaID = "192aa767-2479-42c1-9780-8d65a2efd36a";
  // let mockMangaChapterID: String = "e543ecb3-17a0-452b-8dda-e01c5837453f";
  // let mockMangaPage: String = "https://uploads.mangadex.org/data/44ec8d22d24f219fecd22512ee590a11/2-8d3cfb43f34456c70a37b5f8858a377b6fa8911aafc026bebea52697b8983ea5.jpg";
  // // let id: String = _id;
  //
  // let getMangaInfo = fetch("https://api.mangadex.org/manga/{id}/");
  // // getMangaInfo.attributes.title -> object with keys for different languages
  // // getMangaInfo.attributes.availableTranslatedLanguages
  // // getMangaInfo.attributes.hasAvailableChapters
  //
  // let getMangaChaptersIDs = fetch("https://api.mangadex.org/manga/{id}/feed?includeFuturePublishAt=0");
  // // fetch with limit parameter set to 2000 ?
  // // and translatedLanguage parameter set to selectedLanguage if getMangaInfo.attributes.availableTranslatedLanguages includes selectedLanguage
  // // data[i].id
  // // data[i].type == chapter
  // // data[i].attributes.chapter
  // // data[i].attributes.title
  // // data[i].attributes.translatedLanguage == selectedLanguage ?
  // // data[i].attributes.version ?
  //
  //
  // let getMangaChapterById = fetch("https://api.mangadex.org/at-home/server/{mangaChapterID}");
  // // getMangaChapterById.chapter.hash -> data necessary to retrieve chapter images
  // // getMangaChapterById.chapter.data -> arrary of files for the chapter
  //
  // let getMangaPageByFileName = fetch("https://uploads.mangadex.org/data/{hash}/{fileName}")

  // let url = reqwest::Url::parse(&mockMangaPage).unwrap();
  let url = reqwest::Url::parse("https://api.mangadex.org/manga/192aa767-2479-42c1-9780-8d65a2efd36a").unwrap();
  // // let url = baseUrl.join(mockMangaID);

  // let json: MangaData = client.get(url).send().expect("bad request").json().expect("error parsing json");
  // // get with parameters
  // std::thread::sleep(std::time::Duration::from_secs(1));
  // print!("{:?}", json);

  // let res = reqwest::blocking::get(url).send().expect("bad request").json().expect("error parsing json");
  // let res = client.get(url).text();
  // println!("Response Body: {:?}", res);
  // let json: JSONResponse = serde_json::from_str(&res).unwrap();
  // print!("{:?}", json);


  // let query = vec![
  //   ("limit", 200),
  //   ("translatedLanguage", selectedLanguage),
  // ];
  // .query(&query).send()

  let body: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");
  println!("body = {:?}", body);

  // Ok(())
}
