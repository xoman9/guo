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
| `sorani` | 109 青空次元官方 API 資料結構及既有帳號會話 | `native/sources_legacy109/provider_sorani.go` | 分類／搜尋／詳情／分集／播放授權檢查；支援載入合法舊登入紀錄，不繞過登入權限；模擬 API 測試通過，未做線上驗證 |

- `native/sources_custom` 原有 13 個新增站源維持，搭配恢復的 7 個原生 ID，已在 Go 測試工作區中一起編譯；這是 **13 + 7 = 20 個不重複的註冊 ID**，不是 110 的全部原生站源。
- 已通過的 GitHub Actions：`43518953421`、`43519734728`、`43520677572`、`43522259053`；純離線測試；不代表影片線上播放已成功。
- 110 APK 另外辨識出的原生 ID，尚未逐一恢復：`huangdou`, `yeguo`, `dsd`, `jinguoduanju`, `jinguo`, `beabox`, `cloudfront`, `huangguoai`, `huangguovideo`。其中部分與 custom 頻道用途相近但 ID 不同，整合前必須辨別衝突與重複來源。
- 日後增加站源：在這個分支建立與 110 相容的獨立 Go 模組／Provider（必要時擴增 `native/sources_legacy109`），先更新 `flavor_zhenguojian.go`，CI 使用 `link_sources([sources_custom, sources_missav, sources_legacy109])` 編譯；新增單元測試再核對真實網頁／內容／播放。
- **不能用重新註冊的虛構站源名稱或只有假資料的 Provider 來宣稱已恢復。** 只有真正的解析器與必要的真實網站驗證才能標記為可用。

## 本次追加：紅果 109 網頁版相容界線

- 只重建 109 已知的 `/category/<kind>?page=N`、`/detail?series_id=...`、`/player/<series>/<episode>` 路線，不臆測原版 App 內 CENC、授權憑證與付費串流機制。
- 對分集編號與播放頁的 `vid`/`series_id` 做雙重比對，錯誤集數不得改播預覽片；含 DRM/CENC 標誌或未提供公開 URL 時回報不支援。
- `native/sources_legacy109/provider_hongguo_test.go` 使用本地頁面夾具測試分類、頁碼、劇名、封面、分集配對、錯集拒絕及加密授權拒絕。
- 最新成功 CI：`43520677572`。**尚未確認當前網站資料格式、線上影片能否播放、實機下載是否正常**。

## 本次追加：青空次元 Sorani 109 公開 API 與登入權限

- 源自 109 的 `/video`、`/video/search`、`/video/<id>`、`/video/episode/<id>/play`，以 110 `Host.Fetch` 執行，使用正規的 Provider Registry 註冊。
- 目錄、分頁、分類、分集、劇集／集數 ID 均做限制；播放前會檢查指定劇集與集數是否吻合。
- 播放僅使用原有 `sorani-session.json` 的合法 `accessToken`（透過 `Host.LoadState`）；不製造臨時登入、破解驗證或繞過 `canPlay=false` 的權限拒絕。缺少權限即回報登入或無權播放。
- 本次未移植原版完整的登入 UI、驗證碼及更新 Token 介面；已有會話可能過期，必須另外驗證實際登入、播放和更新機制。
- `native/sources_legacy109/provider_sorani_test.go` 使用模擬站源測試公開目錄、搜尋、詳細資料、分集與 `canPlay` 授權／跨劇集 ID 驗證。GitHub Actions `43522259053` 成功，但**沒有實際在線播放測試**。

## 2026-11-14 續修紀錄：29 ID 重建狀態及驗收分級

