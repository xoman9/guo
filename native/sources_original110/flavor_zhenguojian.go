//go:build zhenguojian

package sources

import "duanjuapp/native/core"

func init() {
	core.RegisterFlavor("zhenguojian", []string{
		"hongguo", "hanxiaoquan", "guipian", "sorani",
		"huangdou", "huangju", "yeguo", "dsd", "jinguoduanju", "jinguo", "heitao", "beabox",
		"huangguo-video", "huangguoai", "cloudfront", "missav",
        // Extra opt-in sources must be in the single-core flavor allowlist.
        "lujj31", "liuguo", "xiaoguo", "youguo", "meiguo", "yingguo",
        "luguo", "liguo", "juguo", "ningguo", "mangguo", "chengguodj", "jiaoguo",
        // TVBox direct JSON APIs and independent Jable compatibility adapter.
        "tvbox-laoguo", "tvbox-jable", "tvbox-155", "tvbox-jkun", "tvbox-senlin", "tvbox-jingpinx",
	})
}
