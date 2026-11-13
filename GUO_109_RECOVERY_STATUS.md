# 原生站源恢復進度（109 → 110）／2026-11-14

## 來源與原則

- 110 Core 仍為 `Xanerix/guoapp` 發佈的 1.0.10+110 架構；使用其正式的 `core.Host`、`core.Provider`、`core.Register`，不覆蓋 110 的 core。
- 109 可讀 Provider 來源：`xoman9/guo2` 的 `custom` 分支中 `native/core/provider_*.go`。**不可**將原 109 全部檔案直接放進 110 `native/core`，因同名常數、內部函式及舊 `Downloader` 欄位會衝突。
- MissAV 來源：使用者提供 110 APK 中 `lib/arm64-v8a/libduanju_core.so` 的 Go 編譯資訊、符號名稱及可見 URL/路由資訊。這不是 APK 完整反編譯的原始碼；動態 HTML 與播放處理仍待核實。
- 正式 `main` 目前仍維持 110＋13 額外站源及介面／下載／合併修改；本恢復分支為**實驗性重建**，未經實機站源驗證不得發布成「全部原生頻道」 APK。

## 已重建並通過離線 Go 編譯／測試

| Provider ID | 來源 | 模組檔案 | 驗證界線 |
|---|---|---|---|
| `hanxiaoquan` | 109 舊解析器 | `native/sources_legacy109/provider_web.go` | 分類、目錄卡片及播放路徑介面；未做實際線上測試 |
| `guipian` | 109 舊解析器 | 同上 | 分類、目錄、RSS 搜尋與播放路徑介面；未做實際線上測試 |
| `heitao` | 109 舊解析器 | `native/sources_legacy109/provider_heitao.go` | MacCMS 目錄／詳情／分集／player_aaaa 巢狀 JSON；未做實際線上測試 |
| `missav` | 110 ARM64 APK 編譯資訊／字串與符號 | `native/sources_missav/provider.go` | 110 路由與基本 HTML 解析骨架；並非還原原本 JS 解包或完成播放支援 |
| `huangju` | 109 API 程式及其訪客驗證流程 | `native/sources_legacy109/provider_huangju.go` | 分類／目錄／詳情／分集與 CloudFront Cookie 播放解析相容版；模擬 API 測試通過，未做線上驗證 |
| `hongguo` | 109 網頁 `_ROUTER_DATA` 備援解析 | `native/sources_legacy109/provider_hongguo.go` | 分類／目錄／詳情／分集／公開播放 URL；不包含 109 App/CENC 授權路徑，拒絕錯集及加密影片，尚未實機驗證 |

- `native/sources_custom` 原有 13 個新增站源維持，搭配恢復的 6 個原生 ID，已在 Go 測試工作區中一起編譯；這是 **13 + 6 = 19 個不重複的註冊 ID**，不是 110 的全部原生站源。
- 已通過的 GitHub Actions：`43518953421`、`43519734728`、`43520677572`；純離線測試；不代表影片線上播放已成功。
- 110 APK 另外辨識出的原生 ID，尚未逐一恢復：`sorani`, `huangdou`, `yeguo`, `dsd`, `jinguoduanju`, `jinguo`, `beabox`, `cloudfront`, `huangguoai`, `huangguovideo`。其中部分與 custom 頻道用途相近但 ID 不同，整合前必須辨別衝突與重複來源。
- 日後增加站源：在這個分支建立與 110 相容的獨立 Go 模組／Provider（必要時擴增 `native/sources_legacy109`），先更新 `flavor_zhenguojian.go`，CI 使用 `link_sources([sources_custom, sources_missav, sources_legacy109])` 編譯；新增單元測試再核對真實網頁／內容／播放。
- **不能用重新註冊的虛構站源名稱或只有假資料的 Provider 來宣稱已恢復。** 只有真正的解析器與必要的真實網站驗證才能標記為可用。

## 本次追加：紅果 109 網頁版相容界線

- 只重建 109 已知的 `/category/<kind>?page=N`、`/detail?series_id=...`、`/player/<series>/<episode>` 路線，不臆測原版 App 內 CENC、授權憑證與付費串流機制。
- 對分集編號與播放頁的 `vid`/`series_id` 做雙重比對，錯誤集數不得改播預覽片；含 DRM/CENC 標誌或未提供公開 URL 時回報不支援。
- `native/sources_legacy109/provider_hongguo_test.go` 使用本地頁面夾具測試分類、頁碼、劇名、封面、分集配對、錯集拒絕及加密授權拒絕。
- 最新成功 CI：`43520677572`。**尚未確認當前網站資料格式、線上影片能否播放、實機下載是否正常**。