- 目前分支已新增剩餘 9 個 ID 的實驗性 Provider：`huangdou`、`yeguo`、`dsd`、`jinguoduanju`、`jinguo`、`beabox`、`cloudfront`、`huangguoai`、`huangguo-video`。加上原有 7 個與 custom 13 個，共有 **29 個靜態註冊 ID**。
- 上方「13 + 7 = 20」表格是當時的歷史階段，**不是當前分支註冊數量**。實際清單以 `scripts/check_reconstructed_sources.py` 和 `native/sources_legacy109/flavor_zhenguojian.go` 為準。
- 驗收必須分開標記：① registry 無重複及靜態檢查，② Go 編譯及夾具測試，③ 目前站點的真實目錄／詳情／播放解析，④ Android ARM64/v7a 實機測試，⑤ 固定憑證簽章驗證。未經 ③④ 的站源不得標記為可用。
- `native/sources_missav/provider.go` 目前只辨識公開 HTML 中直接呈現的媒體 URL。動態腳本／封裝播放器未完整重建，播放能力未驗證；不應宣稱已等價恢復 110 原生 MissAV。
- `.github/workflows/build-recovered-109-test-apks.yml` 現在會在 Provider、Bridge、Core、Flutter、Android 或建置腳本異動時觸發；其輸出仍明確為 **EXPERIMENTAL**，不得當成原廠 110 完整可用版 Release。
- CI 組裝與簽章必須使用既有 Android keystore secrets，並核對原憑證 SHA-256；絕不以臨時 keystore 代替永久簽章。
- 待驗項目：逐站登入與權限、分類與搜尋、封面及片名配對、詳情及分集、播放和返回、下載及合併、異常網路；尤其紅果 CENC/App、青空登入刷新與 MissAV 動態媒體仍有明確缺口。

## 2026-11-14 續修：MissAV 公開網址解析與劇集一致性

- 修正 `native/sources_missav/provider.go` 的公開 MP4／M3U8 URL 規則，先將 HTML/JSON 中的 `\\/` 還原成 `/` 再辨識 HTTPS/HTTP 直連。
- 新增 `mediaPageForTask`，在任何頁面請求之前先核對來源 ID、劇集 ID、分集 ID，以及實際頁面路徑，避免跨影片頁面混用。
- 新增 `directMediaURL` 並只接納符合核心媒體 URL 判斷的直接網址；未公開直連的動態播放器維持明確錯誤，不宣稱可播放。
- `provider_test.go` 補測一般 MP4、跳脫斜線 MP4、M3U8、錯誤來源／劇集／播放頁以及沒有直連網址的情況。
- 相關提交：`8e8bb382`、`d1e2af4b`。已用獨立 Go 正規表示式小程式核對直連匹配，但目前尚未取得 GitHub Actions 此兩筆提交的完整編譯及 APK 結果，因此此項仍屬 **待 CI／實機驗證**。

## 2026-11-14 本輪修復：MissAV 編譯阻礙與播放器來源檢查

- cfe2f65：修復 provider.go 開頭損壞的 Go 宣告，恢復 mediaSlug 最長 128 字元檢查。
- a05aa12：只接受 video/source 標籤或 JSON 播放欄位直接揭露的單一媒體網址；不再從整頁任意選取第一個 MP4/M3U8。
- 6bcd172：補上廣告連結、多個相異播放網址的拒絕測試。
- 349f8da、8c00ebb：統一目錄卡片、詳情與播放可接受的 1–128 字元、大小寫混合影片 ID，加入回歸測試。
- 已以隔離 Go 標準函式庫小程式核對相關 regexp 可編譯並能辨識 HTML video、JSON 跳脫斜線及短影片 ID。尚未取得整個 Flutter/Go 專案的 CI 成功證據，也未完成實機／線上播放驗證。
- 原生 16 加額外 13，共 29 個 Provider ID 的註冊數量未變；110 動態 MissAV 播放器仍未完整恢復。

## 2026-11-14 續修：分頁與跨集播放防護

- MissAV：`007eb18` 與 `93288fa` 將下一頁判斷改為明確的 `rel=next` 或 `aria-label=next`，不再把普通 `page=N` 當成下一頁。補上離線回歸測試；網站若只提供頁碼而沒有下一頁標記，將保守地停止自動翻頁，後續需對照真實網站調整。
- 黃果視頻：`1d3f841` 與 `c8841bf` 新增 `hg109VideoPage`；單片要求 ID 與播放頁相符，系列分集要求章節 ID 與 `/video/<episode>` 路徑相符，拒絕跨集播放網址與錯誤站點。新增本地測試並更新原本期待錯誤頁進入 Fetch 的斷言。
- 以上均為程式碼與測試檔提交，尚未取得 GitHub Actions 的正式 Go 測試／APK 建置通過結果；勿標記為線上播放驗收完成。

## 2026-11-14 續修：29 Provider 全模組 CI 覆蓋

