fn getChapterImages() {
  // print!("\x1B[2J\x1B[1;1H"); // clears terminal
  // print!("{} \n\n", _mangaChapter);

  // gambiarra to download some manga from comico.jp
  // requires manual link and chapter change in order to work
  // TODO - python script to glue the images together

  let mangaTitle = "僕らは彼女に依存症";
  let mangaChapter = "41";

  let mut j = 1;
  loop {
    let imgUrl = if j < 10 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/40/9d5afa877d7e3df233ca5196ee00df44_00{}_00{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvNDAvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTIzNTg5Nn19fV19&Signature=JW6fmmjdlx~FA1ZwhbS~cGQIHzr8g3~oTgUZMTDtp5u8fBBuTOV3gDhUE~cMt8IyfWHLsiE59Kua7sJcIUyt7kl1BMPjTEvTY8vjEztrhoizA-Powu8ZGTdwO3U29gqzYAAZlYQm0RTyqc~x9GdK-nToW995Co62Dl-HApj3K0chF0cM8K5KCjvg2QsAgeFrl7GKmFHnJd15E9FRnrzo1emzm9E6Rw1aBMELNPdxei0q-~aGAJJxMMEwttrNqnd~Gt-Ve38SY5msRe1URrbjfhzxyCgOz0GZSsAM-VFCeEoAzLFBop1e~JbyumwA39cOiuDtKEBjWz0cKAlON6EAyA__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 1, j)
    } else if j > 20 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/40/9d5afa877d7e3df233ca5196ee00df44_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvNDAvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTIzNTg5Nn19fV19&Signature=JW6fmmjdlx~FA1ZwhbS~cGQIHzr8g3~oTgUZMTDtp5u8fBBuTOV3gDhUE~cMt8IyfWHLsiE59Kua7sJcIUyt7kl1BMPjTEvTY8vjEztrhoizA-Powu8ZGTdwO3U29gqzYAAZlYQm0RTyqc~x9GdK-nToW995Co62Dl-HApj3K0chF0cM8K5KCjvg2QsAgeFrl7GKmFHnJd15E9FRnrzo1emzm9E6Rw1aBMELNPdxei0q-~aGAJJxMMEwttrNqnd~Gt-Ve38SY5msRe1URrbjfhzxyCgOz0GZSsAM-VFCeEoAzLFBop1e~JbyumwA39cOiuDtKEBjWz0cKAlON6EAyA__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 2, j)
    } else if j > 40 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/40/9d5afa877d7e3df233ca5196ee00df44_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvNDAvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTIzNTg5Nn19fV19&Signature=JW6fmmjdlx~FA1ZwhbS~cGQIHzr8g3~oTgUZMTDtp5u8fBBuTOV3gDhUE~cMt8IyfWHLsiE59Kua7sJcIUyt7kl1BMPjTEvTY8vjEztrhoizA-Powu8ZGTdwO3U29gqzYAAZlYQm0RTyqc~x9GdK-nToW995Co62Dl-HApj3K0chF0cM8K5KCjvg2QsAgeFrl7GKmFHnJd15E9FRnrzo1emzm9E6Rw1aBMELNPdxei0q-~aGAJJxMMEwttrNqnd~Gt-Ve38SY5msRe1URrbjfhzxyCgOz0GZSsAM-VFCeEoAzLFBop1e~JbyumwA39cOiuDtKEBjWz0cKAlON6EAyA__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 3, j)
    } else {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/40/9d5afa877d7e3df233ca5196ee00df44_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvNDAvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTIzNTg5Nn19fV19&Signature=JW6fmmjdlx~FA1ZwhbS~cGQIHzr8g3~oTgUZMTDtp5u8fBBuTOV3gDhUE~cMt8IyfWHLsiE59Kua7sJcIUyt7kl1BMPjTEvTY8vjEztrhoizA-Powu8ZGTdwO3U29gqzYAAZlYQm0RTyqc~x9GdK-nToW995Co62Dl-HApj3K0chF0cM8K5KCjvg2QsAgeFrl7GKmFHnJd15E9FRnrzo1emzm9E6Rw1aBMELNPdxei0q-~aGAJJxMMEwttrNqnd~Gt-Ve38SY5msRe1URrbjfhzxyCgOz0GZSsAM-VFCeEoAzLFBop1e~JbyumwA39cOiuDtKEBjWz0cKAlON6EAyA__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 1, j)
    };

    let directory = format!("downloads/{}/Ch.{}/", mangaTitle, mangaChapter);
    std::fs::create_dir_all(&directory).unwrap();
    let fileExtension = "jpg";
    let fileName = if j < 10 {
      format!("{}/00{}.{}", &directory, j, fileExtension)
    } else {
      format!("{}/0{}.{}", &directory, j, fileExtension)
    };
    let mut file = std::fs::File::create(fileName).unwrap();
    let url = reqwest::Url::parse(&imgUrl).unwrap();
print!("{}\n", j);
print!("{}\n", url);
    let _mangaImage = reqwest::blocking::get(url).unwrap().copy_to(&mut file).unwrap();
print!("\n{}\n", _mangaImage);

    j += 1;
    if j > 60 {
      break;
    }
  }
}

pub fn comicojp() {
  print!("function disabled\n");
  // getChapterImages();
}
