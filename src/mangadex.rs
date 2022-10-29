extern crate reqwest;
extern crate serde;

use serde::Deserialize;
// use std::collections::HashMap;

#[derive(Debug, Deserialize)]
pub struct ChapterData {
  pub id: u32,
  pub hash: String,
  pub manga_id: u32,
  pub server: String,
  pub page_array: Vec<String>
}

pub async fn getManga(client: &reqwest::blocking::Client) {
  // println!("papibaquigrafo");

  // let selectedLanguage: String = "en";
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
  // let url = baseUrl.join(mockMangaID);

  let json: ChapterData = client.get(url).send().expect("bad request").json().expect("error parsing json");
  std::thread::sleep(std::time::Duration::from_secs(1));
  print!("{:?}", json);
}
