package legacy109

import (
 "context"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"

 "duanjuapp/native/core"
)

// Extracted from the 109 heitao MacCMS route/playlist implementation and
// rewritten to use the public upstream 110 Host + Provider interfaces.
// This is a separate compatibility provider, not a copy into native/core.
const htBase = "https://heitaodj.com"
var htID = regexp.MustCompile(`^[0-9]{1,12}$`)
var htCategory = regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*/index\.php/vod/type/id/(\d+)\.html)["'][^>]*>(.*?)</a>`)
var htCards = regexp.MustCompile(`(?is)<a\b([^>]*class=["'][^"']*\bthumb\b[^"']*["'][^>]*)>(.*?)</a>`)
var htHref = regexp.MustCompile(`(?is)\bhref=["']([^"']+)["']`)
var htTitle = regexp.MustCompile(`(?is)\btitle=["']([^"']+)["']`)
var htImage = regexp.MustCompile(`(?is)<img\b[^>]*\b(?:data-original|data-src|src)=["']([^"']+)["']`)
var htDetail = regexp.MustCompile(`(?i)/index\.php/vod/detail/id/(\d+)\.html`)
var htEpisodes = regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*/index\.php/vod/play/id/(\d+)/sid/(\d+)/nid/(\d+)\.html)["'][^>]*>(.*?)</a>`)
var htNextPage = regexp.MustCompile(`(?i)/index\.php/vod/(?:show|search)/[^"'<>]*?/page/(\d+)\.html`)
var htHeading = regexp.MustCompile(`(?is)<h1\b[^>]*>(.*?)</h1>`)

var heitaoDescriptor = core.SourceDescriptor{
 ID:"heitao", Name:"黑桃 +",
 Description:"109 MacCMS 路線的 110 實驗相容版；尚未實機播放驗證",
 DefaultOff:true,BaseURL:htBase,
 NoHostRotation:true,
 ValidCategory:func(c string)bool{return c==""||htID.MatchString(c)},
}
type heitaoProvider struct{host core.Host}
var _ core.Provider = (*heitaoProvider)(nil)
var _ core.Searcher = (*heitaoProvider)(nil)
func init(){core.Register(heitaoDescriptor,func(host core.Host)core.Provider{return &heitaoProvider{host}})}
func (p *heitaoProvider) Descriptor()core.SourceDescriptor{return heitaoDescriptor}
func (p *heitaoProvider) BaseURL()string{return htBase}
func (p *heitaoProvider) get(ctx context.Context,path string)(string,error){
 if !strings.HasPrefix(path,"/index.php/") && path!="/" {return "",errors.New("invalid heitao URL route")}
 r,err:=p.host.FetchText(ctx,core.TextRequest{
  URL:htBase+path,Referer:htBase+"/",UserAgent:ua,Cookie:"heitao_age_confirmed=1",
 })
 return r.Body,err
}
func (p *heitaoProvider) Categories(ctx context.Context,force bool)([]core.Category,error){
 body,err:=p.get(ctx,"/")
 if err!=nil{return nil,err}
 cats:=[]core.Category{{ID:"",Name:"全部"}}
 seen:=map[string]bool{}
 for _,m:=range htCategory.FindAllStringSubmatch(body,-1){
  if len(m)<4||!htID.MatchString(m[2])||seen[m[2]]{continue}
  name:=core.CleanText(m[3])
  if name==""{continue}
  cats=append(cats,core.Category{ID:m[2],Name:name})
  seen[m[2]]=true
  if len(cats)>=33{break}
 }
 if len(cats)<2{return cats,errors.New("heitao categories not found")}
 return cats,nil
}
func heitaoCatalogRoute(page int,category,search string)(string,error){
 if page<1||page>100000||len(search)>256{return "",errors.New("invalid page")}
 if category!=""&&!htID.MatchString(category){return "",errors.New("invalid category")}
 if search!=""{
  word:=url.PathEscape(strings.TrimSpace(search))
  if page==1{return "/index.php/vod/search/wd/"+word+".html",nil}
  return fmt.Sprintf("/index.php/vod/search/page/%d/wd/%s.html",page,word),nil
 }
 prefix:="/index.php/vod/show"
 if category!=""{prefix+="/id/"+category}
 if page==1{return prefix+".html",nil}
 return fmt.Sprintf("%s/page/%d.html",prefix,page),nil
}
func (p *heitaoProvider) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 path,err:=heitaoCatalogRoute(page,category,"")
 if err!=nil{return nil,false,err}
 body,err:=p.get(ctx,path)
 if err!=nil{return nil,false,err}
 return heitaoParseCards(body),heitaoPageAfter(body,page),nil
}
func (p *heitaoProvider) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 path,err:=heitaoCatalogRoute(page,"",query)
 if err!=nil{return nil,false,err}
 body,err:=p.get(ctx,path)
 if err!=nil{return nil,false,err}
 return heitaoParseCards(body),heitaoPageAfter(body,page),nil
}
func heitaoPageAfter(body string,page int)bool{
 for _,m:=range htNextPage.FindAllStringSubmatch(body,-1){
  n,err:=strconv.Atoi(m[1])
  if err==nil&&n>page&&n<100000{return true}
 }
 return false
}
func heitaoParseCards(body string)[]core.Drama{
 seen:=map[string]bool{}
 var result []core.Drama
 for _,m:=range htCards.FindAllStringSubmatch(body,-1){
  if len(m)<3{continue}
  link:=htHref.FindStringSubmatch(m[1])
  if len(link)<2{continue}
  idMatch:=htDetail.FindStringSubmatch(link[1])
  if len(idMatch)<2{continue}
  id:=idMatch[1]
  if seen[id]{continue}
  title:=""
  if t:=htTitle.FindStringSubmatch(m[1]);len(t)>1{title=core.CleanText(t[1])}
  if title==""{title=core.CleanText(m[2])}
  if title==""{continue}
  cover:=""
  if image:=htImage.FindStringSubmatch(m[2]);len(image)>1{cover=core.ResolveProviderURL(htBase,image[1])}
  seen[id]=true
  result=append(result,core.Drama{
   ID:core.ProviderDramaID("heitao",id),Source:"heitao",
   SourceID:id,Title:title,Name:title,Cover:cover,
  })
  if len(result)>=500{break}
 }
 return result
}
func (p *heitaoProvider) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !htID.MatchString(id){return core.Drama{},nil,errors.New("invalid heitao ID")}
 body,err:=p.get(ctx,"/index.php/vod/detail/id/"+id+".html")
 if err!=nil{return core.Drama{},nil,err}
 name:=""
 if m:=htHeading.FindStringSubmatch(body);len(m)>1{name=core.CleanText(m[1])}
 if name==""{name=core.ExtractPageTitle(body)}
 if name==""{return core.Drama{},nil,errors.New("heitao title unavailable")}
 episodes:=heitaoParseEpisodes(body,id)
 if len(episodes)==0{return core.Drama{},nil,errors.New("heitao episodes unavailable")}
 cover:=""
 if m:=htImage.FindStringSubmatch(body);len(m)>1{cover=core.ResolveProviderURL(htBase,m[1])}
 drama:=core.Drama{ID:core.ProviderDramaID("heitao",id),Source:"heitao",
  SourceID:id,Title:name,Name:name,Cover:cover,Desc:core.ExtractDescription(body),
  TotalEpisode:len(episodes)}
 return drama,episodes,nil
}
func heitaoParseEpisodes(body,id string)[]core.Chapter{
 type episode struct{sid,nid int;title string}
 groups:=map[int][]episode{}
 seen:=map[string]bool{}
 for _,m:=range htEpisodes.FindAllStringSubmatch(body,-1){
  if len(m)<6||m[2]!=id{continue}
  sid,_:=strconv.Atoi(m[3]);nid,_:=strconv.Atoi(m[4])
  if sid<1||nid<1{continue}
  key:=fmt.Sprintf("%d:%d",sid,nid)
  if seen[key]{continue}
  seen[key]=true
  title:=core.CleanText(m[5])
  if title==""{title=fmt.Sprintf("第%d集",nid)}
  groups[sid]=append(groups[sid],episode{sid,nid,title})
 }
 chosen:=0
 for sid,group:=range groups{
  if len(group)>len(groups[chosen])||(len(group)==len(groups[chosen])&&len(group)>0&&(chosen==0||sid<chosen)){chosen=sid}
 }
 var chapters []core.Chapter
 for _,e:=range groups[chosen]{
  page:=fmt.Sprintf("%s/index.php/vod/player/id/%s/nid/%d/sid/%d.html",htBase,id,e.nid,e.sid)
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("heitao",id,fmt.Sprintf("%d:%d",e.sid,e.nid)),
   Source:"heitao",Title:e.title,CurrentEpisode:core.RawEpisode(e.nid),
   PageURL:page,Referer:htBase+"/",
  })
 }
 core.SortProviderChapters(chapters)
 return chapters
}
// heitaoReadPlayerConfig109 preserves the 109 implementation's balanced
// brace scanning instead of a regex that truncates nested player JSON.
func heitaoReadPlayerConfig109(body string)(struct{URL string `json:"url"`; Encrypt any `json:"encrypt"`},error){
 var entry struct{URL string `json:"url"`;Encrypt any `json:"encrypt"`}
 marker:=strings.Index(body,"player_aaaa")
 if marker<0{return entry,errors.New("heitao player settings unavailable")}
 text:=body[marker:]
 start:=strings.Index(text,"{")
 if start<0{return entry,errors.New("heitao player settings invalid")}
 depth:=0
 quote:=byte(0)
 escaped:=false
 end:=-1
 for i:=start;i<len(text);i++{
  b:=text[i]
  if quote!=0{
   if escaped{escaped=false;continue}
   if b=='\\'{escaped=true;continue}
   if b==quote{quote=0}
   continue
  }
  switch b{
   case '"','\'':quote=b
   case '{':depth++
   case '}':
    depth--
    if depth==0{end=i;i=len(text)}
  }
 }
 if end<start{return entry,errors.New("heitao player JSON incomplete")}
 if err:=json.Unmarshal([]byte(text[start:end+1]),&entry);err!=nil{return entry,err}
 if strings.TrimSpace(entry.URL)==""{return entry,errors.New("heitao returned empty player")}
 return entry,nil
}

