一鍵推送使用說明

1. 安裝 Git for Windows：https://git-scm.com/download/win
2. 完整解壓 ZIP，雙擊 PUSH_TO_GITHUB.bat。
3. 首次推送依 Git 提示用瀏覽器登入有 xoman9/guo 寫入權限的 GitHub 帳號。
4. 出現 Upload completed 即完成。

批次檔會複製專案到暫存 clone、提交並推送 main，不強制覆寫歷史。
後續修改原資料夾，再雙擊即可更新；原資料夾刪除的檔案不會自動從遠端刪除。
不要將憑證、Token 或私密資料放入專案。
目前工作流程僅支援 Actions 手動執行，推送不會自動編譯 APK。
建置入口：https://github.com/xoman9/guo/actions
此 ZIP 為原始碼與推送工具，不含編譯完成的 APK。
