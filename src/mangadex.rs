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

    if _userInput.trim().eq("oneshot") || userInputVec.iter().any(|k| k.eq(&_mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string())) || _userInput.trim().eq("") {
      let directory = if _singleFolder && chapterImagesFileName.len() == 1 {
        if chapterImagesFileName.len() == 1 {
          format!("downloads/{}/", _mangaTitle)
        } else {
          format!("downloads/{}/Ch.{} - {}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string(), _mangaChapters.data[i].attributes.title.as_ref().expect("expect title not to be null").to_string())
        }
      } else if _userInput.trim().eq("oneshot") {
        format!("downloads/{}/Oneshot/", _mangaTitle)
      } else {
        match &_mangaChapters.data[i].attributes.title {
          Some(_) => format!("downloads/{}/Ch.{} - {}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string(), _mangaChapters.data[i].attributes.title.as_ref().expect("expect title not to be null").to_string()),
          None => format!("downloads/{}/Ch.{}/", _mangaTitle, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect chapter not to be null").to_string()),
          // None => format!("downloads/{}/Ch.{}/", _mangaTitle, "null chapter"), // if this is null, it's probably a oneshot
        }
      };
      let mut dirVersion = 2;
      let mut stringDir = "".to_string();
      stringDir.push_str(&directory);
      stringDir.pop();
      stringDir.push_str(" - V");
      loop {
        if _singleFolder {
          break;
        }
        if std::fs::metadata(&directory).is_ok() {
          stringDir.push_str(&dirVersion.to_string());
          if !std::fs::metadata(&stringDir).is_ok() {
            std::fs::create_dir_all(&stringDir).unwrap();
            break;
          } else {
            dirVersion += 1;
          }
        } else {
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
            // format!("{}/{}.{}", &directory, j + 1, fileExtension) // change this to just increment a number by the side of the chapter number
            format!("{}/{} - {}.{}", &directory, _mangaChapters.data[i].attributes.chapter.as_ref().expect("expect title not to be null").to_string(), j + 1, fileExtension)
          }
        } else {
          if j + 1 < 10 {
            format!("{}/00{}.{}", &directory, j + 1, fileExtension)
          } else {
            format!("{}/0{}.{}", &directory, j + 1, fileExtension)
          }
        };

        // let mut file = if _singleFolder && std::path::Path::new(&fileName).exists() {
        //   let mut chapterVersion = 2;
        //   let mut stringFile = "".to_string();
        //   stringFile.push_str(&directory);
        //   let charIndex = stringFile.chars().position(|c| c == '.').unwrap();
        //   stringFile[charIndex].push_str(" - V");
        //   loop {
        //     if _singleFolder {
        //       break;
        //     }
        //     if std::path::Path::new(&fileName).exists() {
        //       stringFile.push_str(&chapterVersion.to_string());
        //       if !std::path::Path::new(&stringFile).exists() {
        //         std::fs::File::create(stringFile)
        //         break;
        //       } else {
        //         chapterVersion += 1;
        //       }
        //     } else {
        //       // std::fs::create_dir_all(&directory).unwrap();
        //       std::fs::File::create(fileName)
        //       break;
        //     }
        //   }
        // } else {
        //   std::fs::File::create(fileName)
        // };
        let mut file = std::fs::File::create(fileName).unwrap(); // TODO - if file exists and _singleFolder, add a v2 at the end
        let baseUrl = format!("https://uploads.mangadex.org/data/{}/{}", hash, chapterImagesFileName[j]);
        let url = reqwest::Url::parse(&baseUrl).unwrap();
        let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();

        if j < chapterImagesFileName.len() - 1 {
          j += 1;
        } else {
          break 'j;
        }
      }
      // std::thread::sleep(std::time::Duration::from_millis(1000)); // debugger for directory versions
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

//-------------------------------------------------------------------------------------------------

pub fn mangadex(_mangaId: String) {
  let mangaInfo = getManga(_mangaId);
  if mangaInfo.result != "ok" {
    print!("error fecthing manga. error: {}", mangaInfo.result);
    return;
  }
  let mangaChapters = getMangaChapters(&mangaInfo);
  let mangaTitle = mangaInfo.data.attributes.title.en;
  print!("{}\n", mangaTitle);
  if mangaChapters.data.len() == 0 {
    print!("No chapters available\n");
    return;
  }
  print!("Number of chapters: {}\n\n", mangaChapters.data.len());

  let mut arr: Vec<_> = Vec::<_>::new();
  print!("available chapters:\n");
  if mangaChapters.data.len() == 1 && mangaChapters.data[0].attributes.chapter.as_ref().is_none() {
    print!("Oneshot\n");
  } else {
    for chapter in mangaChapters.data.iter() {
      if chapter.attributes.chapter.as_ref().is_some() {
        arr.push(chapter.attributes.chapter.as_ref().expect("expect title not to be null").to_string().parse::<f32>().unwrap());
      }
    }
    arr.sort_by(|a, b| a.partial_cmp(b).unwrap());
    for i in arr.iter() {
      print!("{}\n", i);
    }
  }

  print!("\nEnter the chapters you want to download\n");
  print!("Options: 'all', 'asf (all chapters in a single folder)', 'chapter numbers separated by spaces', 'oneshot', 'quit'\n");
  loop {
    print!("-> ");
    std::io::stdout().flush().expect("failed to flush stdout");

    let mut userInput = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut userInput).expect("Could not read line");

    match userInput.trim() {
      "all" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), false),
      "asf" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), true),
      "oneshot" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "oneshot".to_string(), true),
      "quit" => break,
      _ => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, userInput, false),
    }
    println!("\nDownload completed!\n");
    break;
  }
}

// TODO
// directory version
// oneshot
// 'asf' (tomo-chan alike. single page per chapter, so it saves everything in one directory)
// program should not panic
// fix chapter list ordering
// per chapter download bar of chapter images below chapter count download bar

// 192aa767-2479-42c1-9780-8d65a2efd36a  // Gachiakuta
// 76ee7069-23b4-493c-bc44-34ccbf3051a8  // Tomo-chan
// eb0494de-3b43-4d52-a808-63429c4a4239  // ako to bambi
// ead4b388-cf7f-448c-aec6-bf733968162c  // Hanabi - oneshot
