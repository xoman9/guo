# guo 升級與客製修改紀錄（2026-11-14）

## 唯一版本基準：上游最新穩定 Release

- **每一次升級都從 `Xanerix/guoapp` 的最新穩定 Release 開始**，使用 GitHub API `GET /repos/Xanerix/guoapp/releases/latest` 得到 `tag_name`，再核對該 tag 對應的 **commit SHA**。不得把 110、109 或 `upstream/main` 固定寫成下一次升級的基底。
- **2026-11-14 此次實際基底**為 `app-v1.0.10-110`（commit `71f706d0e27147ee076f56620ec2b461821f40b7`），這只是一筆歷史紀錄，**不是**下次的預設版號。
- 複製新 Release 的原始檔案，**保留其全部原有站源及功能**；不要從舊 `guo2/custom` 回填或重建原版站源。
- 由助理直接透過 GitHub 修改，無須使用者執行 BAT。先於工作分支測試，再整合進 `guo/main`。衝突時列出檔名、舊功能與新版差異，不能默默跳過或刪除。

## 客製化重放清單（每次升級）

1. **新增站源**：只加入先前 `guo2` 額外新增的真實 Provider 程式（不是僅加頻道名稱），維持原版站源與 Registry 完整。新增 13 個 ID：`lujj31`、`liuguo`、`xiaoguo`、`youguo`、`meiguo`、`yingguo`、`luguo`、`liguo`、`juguo`、`ningguo`、`mangguo`、`chengguodj`、`jiaoguo`。
   - 此次來源為 `xoman9/guo2` 已驗證的 `4aa58c20020f13efcd3d2877a78d39bab2adb52c` 實作。以 `native/core/custom_*.go`、`native/core/provider_apk391_*.go`、`native/core/provider_{liuguo,lujj31,xiaoguo}*.go` 與 `native/sources_custom/` 為來源；`scripts/app_build.py` 預設連接實際自訂 Provider 模組（不是示例來源）。
   - 名稱原有 `+` 後綴與 `DefaultOff` 屬性都要保留；不預先抓取劇庫。
2. **頻道介面**：移除站源切換的地球按鈕。手機／平板／桌面直接點左側頻道名稱，Android TV 點頻道名稱切換；維持新版其他 UI（例如新增的封面模糊按鈕）。
   - 相關檔案：`lib/ui/widgets/page_header.dart`、`lib/ui/widgets/source_bar.dart`。
3. **隱藏頻道開關**：移植 `guo2` 原本的「隱藏頻道-開／隱藏頻道-關」群組按鈕，對 `SourceSite.restrictedValues` 中預設關閉的站源操作；保留每個站源自己的開關與原本已儲存的設定，關閉下載來源時仍需暫停該站下載。
   - 相關檔案：`lib/ui/settings/sources_view.dart`。
4. **下載位置與檔案存取授權**：設定頁顯示「當前下載目錄」，增加 Android「存取授權」入口，使用系統資料夾選擇器。改變目錄前確認權限，遷移失敗不得損壞原始下載檔。
   - 相關檔案：`lib/ui/settings/storage_view.dart`、`android/app/src/main/kotlin/com/guoapp/MainActivity.kt`、`android/app/src/main/AndroidManifest.xml`、`native/core/app_storage.go`。
   - Android SAF 選擇樹 URI 時，必須與 Go 的實體檔案路徑需求相容；不能把 SAF URI 直接交給需絕對路徑的下載管理器。遵守平台允許的目錄與授權範圍。
   - 下載目錄名稱：以當次經確認的 guo2 行為為準，不增加非預期的 `guo/` 子資料夾；若上游目錄結構改變則必須先驗證。
5. **合併功能（已重新列入）**：移植 guo2 已有的 Android 合併輸出位置與 SAF 授權：可選擇固定的合併輸出資料夾，必要時清除並恢復輸出到 Movies。選擇位置的紀錄由 `lib/merge_location.dart` 維護；`lib/merge_queue.dart` 在成功合併後使用設定的輸出路徑，沒有設定則維持原先 Android Movies 輸出。系統端對應 `MainActivity.kt` 的 `chooseMergeDirectory`、`createMergeDocument`、`copyFileToDocument`，設定頁 `storage_view.dart` 顯示與清除合併位置。
   - 保留既有合併佇列、背景作業及媒體格式相容檢查，不將完整舊版檔案盲目覆蓋 110 的其他功能。
   - 目前僅移植 **guo2 已有、可由程式核對的**合併設定與輸出位置；「合併後刪除已下載分集」和「單集可合併」如要升級為新行為，必須分開驗證安全性，不能由記錄宣稱已實作。
   - 合併失敗、使用者取消、文件寫入失敗都不得刪除原始下載檔案；成功後的暫存檔須由現有流程清除。

## 驗證／交付

- 先對比新 Release 的原始版本與新增檔案差異，確保原站源、原版播放器、合併及其他新版功能未被覆蓋。
- Go：`go test ./...`，將 `native/sources_custom` 連入後執行 `go test -tags=zhenguojian ./...`，驗證 13 個額外 Provider 的實例／ID 註冊。
- Flutter：`dart analyze --fatal-infos lib test` 與 `flutter test --dart-define=DISABLE_REMOTE_IMAGES=true --dart-define=ALL_SOURCES=true`。
- Android：依需求建置 ARM64／v7a 並核對包名／簽章。未經驗證不可宣稱 APK 可覆蓋安裝；未要求時不發布 GitHub Release。
- 如果功能移植或測試失敗，列出未完成項目與錯誤，勿自行減少來源或更動合併功能。

## 本次進度

- `guo/main` 已從 109 更新至當時最新 Release 110，並加入 13 個自訂 Provider；基準 commit 為 `ec6faf7f2e8adaba9cfc989cb34f8a6d701cf070`。
- 13 個新增來源的 Go 註冊測試及 Flutter 檢查已於獨立分支通過。
- UI／下載／授權／隱藏頻道／合併輸出位置等後續修改仍需以最新分支的測試結果為準，不得僅憑本文宣稱完成。
