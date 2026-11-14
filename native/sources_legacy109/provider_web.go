package legacy109

import (
 "context"
 "errors"
 "fmt"
 "html"
 "net/url"
 "regexp"
 "strconv"
 "strings"

 "duanjuapp/native/core"
)

// Rebuilt from the working legacy 109 Go parsers in xoman9/guo2/custom.
// Uses the exported upstream 110 Provider/Host interfaces, not private
// Downloader fields; avoids the old core's duplicate global symbols.
const ua = "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36"
var idPattern=regexp.MustCompile(`^[0-9]{1,12}$`)
var hCard=regexp.MustCompile(`(?is)<a\s+href="/hanxiaoquan/(\d+)\.html"[^>]*title="([^"]*)"[^>]*class="module-item-pic"[^>]*>`)
var gCard=regexp.MustCompile(`(?is)<a\s+class="video-link"\s+href="/nv/(\d+)\.html"[^>]*title="([^"]*)"([^>]*)>`)
var hImg=regexp.MustCompile(`(?is)<img[^>]*data-src="([^"]+)"`)
var gImg=regexp.MustCompile(`(?is)data-(?:original|background)="([^"]+)"`)
var hTitle=regexp.MustCompile(`(?is)<h1\s+class="page-title"[^>]*>(.*?)</h1>`)
var gTitle=regexp.MustCompile(`(?is)<h1\s+class="media-title[^"]*"[^>]*>(.*?)</h1>`)
var hEpisode=regexp.MustCompile(`(?is)<a\s+title="([^"]*)"\s+href="(/play/\d+-\d+-\d+\.html)"`)
var gEpisode=regexp.MustCompile(`(?is)<li><a\s+title='([^']*)'\s+href='([^']+)'`)
var nowVar=regexp.MustCompile(`(?is)var\s+now\s*=\s*["']([^"']+)["']`)
var mediaAny=regexp.MustCompile(`(?i)https?://[^"'\s<>\\]+\.(?:m3u8|mp4)(?:\?[^"'\s<>\\]*)?`)
var gRssItem=regexp.MustCompile(`(?is)<item>(.*?)</item>`)
var gRssTitle=regexp.MustCompile(`(?is)<title>\s*(?:<!\[CDATA\[)?(.*?)(?:\]\]>)?\s*</title>`)
var gRssID=regexp.MustCompile(`/nv/(\d+)\.html`)

type spec struct {
 id, name, base string
 cats []core.Category
}
var korean = spec{"hanxiaoquan","韩小圈","https://www.jennyhow.com",[]core.Category{
 {ID:"1",Name:"最新韩剧"},{ID:"2",Name:"韩国电影"},
 {ID:"3",Name:"韩国综艺"},{ID:"4",Name:"韩国动漫"},
}}
var ghost = spec{"guipian","鬼片网","https://guipianwu.com",[]core.Category{
 {ID:"3",Name:"恐怖片"},{ID:"6",Name:"大陆鬼片"},{ID:"7",Name:"日韩鬼片"},
 {ID:"8",Name:"林正英鬼片"},{ID:"9",Name:"港台鬼片"},{ID:"10",Name:"泰国鬼片"},
 {ID:"11",Name:"欧美鬼片"},{ID:"31",Name:"剧情片"},{ID:"12",Name:"国产剧"},
 {ID:"13",Name:"美剧"},{ID:"14",Name:"韩剧"},{ID:"15",Name:"日剧"},
 {ID:"16",Name:"泰剧"},{ID:"17",Name:"港台剧"},{ID:"18",Name:"其他剧"},
 {ID:"23",Name:"日韩动漫"},{ID:"24",Name:"国产动漫"},
 {ID:"25",Name:"欧美动漫"},{ID:"26",Name:"港台动漫"},
}}

