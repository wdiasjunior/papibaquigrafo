fn getChapterImages() {
  // print!("\x1B[2J\x1B[1;1H"); // clears terminal
  // print!("{} \n\n", _mangaChapter);

  // gambiarra to download some manga from comico.jp
  // requires manual link and chapter change in order to work
  // TODO - python script to glue the images together

  let mangaTitle = "僕らは彼女に依存症";
  let mangaChapter = "14";

  let mut j = 1;
  loop {
    let imgUrl = if j < 10 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/13/d100ba3bdb92c9f24e36e5506195856b_00{}_00{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvMTMvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTA2MDE3M319fV19&Signature=RmVPh1cOdcmUSOvQARNuLZ-Buf1DM4ABLZoFUFJKbvcUdCwO3~mB9MKaWTJzzt46p070VR~EZn5fNeCP025gPoz6Tmf~tEfbyE6cB5UYTIpKACKxgSu2Uhk2pLZKmV6ldXBLPC1sDr~mmJEyIqIOxbXAEcfcvKhAWMMnjx7d8BUZG-vGVFYCbrpIsQFP69G17ForQnYGlE921BEo7YQo5Q4XwEVi3Pog1juBuYKYXf5Sfu2SmOGuhTy9GtO4i3sZobI2DA8MvZ3ioqJI91H1fikm8VQU6Gg4GVnWa~PlXUWA8XHrBEI2~BJi7YMPzJwAW5YU3LIMIIeonVtjaI--Ew__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 1, j)
    } else if j > 20 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/13/d100ba3bdb92c9f24e36e5506195856b_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvMTMvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTA2MDE3M319fV19&Signature=RmVPh1cOdcmUSOvQARNuLZ-Buf1DM4ABLZoFUFJKbvcUdCwO3~mB9MKaWTJzzt46p070VR~EZn5fNeCP025gPoz6Tmf~tEfbyE6cB5UYTIpKACKxgSu2Uhk2pLZKmV6ldXBLPC1sDr~mmJEyIqIOxbXAEcfcvKhAWMMnjx7d8BUZG-vGVFYCbrpIsQFP69G17ForQnYGlE921BEo7YQo5Q4XwEVi3Pog1juBuYKYXf5Sfu2SmOGuhTy9GtO4i3sZobI2DA8MvZ3ioqJI91H1fikm8VQU6Gg4GVnWa~PlXUWA8XHrBEI2~BJi7YMPzJwAW5YU3LIMIIeonVtjaI--Ew__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 2, j)
    } else if j > 40 {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/13/d100ba3bdb92c9f24e36e5506195856b_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvMTMvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTA2MDE3M319fV19&Signature=RmVPh1cOdcmUSOvQARNuLZ-Buf1DM4ABLZoFUFJKbvcUdCwO3~mB9MKaWTJzzt46p070VR~EZn5fNeCP025gPoz6Tmf~tEfbyE6cB5UYTIpKACKxgSu2Uhk2pLZKmV6ldXBLPC1sDr~mmJEyIqIOxbXAEcfcvKhAWMMnjx7d8BUZG-vGVFYCbrpIsQFP69G17ForQnYGlE921BEo7YQo5Q4XwEVi3Pog1juBuYKYXf5Sfu2SmOGuhTy9GtO4i3sZobI2DA8MvZ3ioqJI91H1fikm8VQU6Gg4GVnWa~PlXUWA8XHrBEI2~BJi7YMPzJwAW5YU3LIMIIeonVtjaI--Ew__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 3, j)
    } else {
      format!("https://images.comico.io/comic/chapter/image/ja/onetimecontents/app/31134/13/d100ba3bdb92c9f24e36e5506195856b_00{}_0{}.jpg/dims/crop/x2000+0+0/optimize?Policy=eyJTdGF0ZW1lbnQiOiBbeyJSZXNvdXJjZSI6Imh0dHBzOi8vaW1hZ2VzLmNvbWljby5pby9jb21pYy9jaGFwdGVyL2ltYWdlL2phL29uZXRpbWVjb250ZW50cy9hcHAvMzExMzQvMTMvKiIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTY4MTA2MDE3M319fV19&Signature=RmVPh1cOdcmUSOvQARNuLZ-Buf1DM4ABLZoFUFJKbvcUdCwO3~mB9MKaWTJzzt46p070VR~EZn5fNeCP025gPoz6Tmf~tEfbyE6cB5UYTIpKACKxgSu2Uhk2pLZKmV6ldXBLPC1sDr~mmJEyIqIOxbXAEcfcvKhAWMMnjx7d8BUZG-vGVFYCbrpIsQFP69G17ForQnYGlE921BEo7YQo5Q4XwEVi3Pog1juBuYKYXf5Sfu2SmOGuhTy9GtO4i3sZobI2DA8MvZ3ioqJI91H1fikm8VQU6Gg4GVnWa~PlXUWA8XHrBEI2~BJi7YMPzJwAW5YU3LIMIIeonVtjaI--Ew__&Key-Pair-Id=APKAIZOQXEUERT6TH4NQ", 1, j)
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
  getChapterImages();
}
