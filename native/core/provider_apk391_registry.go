package core

import "strings"

// Original 0.2.391+2414 ARM64 APK registry IDs and candidate domains.
// Registered identifiers do not prove a source is currently reachable.
type apk391SourceSpec struct{ ID, Name, Kind, Host string }

var apk391Sources = []apk391SourceSpec{
 {"yanguo","艳果","maccms","xqxq1.cc"},
 {"youguo","柚果","maccms","youavhub.com"},
 {"meiguo","美果","maccms","shiresm.lol"},
 {"yingguo","樱果","maccms","missav.app"},
 {"luguo","露果","maccms","luyitian.com"},
 {"liguo","荔果","maccms","ririlu.cc"},
 {"juguo","橘果","maccms","bulunhufait.buzz"},
 {"zaoguo","枣果","maccms","yhsp5.yachts"},
 {"ningguo","柠果","maccms","toptv15.cyou"},
 {"mangguo","芒果","maccms","zsrenqi.xyz"},
 {"chengguodj","橙果","web","chengguodj.com"},
 {"xiongguo","熊果","api","xiongmao17.com"},
 {"jiaoguo","角果","maccms","haijiao126.cc"},
 {"zhenguo","贞果","maccms","zhenshi27.xyz"},
 {"caoguo","草果","caoguo","ccctv01.top"},
 {"xueguo","学果","maccms","xyxs.buzz"},
 {"av2","二果","av2","2av.is"},
 {"huiguo","会果","maccms","guochan01.sbs"},
 {"liaoguo","撩果","maccms","xslrxx.mom"},
 {"chunguo","春果","maccms","lichunyuan.buzz"},
 {"boguo","波果","javbobo","javbobo.com"},
 {"cangguo","仓果","hsck","hsck.net"},
 {"baguo","巴果","javbus","javbus.com"},
 {"jieguo","杰果","jable","jable.tv"},
 {"dangguo","档果","javdb","javdb.com"},
 {"hanguo","汉果","hanime1","hanime1.me"},
 {"whostv","番果","whostv","whos.tv"},
 {"bangguo","榜果","javranking","javranking.cc"},
 {"tangguo","糖果","tangxin","tangxinvlog.pro"},
 {"bilifun","影果","bilifun","bili-fun.com"},
 {"fengguo","凤果","pornhub","cn.pornhub.com"},
 {"x134","推果","x134","134x.com"},
 {"manman","漫果","maccmsjson","1anime.me"},
 {"hsex","色果","hsex","hsex.tv"},
 {"uu","优果","uu","age.03b8mni.com"},
 {"rb91","九果","rb91","91rb.net"},
 {"dj51","壹果","dj51","51hub.com"},
 {"yeshe","夜果","yeshe","51lantian.org"},
 {"cg51","擦果","cgpost","d1yx0rpuwv3jis.cloudfront.net"},
 {"cg69","六果","cgpost","d39cgxtrya69tg.cloudfront.net"},
 {"maitian","田果","maitian","172.247.227.203"},
 {"yangguo","洋果","redtube","redtube.com"},
 {"badnews","爆果","badnews","bad.news"},
}

var apk391ByID = func() map[string]apk391SourceSpec {
 result:=make(map[string]apk391SourceSpec,len(apk391Sources))
 for _,spec:=range apk391Sources {
  if _,duplicate:=result[spec.ID];duplicate {panic("duplicate provider ID "+spec.ID)}
  result[spec.ID]=spec
 }
 return result
}()

func apk391Spec(source string)(apk391SourceSpec,bool) {
 spec,ok:=apk391ByID[strings.ToLower(strings.TrimSpace(source))]
 return spec,ok
}

func apk391SourceForHost(host string) string {
 host=strings.TrimPrefix(strings.ToLower(strings.TrimSpace(host)),"www.")
 for _,spec:=range apk391Sources {
  if host==strings.TrimPrefix(spec.Host,"www.") {return spec.ID}
 }
 return ""
}

func apk391BaseURL(source string) string {
 if spec,ok:=apk391Spec(source);ok{return "https://"+spec.Host}
 return ""
}
