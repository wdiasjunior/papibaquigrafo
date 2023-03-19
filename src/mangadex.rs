extern crate reqwest;
extern crate serde;

use serde::Deserialize;
use serde::Serialize;
use serde_json::Value as JsonValue;

use std::io::{self, Write};

use indicatif::ProgressBar;

//-------------------------------------------------------------------------------------------------

type MangaData = MangaDataJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
struct MangaDataJSONResponse {
  result: String,
  data: Data,
}

#[derive(Debug, Deserialize, Serialize)]
struct Data {
  id: String,
  attributes: MangaAttributes,
}

#[derive(Debug, Deserialize, Serialize)]
struct MangaAttributes {
  title: Title,
  availableTranslatedLanguages: Vec<String>,
}

#[derive(Debug, Deserialize, Serialize)]
struct Title {
  en: String,
}

fn getManga(_mangaId: String) -> MangaData {

  let baseUrl = format!("https://api.mangadex.org/manga/{}", _mangaId);
  let url = reqwest::Url::parse(&baseUrl).unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mangaData: MangaData = serde_json::from_str(&json.to_string()).unwrap();

  // println!("mangaData = {:?}", mangaData.data.id);

  return mangaData;
}

//-------------------------------------------------------------------------------------------------

type MangaChapters = MangaChaptersJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
struct MangaChaptersJSONResponse {
  result: String,
  data: Vec<ChapterData>,
  total: i32,
}

#[derive(Debug, Deserialize, Serialize)]
struct ChapterData {
  id: String,
  attributes: ChapterAttributes,
}

#[derive(Debug, Deserialize, Serialize)]
struct ChapterAttributes {
  chapter: Option<String>,
  title: Option<String>,
}

fn getMangaChapters(_mangaInfo: &MangaData) -> MangaChapters {
  // println!("_mangaInfo.data.id = {:?}", _mangaInfo.data.id);

  let queryLimit: i32 = 500;
  let mut offset: i32 = 0;
  let selectedLanguage: String = "en".to_string();
  let baseUrl = format!("https://api.mangadex.org/manga/{}/feed?includeFuturePublishAt=0&limit={}&offset={}&translatedLanguage[]={}", _mangaInfo.data.id, queryLimit, offset, selectedLanguage);

  let url = reqwest::Url::parse(&baseUrl).unwrap();

  let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

  let mut mangaChapters: MangaChapters = serde_json::from_str(&json.to_string()).unwrap();

  if mangaChapters.total > queryLimit {
    offset = 500;
    let baseUrl = format!("https://api.mangadex.org/manga/{}/feed?includeFuturePublishAt=0&limit={}&offset={}&translatedLanguage[]={}", _mangaInfo.data.id, queryLimit, offset, selectedLanguage);
    let url = reqwest::Url::parse(&baseUrl).unwrap();
    let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");
    let parsed: MangaChapters = serde_json::from_str(&json.to_string()).unwrap();
    mangaChapters.data.extend(parsed.data.into_iter());
  }

  // println!("mangaChapters = {:?}", mangaChapters);
  return mangaChapters;
}

//-------------------------------------------------------------------------------------------------

type MangaImages = MangaImagesJSONResponse;

#[derive(Debug, Deserialize, Serialize)]
struct MangaImagesJSONResponse {
  result: String,
  chapter: ChapterImages,
}

#[derive(Debug, Deserialize, Serialize)]
struct ChapterImages {
  hash: String,
  data: Vec<String>,
}

