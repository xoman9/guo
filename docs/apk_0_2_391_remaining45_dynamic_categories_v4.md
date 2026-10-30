# 0.2.391 APK：其餘 45 頻道的動態分類反組譯 V4

> 此紀錄僅為使用者授權之 `zhenguojian-0.2.391+2414-arm64-v8a(1).apk` 靜態分析成果；沒有確認網域或串流目前可用。

## 範圍及進度
- **不修改／不重新移植**：桃果 `taoguo`、榴果 `liuguo`、宵果 `xiaoguo`。
- 其餘 45 個頻道維持實際 APK 註冊 ID；新增時名稱才加單一 `🔞`。
- 已從 Go `map.init.17` 對應到 19 個來源的 **436 筆**分類。
- 另外 **26 個**沒有在這張原生靜態分類表裡找到分類項目，不代表站點沒有分類。
- 本次再從 ARM64 `(*Downloader).fetchDuanjuCategories` 的直接 `bl` 呼叫，證明 **其中 6 個**另有專用分類入口。**剩餘 20 個**尚無中央分類函式的直接呼叫證據；可能是通用模板或僅「全部」目錄。

## 六個分類函式

| 頻道 | APK ID | 原生專用函式 | 直接還原的分類路徑常數 |
|---|---|---|---|
| 熊果 🔞 | `xiongguo` | `fetchXiongguoCategories` | `/getDataInit` |
| 九果 🔞 | `rb91` | `fetchRB91Categories` | `/videos/classification` |
| 優果 🔞 | `uu` | `fetchUUCategories` | `/api/v2/home/public/nav/list` |
| 鳳果 🔞 | `fengguo` | `fetchPornhubCategories` | `/categories` |
| 糖果 🔞 | `tangguo` | `fetchTangguoCategories` | 對來源首頁讀取文字；完整分類路徑待驗證 |
| 漫果 🔞 | `manman` | `fetchManmanCategories` | 經 `manmanFetch` 取 JSON；完整路徑待驗證 |

路徑常數來源是 Go 字串指標與長度的 ARM64 指令以及 ELF 實際位元組；例如：
- `xiongguo`: 指令 `0x2e6c9ac`、字串 VA `0x4a0009`、長度 12。
- `rb91`: `0x2e4af30`、VA `0x4b6a9d`、長度 22。
- `uu`: `0x2e5ece8`、VA `0x4c6277`、長度 28。
- `fengguo`: `0x2e48558`、VA `0x49c2cc`、長度 11。

中央分類分派函式地址：`0x2e0b560–0x2e0bf40`。本輪復核上述四條文字常數，以及源函式中原生 `bl` 指令的目標地址。

**這些地址和路徑只能證明 APK 曾經包含這套請求邏輯。還沒有確認 HTTP 方法、headers、token、回應格式或實際站點的分類／影片／播放。**

## 尚待完成
1. 6 個專用分類入口的 HTTP 方法與 JSON/HTML 回應資料結構。
2. 20 個其他來源是否由通用 MacCMS 或別的解析器提供分類；缺乏證據不應自行製造分類。
3. 各來源封面、詳情、播放與即時網站連線核對；未驗證前不在 APP 中顯示可用站源。
4. 完成測試後再接入獨立 ARM64 共存版，保留紅果及已完成三個來源核心。

## 可重現成果
對應 V4 附件 `zhenguo_0391_dynamic_category_v4.zip`：
- `26來源_動態分類追查結果.csv`
- `分類分派_ARM64直接呼叫證據.csv`
- `分類請求_原生路徑與證據.csv`
- 交叉核對及重建的 Python 腳本，與完整 APK SHA-256：

`79dda2f076375a9cfb8fa8dca221f9b8a62ed5babc139e04c9420a51a7b2998d`
