package src

import (
  "fmt"
  beeep "github.com/gen2brain/beeep"
)

// Every connector's top level function returns one of these so the notification
// is sent from a single place instead of each connector rolling its own.
type DownloadStatus int

const (
  DownloadSuccess DownloadStatus = iota
  DownloadFailed
  DownloadCancelled
)

type DownloadResult struct {
  Status     DownloadStatus
  MangaTitle string
  Chapters   int    // chapters processed - 0 when the connector cannot tell
  Reason     string // only set for DownloadFailed
}

func downloadSuccess(_mangaTitle string, _chapters int) DownloadResult {
  return DownloadResult{
    Status:     DownloadSuccess,
    MangaTitle: _mangaTitle,
    Chapters:   _chapters,
  }
}

func downloadFailed(_mangaTitle string, _reason string) DownloadResult {
  return DownloadResult{
    Status:     DownloadFailed,
    MangaTitle: _mangaTitle,
    Reason:     _reason,
  }
}

func downloadCancelled() DownloadResult {
  return DownloadResult{Status: DownloadCancelled}
}

func notifyDownloadResult(_source string, _result DownloadResult) {
  // Quitting out of a connector is not something worth notifying about
  if _result.Status == DownloadCancelled {
    return
  }

  title, message := buildNotificationMessage(_source, _result)

  err := beeep.Notify(title, message, "")
  if err != nil {
    fmt.Println("Error sending notification.")
  }
}

func buildNotificationMessage(_source string, _result DownloadResult) (string, string) {
  mangaTitle := _result.MangaTitle
  if mangaTitle == "" {
    mangaTitle = "Manga"
  }

  if _result.Status == DownloadSuccess {
    if _result.Chapters > 0 {
      return "Download Finished", fmt.Sprintf("%s - %d %s downloaded from %s.", mangaTitle, _result.Chapters, pluralizeChapters(_result.Chapters), _source)
    }
    return "Download Finished", fmt.Sprintf("%s has finished downloading from %s.", mangaTitle, _source)
  }

  if _result.Reason != "" {
    return "Download Failed", fmt.Sprintf("%s (%s): %s", mangaTitle, _source, _result.Reason)
  }
  return "Download Failed", fmt.Sprintf("Could not download %s from %s.", mangaTitle, _source)
}

func pluralizeChapters(_chapters int) string {
  if _chapters == 1 {
    return "chapter"
  }
  return "chapters"
}
