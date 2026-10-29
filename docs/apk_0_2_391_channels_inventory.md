# 真果鑑 0.2.391+2414（ARM64）截圖頻道清單

來源：使用者提供的 APK 內 `lib/arm64-v8a/libduanju_core.so` 原生字串，以及 2026-10-29 站源選擇畫面。此為**抽取清單，不是可使用的站源介面清單**。

**UI 命名規則：所有真正接入的頻道名稱皆以單一字元 `🔞` 結尾**；用例如 `超果 🔞`，避免以兩個字元的 ASCII `18+` 取代。

共有 7 排、48 格，其中 44 格可直接讀出名字；第一排第 2、3、5 格與第二排第 1 格被遮蔽，這 4 格依 APK 內連續字串順序推定，尚未經畫面直接驗證。48 個名稱的字串均可在 APK 原生核心中找到。

| 排 | 名稱（從左至右） |
| --- | --- |
| 1 | 超果 🔞、艳果 🔞（遮擋推定）、桃果 🔞（遮擋推定）、柚果 🔞、榴果 🔞（遮擋推定）、美果 🔞、城果 🔞 |
| 2 | 宵果 🔞（遮擋推定）、樱果 🔞、露果 🔞、荔果 🔞、橘果 🔞、枣果 🔞、柠果 🔞 |
| 3 | 芒果 🔞、橙果 🔞、熊果 🔞、角果 🔞、贞果 🔞、草果 🔞、学果 🔞 |
| 4 | 二果 🔞、会果 🔞、撩果 🔞、春果 🔞、波果 🔞、仓果 🔞、巴果 🔞 |
| 5 | 杰果 🔞、档果 🔞、汉果 🔞、番果 🔞、榜果 🔞、糖果 🔞、影果 🔞 |
| 6 | 凤果 🔞、推果 🔞、漫果 🔞、色果 🔞、优果 🔞、九果 🔞、壹果 🔞 |
| 7 | 夜果 🔞、擦果 🔞、六果 🔞、田果 🔞、洋果 🔞、爆果 🔞 |

## 分類與接入狀態

- **分類尚未完整恢復。** 此 APK 的分類來源可能是執行時連網動態解析；僅有編譯過的 `libduanju_core.so` 並不構成每一站源的可讀取分類 ID / URL / 分頁 / 影片詳情 / 播放 API。不可將畫面頻道名稱視為已完成分類移植。
- 桃果、榴果、宵果已在本分支建立獨立即時解析器並加入 `🔞`；功能仍以後續實機測試為準。
- 其餘來源需逐個建立經驗證的分類、分頁、封面、詳情、播放路由後才能在 `SourceSite.values` 啟用，避免產生「有頻道無分類／不可播放」的虛假入口。
- 不要替換現有紅果、榴果、桃果、宵果原生播放核心；不從 APK 直接覆蓋整個 `libduanju_core.so`，以免版本不相容與回歸。

## 機器可讀資料

```json
[
  {
    "row": 1,
    "col": 1,
    "name": "超果",
    "display_name": "超果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 1,
    "col": 2,
    "name": "艳果",
    "display_name": "艳果 🔞",
    "visible": false,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 1,
    "col": 3,
    "name": "桃果",
    "display_name": "桃果 🔞",
    "visible": false,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "此前已設獨立解析器；實機仍需驗證"
  },
  {
    "row": 1,
    "col": 4,
    "name": "柚果",
    "display_name": "柚果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 1,
    "col": 5,
    "name": "榴果",
    "display_name": "榴果 🔞",
    "visible": false,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "此前已設獨立解析器；實機仍需驗證"
  },
  {
    "row": 1,
    "col": 6,
    "name": "美果",
    "display_name": "美果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 1,
    "col": 7,
    "name": "城果",
    "display_name": "城果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 1,
    "name": "宵果",
    "display_name": "宵果 🔞",
    "visible": false,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "此前已設獨立解析器；實機仍需驗證"
  },
  {
    "row": 2,
    "col": 2,
    "name": "樱果",
    "display_name": "樱果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 3,
    "name": "露果",
    "display_name": "露果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 4,
    "name": "荔果",
    "display_name": "荔果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 5,
    "name": "橘果",
    "display_name": "橘果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 6,
    "name": "枣果",
    "display_name": "枣果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 2,
    "col": 7,
    "name": "柠果",
    "display_name": "柠果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 1,
    "name": "芒果",
    "display_name": "芒果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 2,
    "name": "橙果",
    "display_name": "橙果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 3,
    "name": "熊果",
    "display_name": "熊果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 4,
    "name": "角果",
    "display_name": "角果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 5,
    "name": "贞果",
    "display_name": "贞果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 6,
    "name": "草果",
    "display_name": "草果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 3,
    "col": 7,
    "name": "学果",
    "display_name": "学果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 1,
    "name": "二果",
    "display_name": "二果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 2,
    "name": "会果",
    "display_name": "会果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 3,
    "name": "撩果",
    "display_name": "撩果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 4,
    "name": "春果",
    "display_name": "春果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 5,
    "name": "波果",
    "display_name": "波果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 6,
    "name": "仓果",
    "display_name": "仓果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 4,
    "col": 7,
    "name": "巴果",
    "display_name": "巴果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 1,
    "name": "杰果",
    "display_name": "杰果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 2,
    "name": "档果",
    "display_name": "档果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 3,
    "name": "汉果",
    "display_name": "汉果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 4,
    "name": "番果",
    "display_name": "番果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 5,
    "name": "榜果",
    "display_name": "榜果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 6,
    "name": "糖果",
    "display_name": "糖果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 5,
    "col": 7,
    "name": "影果",
    "display_name": "影果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 1,
    "name": "凤果",
    "display_name": "凤果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 2,
    "name": "推果",
    "display_name": "推果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 3,
    "name": "漫果",
    "display_name": "漫果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 4,
    "name": "色果",
    "display_name": "色果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 5,
    "name": "优果",
    "display_name": "优果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 6,
    "name": "九果",
    "display_name": "九果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 6,
    "col": 7,
    "name": "壹果",
    "display_name": "壹果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 1,
    "name": "夜果",
    "display_name": "夜果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 2,
    "name": "擦果",
    "display_name": "擦果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 3,
    "name": "六果",
    "display_name": "六果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 4,
    "name": "田果",
    "display_name": "田果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 5,
    "name": "洋果",
    "display_name": "洋果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  },
  {
    "row": 7,
    "col": 6,
    "name": "爆果",
    "display_name": "爆果 🔞",
    "visible": true,
    "category_status": "未取得該來源動態分類資料",
    "integration_status": "尚未移植，禁止顯示失效頻道"
  }
]
```
