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

## 預設隱藏頻道標示（2026-11-14）

- **所有 `defaultOff: true` 的頻道顯示名稱後加 ` +`**，包含 110 既有頻道與 guo2 新增頻道；原本已帶 `+` 者不可重複新增。
- 由 `lib/models.dart` 的 `SourceSite.fromJson` 統一處理 `name`、`groupName`、`entryName`，保留原始頻道 ID、`defaultOff` 及啟用狀態不變；原本 `defaultOff: false` 的頻道名稱保持原樣。
- 測試 `test/source_default_off_suffix_test.dart` 覆蓋已帶加號、未帶加號、一般站源及群組。
- **原生頻道遺失的獨立問題**：110 官方站源實作位於另外的私有 Go 模組 `guoapp/sources`，只編譯 `native/sources_custom` 的 13 個 Provider 不等於「全部原生站源」。任何正式包都必須先接入 110 原版站源模組，且額外接入 `native/sources_custom`，在產物中驗證兩者並存；無法取得原版模組時不得宣稱已修復、也不得發布標為完整站源的 APK。

## 使用條款確認畫面（2026-11-14）

- `lib/ui/terms_page.dart`：保留原有條款內容與 `termsVersion` 接受紀錄；刪除手動輸入「我已确认并同意以上使用条款」的文字欄，改為可勾選的「我已同意」`CheckboxListTile`。
- 提示文案採「請在下方勾選『我已同意』，再點擊『同意並繼續』。」；**未勾選時按鈕不可點擊，勾選後可立即點擊**，不再等待倒數，也不要求輸入確認字串。
- 手機與 Android TV 都需以明確勾選表達同意，不得在 TV 上繞過勾選。按下「同意並繼續」後仍走 `store.setAcceptedTerms(termsVersion)` 的既有保存機制，失敗時不視為已同意。
- 每次從最新穩定 Release 重套本項 UI 差異，重新執行 Dart 分析、Flutter 測試及 Android APK 建置驗證。

## 驗證／交付

- 先對比新 Release 的原始版本與新增檔案差異，確保原站源、原版播放器、合併及其他新版功能未被覆蓋。
- Go：`go test ./...`，將 `native/sources_custom` 連入後執行 `go test -tags=zhenguojian ./...`，驗證 13 個額外 Provider 的實例／ID 註冊。
- Flutter：`dart analyze --fatal-infos lib test` 與 `flutter test --dart-define=DISABLE_REMOTE_IMAGES=true --dart-define=ALL_SOURCES=true`。
- Android：依需求建置 ARM64／v7a 並核對包名／簽章。未經驗證不可宣稱 APK 可覆蓋安裝；未要求時不發布 GitHub Release。
- 如果功能移植或測試失敗，列出未完成項目與錯誤，勿自行減少來源或更動合併功能。

## APK 測試建置（ARM64 + ARMv7）

- `.github/workflows/build-guo-test-apks.yml` 為可重複執行的 GitHub Actions 建置流程，使用 `scripts/build_android.py --all-sources --abi arm64-v8a --abi armeabi-v7a`，包含真實 `native/sources_custom` 站源及 Go 核心，並檢查兩個 APK 的套件名稱、簽章和 native libraries。
- 本次共存版套件名稱採 `com.duanju.coexist`，顯示名稱 `小果影視-共存版`；升級至最新 Release 後，先確認包名和對應的版號設定是否仍正確。
- **工作流程預設沒有永久簽章私鑰**。產物為測試簽章，不能保證覆蓋安裝已存在的固定簽章版本。正式發佈前須用同一組原始永久簽章，在安全的金鑰管理環境中重新建置與驗證。
- 只上傳 GitHub Actions 的測試 artifact（保留 7 天），不自動建立 GitHub Release；必須完整驗證通過才提供測試 APK 的產物連結。
- `latest` 指的是每次**起始來源**，不是自動修改現有裝置，也不是保證每一次 upstream Release 都保有相同程式結構；不相容時逐項修正並記錄失敗。

## 原有永久固定簽章（已實測成功，2026-11-14）

