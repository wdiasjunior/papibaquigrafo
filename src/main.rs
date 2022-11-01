#![allow(non_snake_case)]

mod mangadex;

fn main() {
  let mangaInfo = mangadex::getManga();
  let mangaChapters = mangadex::getMangaChapters(&mangaInfo);
  let mangaTitle = mangaInfo.data.attributes.title.en;
  mangadex::getMangaChapterImages(mangaTitle.to_string(), mangaChapters);

}
