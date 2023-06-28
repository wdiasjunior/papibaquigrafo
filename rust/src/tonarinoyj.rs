// https://github.com/Lighter207/OnePunchMan
// https://github.com/Ay-De/OPM_tonarinoyj_webscrapper

// https://tonarinoyj.jp/episode/3270296674379223025 - Erio and the Electric Doll
// https://tonarinoyj.jp/episode/4856001361072081927 - One Punch Man

pub fn tonarinoyj() {
  print!("tonarinoyj {}", "test");
}



// python code
//
// import json
// from selenium import webdriver
// from selenium.webdriver.common.by import By
// from selenium.common.exceptions import TimeoutException, StaleElementReferenceException
// from selenium.webdriver.support import expected_conditions as EC
// from selenium.webdriver.support.ui import WebDriverWait
// from selenium.webdriver.edge.options import Options
// import re
// from multiprocessing import Pool
// from tqdm import tqdm
//
// from modules.webdriver import setup_webdriver
// from modules.helpers import download
//
//
// class TonariScrapper:
//
//     #Use a custom enter and exit function to ensure the chromium process is killed
//     #even if the programm crashes during runtime
//     def __enter__(self):
//         return self
//
//     def __init__(self):
//
//         self.manga_url = 'https://tonarinoyj.jp/episode/4855956445072905450'
//         self.download_location = './downloads/'
//
//         self._chapter_links = {}
//         self._chapter_page_links = {}
//
//         #Get the Website base URL
//         self._tonarinoyj_url = self.manga_url.rsplit('/', 1)[0] + '/'
//
//         self._options = Options()
//         self._options.add_argument('headless')
//         self._options.add_argument('window-size=1920x1080')
//         self._options.add_argument('disable-extensions')
//         self.webdriver = webdriver.Edge(options=self._options)
//
//         self._get_chapters()
//
//         self._chapter_selection()
//
//         self._chapter_download()
//
//         print('Download complete.')
//
//     def _chapter_selection(self):
//         """
//         This function will check which chapters of the given manga are available
//         """
//         _chapters_nums = list(self._chapter_links.keys())
//         _chapters_nums.sort(reverse=False)
//
//         print(f'********\nFound {len(_chapters_nums)} '
//               f'chapters. Select chapters for download.\nOptions:')
//
//         while True:
//             _download_selection = input('Chapter number, all or latest\n')
//
//             if _download_selection.lower() == 'latest':
//                 print('Downloading chapter {}.'.format(_chapters_nums[-1]))
//                 self._get_chapter_page_links(_chapters_nums[-1])
//                 break
//
//             elif _download_selection.lower() == 'all':
//                 print('All {} Chapters will be downloaded.'.format(
//                     len(list(self._chapter_links.keys()))))
//
//                 for c in tqdm(_chapters_nums, unit='iB', unit_scale=True):
//                     self._get_chapter_page_links(c)
//                 break
//
//             elif _download_selection.isdigit():
//                 self._get_chapter_page_links(int(_download_selection))
//                 break
//
//             else:
//                 print('Invalid input. Example input (without ""): "42", "all" or "latest"')
//
//     def _get_chapters(self):
//         """
//         This function will scrap the available chapters from the tonarinoyj website and
//         split found chapters into available and unavailable manga chapters
//         """
//         self.webdriver.get(self.manga_url)
//         wait = WebDriverWait(self.webdriver, 10)
//
//         _chap_list_cont = self.webdriver.find_element(By.XPATH,
//                                                       '//button[@class="js-read-more-button"]')
//
//         self.webdriver.execute_script('return arguments[0].scrollIntoView();',
//                                       _chap_list_cont)
//
//         _chapter_container_loaded = wait.until(EC.visibility_of_element_located((By.XPATH,
//                                                          '//span[@class="loading-text"]')))
//         wait.until(lambda x: 'hidden' in _chapter_container_loaded.get_attribute('class'))
//
//         while True:
//
//             try:
//                 wait.until(EC.visibility_of_element_located(
//                             (By.XPATH, '//button[@class="js-read-more-button"]'))).click()
//
//             except (TimeoutException, StaleElementReferenceException):
//                 break
//
//         _chapters_available = self.webdriver.find_elements(By.XPATH,
//                 '//*[@class=" episode" and not(contains(@class, "private episode"))] | '
//                 '//*[@class="episode current-readable-product"]')
//
//         _chapters_private = self.webdriver.find_elements(By.XPATH,
//                 '//*[@class="private episode"]')
//
//         regx = re.compile(r".*?\[[^\d]*(\d+)[^\d]*\].*")
//
//         _chapters_private = [regx.findall(x.text)[0] for x in _chapters_private]
//
//         for c in _chapters_available:
//             c_num = regx.findall(c.text)
//             c_url = self._tonarinoyj_url + c.get_attribute('data-id')
//
//             if c_num and c_num not in _chapters_private:
//                 self._chapter_links.update({int(c_num[0]): c_url})
//
//         if len(_chapters_private) > 0:
//             print(f'Note: Chapters {_chapters_private} are private and not available.')
//
//
//     def _get_chapter_page_links(self, chapter_num):
//         """
//         This function scraps the direct links to the manga pages for
//         the user selected manga chapters.
//         """
//         self.webdriver.get('view-source:' + self._chapter_links[chapter_num] + '.json')
//         content = self.webdriver.page_source
//         chapter_content = self.webdriver.find_element(By.CLASS_NAME,
//                                                            'line-content').text
//         chapter_data = json.loads(chapter_content)
//
//         chapter_pages = chapter_data['readableProduct']['pageStructure']['pages']
//
//         _page_num = 0
//         for page in chapter_pages:
//             if 'src' in page:
//                 _page_num = _page_num + 1
//                 _page_link = page['src']
//
//                 self._chapter_page_links.update({(chapter_num, _page_num): _page_link})
//
//     def _chapter_download(self):
//         """
//         This function is downloading the manga pages. Manga pages are stored in their
//         corresponding chapter folders. The download is happening in parallel with 3 parallel
//         downloads.
//         """
//         input_args = [(p_url,
//                        self.download_location + str(c_num) + '\\',
//                        str(p_num) + '.jpeg')
//                       for (c_num, p_num), p_url in self._chapter_page_links.items()]
//
//         with Pool(3) as pool:
//             pool.starmap(download,
//                          tqdm(input_args, total=len(input_args),
//                               unit='iB', unit_scale=True))
//
//
//     def __exit__(self, exc_type, exc_val, exc_tb):
//         self.webdriver.quit()
//
//
// def main():
//
//     setup_webdriver()
//
//     with TonariScrapper() as t:
//         print(t)
//
// if __name__ == '__main__':
//     main()