type provider struct { host core.Host; cfg spec }
var _ core.Provider = (*provider)(nil)
var _ core.Searcher = (*provider)(nil)
func (p *provider) validCat(v string)bool {
 if v=="" {return true}
 for _,cat:=range p.cfg.cats {if cat.ID==v{return true}}
 return false
}
func (p *provider) Descriptor() core.SourceDescriptor{
 return core.SourceDescriptor{
  ID:p.cfg.id,Name:p.cfg.name,BaseURL:p.cfg.base,
  Description:"依 109 原生 Provider 重新接入 110 的實驗相容版",
  ValidCategory:p.validCat,
 }
}
func (p *provider) BaseURL()string{return p.cfg.base}
func init(){
 for _,cfg:=range []spec{korean,ghost}{
  c:=cfg
  descriptor:=(&provider{cfg:c}).Descriptor()
  core.Register(descriptor,func(host core.Host)core.Provider{
   return &provider{host:host,cfg:c}
  })
 }
}
func (p *provider) Categories(context.Context,bool)([]core.Category,error){
 return append([]core.Category(nil),p.cfg.cats...),nil
}
func (p *provider) get(ctx context.Context,path string)(string,error){
 addr:=p.cfg.base+path
 if strings.HasPrefix(path,"https://") {addr=path}
 u,err:=url.Parse(addr)
 if err!=nil || u.Hostname()!=strings.TrimPrefix(strings.TrimPrefix(p.cfg.base,"https://"),"www.") &&
   u.Hostname()!=strings.TrimPrefix(p.cfg.base,"https://"){
  return "",errors.New("unexpected legacy109 source host")
 }
 r,err:=p.host.FetchText(ctx,core.TextRequest{URL:addr,Referer:p.cfg.base+"/",UserAgent:ua})
 return r.Body,err
}
func clean(raw string)string{return strings.TrimSpace(html.UnescapeString(core.CleanText(raw)))}
func (p *provider) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1 || page>100000 || !p.validCat(category){return nil,false,errors.New("invalid category or page")}
 if category==""{category=p.cfg.cats[0].ID}
 path:=""
 if p.cfg.id=="hanxiaoquan"{
  path="/hxq/"+category+".html"
  if page>1 {path=fmt.Sprintf("/hxq/%s-%d.html",category,page)}
 }else{
  path="/list/"+category+".html"
  if page>1 {path=fmt.Sprintf("/list/%s_%d.html",category,page)}
 }
 body,err:=p.get(ctx,path)
 if err!=nil{return nil,false,err}
 rows:=p.parseCards(body)
 paging:="/hxq/"+category+"-"
 if p.cfg.id=="guipian" {paging="/list/"+category+"_"}
 hasNext:=strings.Contains(body,paging+strconv.Itoa(page+1)+".html")||len(rows)>=24
 return rows,hasNext,nil
}
func (p *provider) parseCards(body string)[]core.Drama{
 var reg *regexp.Regexp
 if p.cfg.id=="hanxiaoquan"{reg=hCard}else{reg=gCard}
 seen:=map[string]bool{}
 var rows []core.Drama
 for _,m:=range reg.FindAllStringSubmatchIndex(body,-1){
  if len(m)<6 {continue}
  id,title:=body[m[2]:m[3]],clean(body[m[4]:m[5]])
  if !idPattern.MatchString(id)||title==""||seen[id]{continue}
  chunk:=body[m[1]:]
  if len(chunk)>650 {chunk=chunk[:650]}
  cover:=""
  img:=gImg
  if p.cfg.id=="hanxiaoquan"{img=hImg}
  if imgMatch:=img.FindStringSubmatch(chunk);len(imgMatch)>1{
   cover=core.ResolveProviderURL(p.cfg.base+"/",imgMatch[1])
  }
  seen[id]=true
  rows=append(rows,core.Drama{
   ID:core.ProviderDramaID(p.cfg.id,id),Source:p.cfg.id,
   SourceID:id,Title:title,Name:title,Cover:cover,
  })
 }
 return rows
}
func (p *provider) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if page<1 || len(query)>256 {return nil,false,errors.New("invalid search")}
 if query=="" {return nil,false,nil}
 if p.cfg.id=="hanxiaoquan"{
  if page>1{return nil,false,nil}
  body,err:=p.get(ctx,"/search.php?"+url.Values{"searchword":{query}}.Encode())
  if err!=nil{return nil,false,err}
  return p.parseCards(body),false,nil
 }
 body,err:=p.get(ctx,"/xml/rss.xml")
 if err!=nil{return nil,false,err}
 var found []core.Drama
 seen:=map[string]bool{}
 for _,m:=range gRssItem.FindAllStringSubmatch(body,-1){
  title:=gRssTitle.FindStringSubmatch(m[1]);id:=gRssID.FindStringSubmatch(m[1])
  if len(title)<2||len(id)<2||!idPattern.MatchString(id[1])||seen[id[1]]{continue}
  text:=clean(title[1])
  if !strings.Contains(strings.ToLower(text),strings.ToLower(query)){continue}
  seen[id[1]]=true
  found=append(found,core.Drama{ID:core.ProviderDramaID(p.cfg.id,id[1]),
   Source:p.cfg.id,SourceID:id[1],Title:text,Name:text})
 }
 start:=(page-1)*24
 if start>=len(found){return nil,false,nil}
 end:=start+24
 if end>len(found){end=len(found)}
 return found[start:end],end<len(found),nil
}
func (p *provider) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !idPattern.MatchString(id){return core.Drama{},nil,errors.New("invalid source ID")}
 path:="/hanxiaoquan/"+id+".html"
 if p.cfg.id=="guipian"{path="/nv/"+id+".html"}
 body,err:=p.get(ctx,path)
 if err!=nil{return core.Drama{},nil,err}
 titleReg:=hTitle
 if p.cfg.id=="guipian"{titleReg=gTitle}
 title:=""
 if m:=titleReg.FindStringSubmatch(body);len(m)>1{title=clean(m[1])}
 if title==""{title=core.ExtractPageTitle(body)}
 if title==""{return core.Drama{},nil,errors.New("legacy109 detail title missing")}
 episodeReg:=hEpisode
 if p.cfg.id=="guipian"{episodeReg=gEpisode}
 seen:=map[string]bool{}
 var episodes []core.Chapter
 for _,m:=range episodeReg.FindAllStringSubmatch(body,-1){
  ep:=clean(m[1]);link:=core.ResolveProviderURL(p.cfg.base+"/",m[2])
  if !p.validPlayURL(link)||seen[link]{continue}
  seen[link]=true
  index:=len(episodes)+1
  if ep==""{ep=fmt.Sprintf("第%d集",index)}
  episodes=append(episodes,core.Chapter{
   ID:core.ProviderChapterID(p.cfg.id,id,strconv.Itoa(index)),
   Source:p.cfg.id,Title:ep,PageURL:link,
   Referer:p.cfg.base+"/",CurrentEpisode:core.RawEpisode(index),
  })
 }
 if len(episodes)==0{return core.Drama{},nil,errors.New("legacy109 detail episodes unavailable")}
 cover:=""
 m:=regexp.MustCompile(`(?is)property=["']og:image["'][^>]*content=["']([^"']+)`).FindStringSubmatch(body)
 if len(m)>1 {cover=core.ResolveProviderURL(p.cfg.base+"/",m[1])}
 drama:=core.Drama{ID:core.ProviderDramaID(p.cfg.id,id),Source:p.cfg.id,
  SourceID:id,Title:title,Name:title,TotalEpisode:len(episodes),Cover:cover,
  Desc:core.ExtractDescription(body)}
 return drama,episodes,nil
}
func (p *provider) validPlayURL(raw string)bool{
 u,err:=url.Parse(raw)
 if err!=nil || u.Scheme!="https"{return false}
 host:=u.Hostname()
 if host!=strings.TrimPrefix(p.cfg.base,"https://") &&
    host!="www."+strings.TrimPrefix(p.cfg.base,"https://") &&
    host!=strings.TrimPrefix(strings.TrimPrefix(p.cfg.base,"https://"),"www.") {return false}
 return strings.HasSuffix(u.Path,".html")&&strings.Contains(u.Path,"/play/")
}
// verifyChapterForDrama ties the selected playback page to the episode
// list returned by the currently selected drama, not just the source host.
func (p *provider) verifyChapterForDrama(ctx context.Context, task core.Task) error {
 source, id, ok := core.SplitProviderDramaID(task.DramaID)
 if !ok || source != p.cfg.id || !idPattern.MatchString(id) {
  return errors.New("legacy109 playback drama invalid")
 }
 if !p.validPlayURL(task.Chapter.PageURL) {
  return errors.New("legacy109 player URL invalid")
 }
 if !strings.HasPrefix(task.Chapter.ID, core.ProviderChapterID(p.cfg.id,id,"")) {
  return errors.New("legacy109 chapter belongs to another drama")
 }
 _, chapters, err := p.Detail(ctx,id)
 if err != nil { return fmt.Errorf("legacy109 episode verification failed: %w",err) }
 for _, chapter := range chapters {
  if chapter.ID == task.Chapter.ID && chapter.PageURL == task.Chapter.PageURL {
   return nil
  }
 }
 return errors.New("legacy109 chapter not found in selected drama")
}
func (p *provider) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 page:=task.Chapter.PageURL
 if err:=p.verifyChapterForDrama(ctx,task);err!=nil{return core.ProviderMedia{},err}
 body,err:=p.get(ctx,page)
 if err!=nil{return core.ProviderMedia{},err}
 media:=""
 if m:=nowVar.FindStringSubmatch(body);len(m)>1{media=m[1]}
 if media==""{media=mediaAny.FindString(body)}
 if strings.HasPrefix(media,"//"){media="https:"+media}
 if !core.IsProviderHTTPMediaURL(media) {return core.ProviderMedia{},errors.New("media unavailable")}
 output:=core.ProviderMedia{URL:media,Referer:p.cfg.base+"/"}
 if strings.Contains(strings.ToLower(media),".m3u8"){
  return p.host.PrepareWebMedia(ctx,output,p.cfg.name)
 }
 return output,nil
}
