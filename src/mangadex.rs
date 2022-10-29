extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

type MangaData = JSONResponse;

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
struct JSONResponse {
  data: Data,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
struct Data {
  id: String,
  attributes: Attributes,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
struct Attributes {
  title: Title,
  availableTranslatedLanguages: Vec<String>,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
struct Title {
  en: String,
}

pub fn getManga() {

  let url = reqwest::Url::parse("https://api.mangadex.org/manga/192aa767-2479-42c1-9780-8d65a2efd36a").unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  // let jsonData: JsonValue = serde_json::from_str(&json.to_string()).unwrap();
  let jsonData: MangaData = serde_json::from_str(&json.to_string()).unwrap();
  // let mangaData: MangaData = serde_json::from_str(&jsonData.to_string()).unwrap();

  println!("mangaData = {:?}", jsonData);
}