- `5747fbc`：離線測試工作流程現在同時執行 16 個重建原生 ID 所屬的 legacy109/missav 模組，以及 13 個額外 ID 所屬的 sources_custom 模組；避免僅測原生重建模組就宣稱全部 29 個通過。
- `5038a87`：實驗性永久簽章 APK 工作流程亦先執行以上三個模組的 Go 測試，且不使用 Go 測試快取（`-count=1`），未通過則不得進入固定簽章建置。
- `eb5cf02`：一般離線測試工作流程亦使用 `-count=1`；Go 格式檢查以非改寫模式進行，不再悄悄修改 checkout 原始碼。
- 這些提交修改的是 CI 驗證覆蓋與工作流程；目前尚未取得新提交對應的 Actions 成功紀錄、實際播放測試或經憑證驗證的 APK，因此其狀態仍為待驗。

## 2026-11-14 續修：黃果視頻目錄網址驗證與帶參數分集

- `fce63c6`：黃果視頻目錄卡片新增 `hg109CardID`；只接受 `huangguo.video` 同站 HTTPS `/series/{id}` 或 `/video/{id}`，拒絕外站偽裝或多層非標準路徑。
- `ed354ed`：補測外站卡片、相似網域、巢狀路徑、編碼路徑及含排序查詢參數的正常卡片。
- `880555c`：`hg109VideoPage` 對分集網址改以 HTTPS 主機與標準路徑檢查，不再完整字串比對，允許同一集附帶既有查詢參數並保留送給站源；禁止跨站與 fragment。
- `e5952a7`：新增帶 token 查詢參數、主機偽裝、錯誤協定及 fragment 的離線回歸測試。
- 所有修改保留在 `feat/recover-109-missav-20261114`，尚需本輪 Actions 成功紀錄與網站實測才能驗收。

## 2026-11-14 續修：韓小圈／鬼片分集及封面配對

- `ec2f72d`：兩個 Provider 在正式取得播放器 HTML 前，從目前作品 Detail 重新讀取分集清單，核對章節 ID 及章節 URL 是否同屬目前作品；拒絕跨作品混播。此機制新增一次 Detail 請求，若站源當下無法回傳詳情，將保守拒絕播放。
- `3c7db4c`：新增韓小圈及鬼片的有效分集、錯誤播放頁、跨作品分集和錯誤 Provider 離線回歸測試。
- `fb85ad2`：卡片擷取封面只查本張卡片所在區段，不再跨入下一張卡片導致封面／片名錯配。
- `8f21d5d`：對韓小圈及鬼片分別新增相鄰卡片只有第二張具封面的測試，要求第一張不能借用第二張封面。
- 上述為程式碼與回歸測試提交，尚未取得對應 GitHub Actions 完整測試成功及實際網站播放證據。

## 2026-11-14 續修：黑桃站源跨集防護與選線一致性

- `3b935d5`：新增 heitaoPlayerRoute，播放前核對 heitao Provider、作品 ID、播放線路 sid、集數 nid 與完整原生 player 路徑；不相符時不發送頁面請求。
- `b6cf1c8`：新增有效播放、跨作品、跨集、偽裝網域、非法播放參數以及驗證失敗不應觸發網路請求等測試。
- `63d2886`：修正同集數多線路原本因 map 迭代順序而隨機選線，改成集數最多優先，相同集數時以較小的 sid 優先。
- `3f7febe`：新增重複 200 次的同集數選線回歸測試。
- `90394d3`：黑桃目錄卡片僅採用與原站同域、HTTPS、標準 /index.php/vod/detail/id/{id}.html 路徑，拒絕其他網站偽裝。
- `17c509d`：新增跨網域、相似網域、嵌套路徑、編碼路徑及正常查詢參數等卡片測試。
- 此次是對原生黑桃兼容 Provider 的程式碼與測試修正；尚未確認本輪 GitHub Actions 綠燈、線上播放或固定簽章 APK。

## 2026-11-14 續修：青空次元跨作品播放防護

- `210ca7d`：青空次元在查詢分集播放 URL 及送出已保存的帳號授權之前，先透過目前作品 Detail 核對章節 ID 是否存在，避免其他作品的分集 ID 被錯誤播放。
- `462ded9`：補上跨作品分集不能觸發播放網路請求的測試，並更新既有授權允許／拒絕測試的 Detail 模擬資料。
- 驗證會多呼叫一次公開 Detail API；若作品詳情當時不可取得則保守地停止播放，不會以無法核實的分集繼續請求。
- 本次程式碼已提交，但此記錄不表示 GitHub Actions 通過、網路影音可播放或固定簽章 APK 已完成。