- 使用 GitHub Secrets `ANDROID_KEYSTORE_BASE64`、`ANDROID_KEYSTORE_PASSWORD`、`ANDROID_KEY_ALIAS`、`ANDROID_KEY_PASSWORD` 提供**原本固定 JKS**。不要建立新金鑰來替代舊金鑰。
- 原本憑證 SHA-256：`78cf3a09001eba2c5b8e7b169d2af15301e76bf5c442af00c4031120dcb1fd6a`。產出時須檢查 `apksigner verify --verbose --print-certs` 的實際 APK 簽章指紋；不一致立即停止交付。
- 工作流程 `.github/workflows/build-guo-fixed-signed-apks.yml`：完整建置 ARM64／v7a 並使用原金鑰；`.github/workflows/resign-guo-test-apks-fixed-key.yml`：對已驗證、且程式來源未變更的兩個 APK 用原金鑰重新簽章，並驗證輸出。
- **2026-11-14 成功產物**：GitHub Actions run `43507510812`、artifact `guo-original-permanent-JKS-ARM64-v7a`，兩個 ABI 的 `com.duanju.coexist` APK 均通過原始憑證驗證與原生函式庫檢查。
- 過去 guo 歷史固定簽章版曾使用 `com.duanju.duanju_app.coexist`，和目前 `com.duanju.coexist` **套件 ID 不相同**；即使憑證相同也不能互相覆蓋安裝。之後如需覆蓋，必須比對實機既有 APK 的套件 ID 與憑證，不能只核對 SHA-256 簽章。
- JKS 內容與密碼不得列印到 Actions 日誌或提交到 GitHub 原始碼；只顯示公開憑證指紋。

## 110 原生頻道遺失：根因與阻擋規則（2026-11-14）

**已確認根因**：上游 1.0.10+110 公開儲存庫只有共用 Core 和示例站源；正式原生站源放在獨立的私有 Go 模組 `guoapp/sources`。2026-11-14 先前生成的 ARM64／v7a 固定簽章 APK（包括 Actions `43507510812`、`43511171430`）雖然程式及簽章驗證通過，**只內建了 `native/sources_custom` 的新增站源，未包含 110 全部原生頻道**，因此不得標示為完整／推薦更新版本。

- 官方原生站源的存取檢查：`43515278760`，`SOURCES_DEPLOY_KEY` 未設定，無法在 `xoman9/guo` 取回私有模組。
- 從 `guo2/custom` 回填較舊原生 Provider 的診斷：`43514969656`，與 110 的 Core 出現重複宣告，未通過 Go 編譯；只放在獨立診斷分支，**未整合進 main**，不得假裝已恢復 110 原版站源。
- `scripts/app_build.py`：若可合法取得 `guoapp-sources` 模組，依序載入 `guoapp/sources` 與 `native/sources_custom`；`GUO_REQUIRE_OFFICIAL_SOURCES=1` 時缺少任一模組就失敗。沒有官方模組的普通開發建置會明確警告僅含新增 13 站源。
- 正式固定簽章建置 `.github/workflows/build-guo-fixed-signed-apks.yml` 只接受受權的 `SOURCES_DEPLOY_KEY` 從 `Xanerix/guoapp-sources` 取得模組，並以 `--sources guoapp-sources --sources native/sources_custom` 一次載入；無權限不發布缺頻道的 APK。
- **待完成**：取得上游原生站源的合法原始碼／存取權限、通過雙模組 Go＋Flutter＋雙架構 APK 驗證、核對來源清單與實機下載播放後，才可重新發布完整修正版。簽章不變，保留現有客製設定。

## 本次進度

- `guo/main` 已從 109 更新至當時最新 Release 110，並加入 13 個自訂 Provider；基準 commit 為 `ec6faf7f2e8adaba9cfc989cb34f8a6d701cf070`。
- 13 個新增來源的 Go 註冊測試及 Flutter 檢查已於獨立分支通過。
- UI／下載／授權／隱藏頻道／合併輸出位置等後續修改仍需以最新分支的測試結果為準，不得僅憑本文宣稱完成。