// heitaoPlayerRoute validates that the selected drama, chapter and player
// route all identify the same series/playlist/episode before fetching.
func heitaoPlayerRoute(task core.Task)(string,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!="heitao"||!htID.MatchString(id){
  return "",errors.New("invalid heitao series")
 }
 prefix:=core.ProviderChapterID("heitao",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){
  return "",errors.New("heitao chapter belongs to another series")
 }
 parts:=strings.Split(strings.TrimPrefix(task.Chapter.ID,prefix),":")
 if len(parts)!=2||!htID.MatchString(parts[0])||!htID.MatchString(parts[1]){
  return "",errors.New("invalid heitao playlist or episode")
 }
 sid,nid:=parts[0],parts[1]
 route:=fmt.Sprintf("/index.php/vod/player/id/%s/nid/%s/sid/%s.html",id,nid,sid)
 if task.Chapter.PageURL!=htBase+route{
  return "",errors.New("heitao chapter player URL does not match episode")
 }
 return route,nil
}
func (p *heitaoProvider) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 page:=task.Chapter.PageURL
 route,verifyErr:=heitaoPlayerRoute(task)
 if verifyErr!=nil{return core.ProviderMedia{},verifyErr}
 body,err:=p.get(ctx,route)
 if err!=nil{return core.ProviderMedia{},err}
 entry,err:=heitaoReadPlayerConfig109(body)
 if err!=nil{return core.ProviderMedia{},err}
 mode:=fmt.Sprint(entry.Encrypt)
 address:=strings.TrimSpace(entry.URL)
 switch mode {
 case "0","<nil>":
 case "1":
  address,err=url.QueryUnescape(address)
  if err!=nil{return core.ProviderMedia{},err}
 case "2":
  var b []byte
  b,err=base64.StdEncoding.DecodeString(address)
  if err!=nil{return core.ProviderMedia{},err}
  address=string(b)
 default:return core.ProviderMedia{},errors.New("unsupported heitao player codec")
 }
 if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("invalid heitao media")}
 media:=core.ProviderMedia{URL:address,Referer:htBase+"/"}
 if strings.Contains(strings.ToLower(address),".m3u8"){return p.host.PrepareWebMedia(ctx,media,"黑桃")}
 return media,nil
}
