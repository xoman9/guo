package missavrecovery

// The routes below are transcribed from the user-supplied WebHome HTML
// (ADAPTERS.missav.tabs). They are web page routes, not a TVBox JSON API.
type webHomeCategory struct{ ID, Name, Path string }
var webHomeCategories = []webHomeCategory{
 {"cnsub","中文字幕","/dm278/cn/chinese-subtitle"},
 {"new","最近更新","/dm539/cn/new"},
 {"release","新作上市","/dm635/cn/release"},
 {"leak","無碼流出","/dm817/cn/uncensored-leak"},
 {"today","今日熱門","/dm301/cn/today-hot"},
 {"weekly","本週熱門","/dm170/cn/weekly-hot"},
 {"monthly","本月熱門","/dm273/cn/monthly-hot"},
 {"vr","VR","/cn/genres/VR"},
 {"ensub","英文字幕","/dm21/cn/english-subtitle"},
 {"siro","SIRO","/dm36/cn/siro"},
 {"luxu","LUXU","/dm34/cn/luxu"},
 {"gana","GANA","/dm34/cn/gana"},
 {"maan","PRESTIGE PREMIUM","/dm1004/cn/maan"},
 {"scute","S-CUTE","/dm38/cn/scute"},
 {"ara","ARA","/dm34/cn/ara"},
 {"fc2","FC2","/dm597/cn/fc2"},
 {"heyzo","HEYZO","/dm2208642/cn/heyzo"},
 {"tokyohot","Tokyo Hot","/dm42/cn/tokyohot"},
 {"onepondo","1pondo","/dm5199603/cn/1pondo"},
 {"caribbeancom","Caribbeancom","/dm7704788/cn/caribbeancom"},
 {"caribbeancompr","Caribbeancompr","/dm91887/cn/caribbeancompr"},
 {"musume10","10musume","/dm7208981/cn/10musume"},
 {"pacopacomama","pacopacomama","/dm3600557/cn/pacopacomama"},
 {"gachinco","Gachinco","/dm150/cn/gachinco"},
 {"xxxav","XXX-AV","/dm42/cn/xxxav"},
 {"marriedslash","人妻斬","/dm37/cn/marriedslash"},
 {"naughty4610","頑皮4610","/dm33/cn/naughty4610"},
 {"naughty0930","頑皮0930","/dm37/cn/naughty0930"},
 {"madou","麻豆傳媒","/dm63/cn/madou"},
 {"twav","TWAV","/dm31/cn/twav"},
 {"furuke","Furuke","/dm15/cn/furuke"},
 {"klive","韓國直播","/cn/klive"},
 {"clive","中國直播","/cn/clive"},
}
func webHomeCategoryPath(id string)(string,bool){
 for _,c:=range webHomeCategories{if c.ID==id{return c.Path,true}}
 return "",false
}
