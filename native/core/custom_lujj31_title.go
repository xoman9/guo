package core

import "strings"

// Reject generic detail headings that used to overwrite the real Tao catalog title.
func nativeLujj31CatalogTitleValid(title, sourceID string) bool {
 title = strings.TrimSpace(title)
 if title == "" || title == "短剧" || title == "短劇" ||
  title == "影片 "+sourceID || title == "影片" || title == "桃果" ||
  title == "播放線路" || title == "播放线路" {
  return false
 }
 return true
}
