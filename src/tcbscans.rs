extern crate reqwest;
extern crate serde;

use regex::Regex;

use std::io::{self, Write};

// use indicatif::ProgressBar;

fn getMangaList() -> Vec<String> {
  let baseUrl = "https://onepiecechapters.com/projects";
  let url = reqwest::Url::parse(&baseUrl).unwrap();
  let json = reqwest::blocking::get(url).expect("bad request").text().unwrap();

  let mangaList: Vec<String> = Regex::new(r#"/mangas/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();
  let mut filteredMangaList: Vec<String> = Vec::new();
  let mut n = 0;
  print!("Enter the Manga ID: \n");
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

  let chapterList: Vec<String> = Regex::new(r#"/chapters/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();
  for chapter in &chapterList {
    print!("{}\n", chapter);
  }
  return chapterList;
}

fn getChapterImages(_mangaTitle: String, _mangaChapter: String) {
  print!("\x1B[2J\x1B[1;1H"); // clears terminal
  print!("{} \n\n", _mangaChapter);

  let baseUrl = format!("https://onepiecechapters.com{}", &_mangaChapter);
  let url = reqwest::Url::parse(&baseUrl).unwrap();
  let json = reqwest::blocking::get(url).expect("bad request").text().unwrap();

  let chapterImages: Vec<String> = Regex::new(r#"https://cdn.onepiecechapters.com/file/(.*?)[^""]+"#).unwrap().find_iter(&json).map(|e| e.as_str().to_string()).collect();

  let mangaTitle: Vec<String> =  Regex::new(r#"[^/]*$"#).unwrap().find_iter(&_mangaTitle).map(|e| e.as_str().to_string().replace("-", " ")).collect();
  let mangaChapter: Vec<String> = Regex::new(r#"\d+$"#).unwrap().find_iter(&_mangaChapter).map(|e| e.as_str().to_string()).collect();

  // print!("mangaTitle {}\n", mangaTitle[0]);
  // print!("mangaChapter {}\n", mangaChapter[0]);

  let directory = format!("downloads/{}/Ch.{}/", mangaTitle[0], mangaChapter[0]);
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
    // let baseUrl = format!("https://uploads.mangadex.org/data/{}/{}", hash, chapterImagesFileName[j]);
    let url = reqwest::Url::parse(&image).unwrap();
    let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();
    i += 1;
  }
}

pub fn tcbscans() {
  let _mangaList = getMangaList();

  // let mangaInfo = getManga(_mangaId);
  // let mangaChapters = getMangaChapters(&mangaInfo);
  // // print!("mangaChapters length {}", mangaChapters.data.len());
  // let mangaTitle = mangaInfo.data.attributes.title.en;
  // // getMangaChapterImages(mangaTitle.to_string(), mangaChapters);
  //
  // let mut arr: Vec<_> = Vec::<_>::new();
  // print!("available chapters:\n");
  // for chapter in mangaChapters.data.iter() {
  //   // print!("{}\n", chapter.attributes.chapter);
  //   if chapter.attributes.chapter.as_ref().is_some() {
  //     arr.push(chapter.attributes.chapter.as_ref().expect("expect title not to be null").to_string());
  //   }
  // }
  // // arr.sort();
  // arr.sort_by(|a, b| a.partial_cmp(b).unwrap()); // this is fucking ridiculous, Rust. just sort the goddamn float vector by default.
  // for i in arr.iter() {
  //   print!("{}\n", i);
  // }
  //
  // print!("\nEnter the chapters you want to download\n");
  // print!("Options: 'all', 'asf (all chapters in a single folder)', 'chapter numbers separated by spaces', 'oneshot', 'quit'\n");

  // loop {
    print!("-> ");
    std::io::stdout().flush().expect("failed to flush stdout");

    let mut userInput = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut userInput).expect("Could not read line");

    // match userInput.trim() {
    //   "all" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), false),
    //   "asf" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "".to_string(), true),
    //   "oneshot" => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, "oneshot".to_string(), true),
    //   "quit" => break,
    //   _ => getMangaChapterImages(mangaTitle.to_string(), &mangaChapters, userInput, false), // println!("userInput {}", userInput),
    // }
    // println!("\nDownload completed!\n");
    // break;
  // }
  let index: usize = userInput.trim().parse().unwrap();
  let _mangaTitle = _mangaList[index].clone();
  let _chapterList = getChapterList(_mangaList[index].clone());
  // getChapterList(_mangaList[index].clone());

  // print!("\x1B[2J\x1B[1;1H"); // clears terminal
  for chapter in _chapterList {
    // print!("{}\n", chapter);
    getChapterImages(_mangaTitle.clone(), chapter);
    // getChapterImages(_mangaTitle, _chapterList[0].clone());
  }
}