fn getMangaChapterImages(_mangaTitle: String, _mangaChapters: &MangaChapters, _userInput: String, _singleFolder: bool) {
  let mut userInputVec: Vec<_> = [].to_vec();
  let progressBarLength = if !_userInput.trim().eq("") {
    let chapterSeletion = _userInput.trim().split(" ").map(|chapter| chapter.to_string());
    userInputVec = chapterSeletion.collect();
    userInputVec.len()
  } else {
    _mangaChapters.data.len()
  };

  println!("\nDownloading");
  let progressBar = ProgressBar::new(progressBarLength.try_into().unwrap());
  progressBar.inc(0);

  let mangaTitleDirectory = format!("downloads/{}", _mangaTitle);
  std::fs::create_dir_all(mangaTitleDirectory).unwrap();

  let mut i: usize = 0;
  let mut k: usize = 0;
  'i: loop {
    let baseUrl = format!("https://api.mangadex.org/at-home/server/{}", _mangaChapters.data[i].id);

    let url = reqwest::Url::parse(&baseUrl).unwrap();

    let json: JsonValue = reqwest::blocking::get(url).expect("bad request").json().expect("error parsing json");

    let mangaChapterImages: MangaImages = serde_json::from_str(&json.to_string()).unwrap();

    let hash: String = mangaChapterImages.chapter.hash;
    let chapterImagesFileName: Vec<String> = mangaChapterImages.chapter.data;

    if userInputVec.iter().any(|k| k.eq(&_mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string())) || _userInput.trim().eq("") {
      let directory = if _singleFolder && chapterImagesFileName.len() == 1 {
        if chapterImagesFileName.len() == 1 {
          format!("downloads/{}/", _mangaTitle)
        } else {
          format!("downloads/{}/Ch.{} - {}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string(), _mangaChapters.data[i].attributes.title.as_ref().expect("expect title not to be null").to_string())
        }
      } else {
        match &_mangaChapters.data[i].attributes.title {
          Some(_) => format!("downloads/{}/Ch.{} - {}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string(), _mangaChapters.data[i].attributes.title.as_ref().expect("expect title not to be null").to_string()),
          // None => format!("downloads/{}/Ch.{}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string()),
          None => format!("downloads/{}/Ch.{}/", _mangaTitle, "null chapter"), // if this is null, it's probably a oneshot
        }
      };
      let mut dirVersion = 2;
      // let mut shouldHaveVersion = false;
      // let chapterVersion: String = &directory.push_str(format!(" - V{}", dirVersion).to_string()).clone().to_string();
      // let mut stringDir = String::new();
      let mut stringDir = "".to_string();
      stringDir.push_str(&directory);
      stringDir.push_str(" - V");
      let test = &mut stringDir.clone();
///////////////////////////////////////////////////////////////////////////////////////////////////
      loop { // change this to just increment a number by the side of the chapter number
        if std::fs::metadata(&directory).is_ok() && chapterImagesFileName.len() > 1 {
          // println!("inside loop if");
          // shouldHaveVersion = true;
          // println!("v2 dir");
          // let mut stringDir = String::new();
          // stringDir.push_str(&directory);
          // stringDir.push_str(" - V");
          stringDir.push_str(&dirVersion.to_string());
          if !std::fs::metadata(&stringDir).is_ok() {
            // println!("{}", &stringDir);
            std::fs::create_dir_all(&test).unwrap();
            // shouldHaveVersion = false;
            // dirVersion = 1;
            break;
          } else {
            dirVersion += 1;
          }
        } else {
          // println!("inside loop else");
          std::fs::create_dir_all(&directory).unwrap();
          break;
        }
      }
///////////////////////////////////////////////////////////////////////////////////////////////////
      let mut j: usize = 0;
      'j: loop {
        let fileExtension = if chapterImagesFileName[j].contains(".jpg") {
          "jpg"
        } else if chapterImagesFileName[j].contains(".jpeg") {
          "jpeg"
        } else {
          "png"
        };

        let fileName = if _singleFolder {
          if chapterImagesFileName.len() == 1 {
            format!("{}/{}.{}", &directory, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect title not to be null").to_string(), fileExtension)
          } else {
            format!("{}/{}.{}", &directory, j + 1, fileExtension) // change this to just increment a number by the side of the chapter number
          }
        } else {
          if j + 1 < 10 {
            format!("{}/00{}.{}", &directory, j + 1, fileExtension)
          } else {
            format!("{}/0{}.{}", &directory, j + 1, fileExtension)
          }
        };

        let mut file = std::fs::File::create(fileName).unwrap();
        let baseUrl = format!("https://uploads.mangadex.org/data/{}/{}", hash, chapterImagesFileName[j]);
        let url = reqwest::Url::parse(&baseUrl).unwrap();
        let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();

        if j < chapterImagesFileName.len() - 1 {
          j += 1;
        } else {
          break 'j;
        }
      }
      if k < progressBarLength - 1 {
        k += 1;
        progressBar.inc(1);
      } else {
        progressBar.finish_with_message("done");
        break 'i;
      }
    }
    i += 1;
  }
}

pub fn mangadex(_mangaId: String) {
  let mangaInfo = getManga(_mangaId);
  let mangaChapters = getMangaChapters(&mangaInfo);
  // print!("mangaChapters length {}", mangaChapters.data.len());
  let mangaTitle = mangaInfo.data.attributes.title.en;
  // getMangaChapterImages(mangaTitle.to_string(), mangaChapters);

  let mut arr: Vec<_> = Vec::<_>::new();
  print!("available chapters:\n");
  for chapter in mangaChapters.data.iter() {
    // print!("{}\n", chapter.attributes.chapter);
    if chapter.attributes.chapter.as_ref().is_some() {
      arr.push(chapter.attributes.chapter.as_ref().expect("expect title not to be null").to_string());
    }
  }
  // arr.sort();
  arr.sort_by(|a, b| a.partial_cmp(b).unwrap()); // this is fucking ridiculous, Rust. just sort the goddamn float vector by default.
  for i in arr.iter() {
    print!("{}\n", i);
  }

  print!("\nEnter the chapters you want to download\n");
  print!("Options: 'all', 'asf (all chapters in a single folder)', 'chapter numbers separated by spaces', 'quit'\n");
  loop {
    print!("-> ");
    std::io::stdout().flush().expect("failed to flush stdout");

    let mut userInput = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut userInput).expect("Could not read line");

    match userInput.trim() {
      "all" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), false),
      "asf" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), true),
      "quit" => break,
      _ => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, userInput, false), // println!("userInput {}", userInput),
    }
    println!("\nDownload completed!\n");
    break;
  }
}
// 192aa767-2479-42c1-9780-8d65a2efd36a  // Gachiakuta id for testing
// 76ee7069-23b4-493c-bc44-34ccbf3051a8  // Tomo-chan id for testing
