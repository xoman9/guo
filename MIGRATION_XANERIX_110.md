# 真果鑑：以 Xanerix/guoapp 重建（工作分支）

## 來源與目標
- 上游主程式： https://github.com/Xanerix/guoapp (main)
- 既有共存版：xoman9/guo `fix/download-merge-lowest-20261101`，對應 1.0.10+110
- 目標：以 **Xanerix 的原始碼為主**，在 `xoman9/guo` 重建 ARM64 共存版；原 `main` 不修改。
- 保留上游 GPL-3.0 LICENSE 和原始作者資訊，不移除版權聲明。

## 需要移植的使用者操作
1. **下載位置**：保留已設定的下載目錄、顯示目前目錄、選擇目錄、儲存與搬移資料；Android SAF 與應用專屬目錄需分別處理。
2. **合併位置**：合併輸出位置可選擇；沿用 Xanerix 的合併檔案格式、檔案選擇與輸出流程，並確認單集與多集。
3. **頻道選擇**：點擊左側/標題站源名稱即可選擇；不得只依靠地球圖示。
4. **隱藏頻道**：單站源啟用/停用、集中切換適用站源、持久化並讓發現頁/搜尋頁/下載頁同步。
5. **其餘 1.0.10 介面**：下載佇列、已下載、合併狀態、離開合併視窗後進度、下載速度、已合併管理，逐項核對後才標記完成。

## 已確認的上游差異（2026-11-07）
- Xanerix 已實作 `lib/ui/settings/storage_view.dart` 下載目錄選擇與搬移。
- Xanerix 已實作 `lib/ui/downloads/merge_dialog.dart`；Android 使用系統的輸出檔案選擇。
- Xanerix 已實作 `lib/ui/explore/source_actions.dart` 選源，以及 `lib/ui/settings/sources_view.dart` 個別站源停用。
- Xanerix `lib/ui/widgets/source_bar.dart` 仍有獨立的地球圖示，須改成站源名稱可點選。
- Xanerix 的下載合併架構與舊版不同；禁止直接覆蓋整個 `lib/` 或 `native/`。

## 驗收要求
- 以 Xanerix 程式碼完整建置；`flutter analyze`、相關 Flutter/Go tests 通過。
- Android ARM64 共存套件名稱與 1.0.10 版一致（`com.duanju.duanju_app.coexist`，需驗證實際舊 APK）。
- 使用與舊版相同的簽章金鑰；若無金鑰，禁止宣稱可覆蓋更新。
- `versionCode` 大於 110，正式新版本不得假冒 1.0.10+110。
- 確認真實 APK 的 signing certificate、package name、ABI、已下載資料讀取，並實測以上功能。

## 狀態
此分支目前為**重建工作分支/規格與差異紀錄**，並非已完成移植，更未通過 APK 編譯驗證。上游程式碼尚未真正匯入，不能作為 Xanerix 重建版安裝。
