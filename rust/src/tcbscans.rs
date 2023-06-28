extern crate reqwest;
extern crate serde;

use regex::Regex;

use std::io::{self, Write};

fn getMangaList() -> Vec<String> {
  let baseUrl = "https://onepiecechapters.com/projects";
  let url = reqwest::Url::parse(&baseUrl).unwrap();
  let json = reqwest::blocking::get(url).expect("bad request").text().unwrap();

  let mangaList: Vec<String> = Regex::new(r#"/mangas/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();
  let mut filteredMangaList: Vec<String> = Vec::new();
  let mut n = 0;

  for manga in &mangaList {
    if n % 2 == 0 {
      print!("{} {}\n", n / 2, manga);
      filteredMangaList.push(manga.to_string());
    }
    n += 1;
  }
  return filteredMangaList;
}

fn getChapterList(_manga: String) -> Vec<String> {
  print!("\x1B[2J\x1B[1;1H"); // clears terminal
  print!("{} \n\n", _manga);

  let baseUrl = format!("https://onepiecechapters.com{}", _manga);
  let url = reqwest::Url::parse(&baseUrl).unwrap();
  let json = reqwest::blocking::get(url).expect("bad request").text().unwrap();

  let chapterListDesc: Vec<String> = Regex::new(r#"/chapters/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();
  let mut chapterList: Vec<String> = chapterListDesc;
  chapterList.reverse();
  // for chapter in &chapterList {
  //   print!("{}\n", chapter);
  // }
  return chapterList;
}

fn getChapterImages(_mangaTitle: String, _mangaChapter: String) {
  // print!("\x1B[2J\x1B[1;1H"); // clears terminal
  // print!("{} \n\n", _mangaChapter);

  let baseUrl = format!("https://onepiecechapters.com{}", &_mangaChapter);
  let url = reqwest::Url::parse(&baseUrl).unwrap();
  let json = reqwest::blocking::get(url).expect("bad request").text().unwrap();

  let chapterImages: Vec<String> = Regex::new(r#"https://cdn.onepiecechapters.com/file/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();

  let mangaTitle: Vec<String> =  Regex::new(r#"[^/]*$"#).unwrap().find_iter(&_mangaTitle).map(|e| e.as_str().to_string().replace("-", " ")).collect();
  // let mangaChapter: Vec<String> = Regex::new(r#"/chapters/\d+/one-piece-chapter-([0-9.]+|-new)"#).unwrap().find_iter(&_mangaChapter).map(|e| e.as_str().to_string()).collect();
  // Regex::new(r#"\d+$"#)
  let re = if _mangaChapter.contains("one-piece") {
    Regex::new(r#"/chapters/\d+/one-piece-chapter-([0-9.]+|-new)"#).unwrap()
  } else {
    Regex::new(r#"/chapters/\d+/attack-on-titan-chapter-([0-9.]+|-new)"#).unwrap()
  };

  let mut mangaChapter: String = "".to_string();
  if let Some(captures) = re.captures(&_mangaChapter) {
    if captures.get(2).is_some() {
      // let _auxString = captures.get(1).unwrap().as_str().to_owned();
      // mangaChapter = _auxString.push_str(".".to_owned() + captures.get(2).unwrap().as_str());
      mangaChapter = format!("{}.{}", captures.get(1).unwrap().to_owned().as_str(), captures.get(2).unwrap().to_owned().as_str());
    } else {
      mangaChapter = format!("{}", captures.get(1).unwrap().to_owned().as_str());
    }

    println!("{}", mangaChapter);
  }

  // print!("mangaTitle {}\n", mangaTitle[0]);
  // print!("mangaChapter {}\n", mangaChapter);

  let directory = format!("downloads/{}/Ch.{}/", mangaTitle[0], mangaChapter);
  std::fs::create_dir_all(&directory).unwrap();

  let mut i = 0;
  for image in &chapterImages {
    print!("{}\n", image);
    let fileExtension = if image.contains(".jpg") {
      "jpg"
    } else if image.contains(".jpeg") {
      "jpeg"
    } else if image.contains(".png") {
      "png"
    } else {
      "webp"
    };
    let fileName = if i + 1 < 10 {
      format!("{}/00{}.{}", &directory, i + 1, fileExtension)
    } else {
      format!("{}/0{}.{}", &directory, i + 1, fileExtension)
    };
    let mut file = std::fs::File::create(fileName).unwrap();
    let url = reqwest::Url::parse(&image).unwrap();
    let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();
    i += 1;
  }
}

pub fn tcbscans() {
  let _mangaList = getMangaList();

  print!("Enter the Manga ID: \n");
  print!("-> ");
  std::io::stdout().flush().expect("failed to flush stdout");
  let mut userInput = String::new();
  let stdin = io::stdin();
  stdin.read_line(&mut userInput).expect("Could not read line");

  let mangaIndex: usize = userInput.trim().parse().unwrap();
  let _mangaTitle = _mangaList[mangaIndex].clone();
  let _chapterList = getChapterList(_mangaList[mangaIndex].clone());

  print!("initial chapter -> ");
  std::io::stdout().flush().expect("failed to flush stdout");
  let mut userInput = String::new();
  let stdin = io::stdin();
  stdin.read_line(&mut userInput).expect("Could not read line");

  let chapterIndex: usize = userInput.trim().parse().unwrap();

  print!("last chapter -> ");
  std::io::stdout().flush().expect("failed to flush stdout");
  let mut userInput = String::new();
  let stdin = io::stdin();
  stdin.read_line(&mut userInput).expect("Could not read line");

  let chapterIndexLimit: usize = userInput.trim().parse().unwrap();

  // print!("\x1B[2J\x1B[1;1H"); // clears terminal
  for (i, chapter) in _chapterList.iter().enumerate() {
  // for chapter in _chapterList {
    if i >= chapterIndex - 1 && i <= chapterIndexLimit - 1 {
      // print!("{}\n", chapter);
      getChapterImages(_mangaTitle.clone(), chapter.to_string());
      // getChapterImages(_mangaTitle, _chapterList[8].clone());
    }
  }
  println!("\nDownload completed!\n");
}

// TODO
// - make it work for attack on titan
