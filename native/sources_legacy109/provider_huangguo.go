package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "regexp"
 "strconv"
 "strings"
 "time"
 "duanjuapp/native/core"
)

const hgAIBase="https://huangguoai.com"
const hgVideoBase="https://huangguo.video"
var hgAISlugs=[]string{"recommend","newest","ai-duanju","ai-manju","ai-huanlian","ai-mogai","ranks/hot"}
var hgVideoCardRX=regexp.MustCompile(`(?is)<article\b[^>]*class=["'][^"']*\bvideo-card\b[^"']*["'][^>]*>.*?</article>`)
var hgVideoLinkRX=regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*/(?:series|video)/([a-zA-Z0-9_-]+)[^"']*)["']`)
var hgVideoImgRX=regexp.MustCompile(`(?is)<img\b[^>]*(?:data-src|data-original|src)=["']([^"']+)["']`)
var hgDataHLSRX=regexp.MustCompile(`(?is)data-hls=["']([^"']+)["']`)
var hgTitleRX=regexp.MustCompile(`(?is)<h1\b[^>]*>(.*?)</h1>`)
var hgAttrTitleRX=regexp.MustCompile(`(?is)\b(?:alt|title)=["']([^"']+)["']`)
var hgIDRX=regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)
type huangguo109 struct{host core.Host; ai bool}
func (p *huangguo109) source()string{if p.ai{return "huangguoai"};return "huangguo-video"}
func (p *huangguo109) base()string{if p.ai{return hgAIBase};return hgVideoBase}
func (p *huangguo109) Descriptor()core.SourceDescriptor{
 name:="黃果視頻 +";if p.ai{name="黃果AI +"}
 return core.SourceDescriptor{ID:p.source(),Name:name,
  Description:"109 黃果網頁與公開 API 相容版；未經線上驗證",
  BaseURL:p.base(),DefaultOff:true,NoHostRotation:true,
  ValidCategory:p.validCategory,
 }
}
func (p *huangguo109) BaseURL()string{return p.base()}
func (p *huangguo109) validCategory(id string)bool{
 if id==""{return true}
 if p.ai{for _,v:=range hgAISlugs{if id==v{return true}};return false}
 return id=="series"||id=="video"
}
func init(){
 for _,value:=range []bool{true,false}{
  ai:=value
  spec:=&huangguo109{ai:ai}
  core.Register(spec.Descriptor(),func(h core.Host)core.Provider{return &huangguo109{host:h,ai:ai}})
 }
}
var _ core.Provider=(*huangguo109)(nil)
var _ core.Searcher=(*huangguo109)(nil)
func (p *huangguo109) Categories(context.Context,bool)([]core.Category,error){
 var out []core.Category
 if p.ai{for _,id:=range hgAISlugs{out=append(out,core.Category{ID:id,Name:id})}
 }else{out=[]core.Category{{ID:"series",Name:"系列"},{ID:"video",Name:"單片"}}}
 return out,nil
}
func (p *huangguo109) get(ctx context.Context,path string)(string,error){
 if !strings.HasPrefix(path,"/"){return "",errors.New("invalid route")}
 r,err:=p.host.FetchText(ctx,core.TextRequest{URL:p.base()+path,Referer:p.base()+"/",UserAgent:ua})
 if err!=nil{return "",err};return r.Body,nil
}
func (p *huangguo109) aiAPI(ctx context.Context,path string)(any,error){
 if !p.ai||!strings.HasPrefix(path,"/api/"){return nil,errors.New("invalid Huangguo AI API route")}
 envelope,_:=json.Marshal(map[string]string{"method":"GET","path":path,"body":"","content_type":""})
 response,err:=p.host.Fetch(ctx,core.HTTPRequest{
  Label:"黃果AI",Method:http.MethodPost,URL:hgAIBase+"/api/edge/request",
  Body:envelope,Header:map[string]string{"Content-Type":"application/json","Origin":hgAIBase,"Referer":hgAIBase+"/"},
  Timeout:15*time.Second,Limit:8<<20,Direct:true,
 })
 if err!=nil{return nil,err}
 var reply struct{Status int `json:"status"`;Msg string `json:"msg"`;Data json.RawMessage `json:"data"`}
 if err:=json.Unmarshal(response.Body,&reply);err!=nil{return nil,err}
 if reply.Status!=1{return nil,fmt.Errorf("Huangguo AI API returned failure: %s",reply.Msg)}
 var result any
 if err:=json.Unmarshal(reply.Data,&result);err!=nil{return nil,err}
 return result,nil
}
func hgAIDrama(row map[string]any)(core.Drama,error){
 id:=hgField(row,"id","video_id")
 name:=hgField(row,"title","name")
 if !hgIDRX.MatchString(id)||name==""{return core.Drama{},errors.New("invalid Huangguo AI drama")}
 image:=hgField(row,"cover","cover_url","poster","image")
 episodes,_:=hjInt(row["episode_count"],100000)
 return core.Drama{ID:core.ProviderDramaID("huangguoai",id),Source:"huangguoai",
  SourceID:id,Title:name,Name:name,Cover:core.ResolveProviderURL(hgAIBase+"/",image),
  TotalEpisode:episodes,Desc:hgField(row,"description","intro")},nil
}
func hgVideoItems(raw,origin string)[]core.Drama{
 var out []core.Drama;seen:=map[string]bool{}
 for _,block:=range hgVideoCardRX.FindAllString(raw,-1){
  m:=hgVideoLinkRX.FindStringSubmatch(block);if len(m)<3{continue}
  u,err:=url.Parse(m[1]);if err!=nil{continue}
  segments:=strings.Split(strings.Trim(u.Path,"/"),"/")
  id:=""
  for i:=0;i+1<len(segments);i++{
   if segments[i]=="series"||segments[i]=="video"{id=segments[i]+"/"+segments[i+1];break}
  }
  if id==""||seen[id]{continue}
  title:=""
  if t:=hgAttrTitleRX.FindStringSubmatch(block);len(t)>1{title=t[1]}
  if title==""{title=core.CleanText(block)}
  if title==""{continue}
  cover:=""
  if img:=hgVideoImgRX.FindStringSubmatch(block);len(img)>1{cover=core.ResolveProviderURL(origin,img[1])}
  seen[id]=true
  out=append(out,core.Drama{ID:core.ProviderDramaID("huangguo-video",id),Source:"huangguo-video",
   SourceID:id,Title:title,Name:title,Cover:cover})
 }
 return out
}
func (p *huangguo109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1||page>100000||!p.validCategory(category){return nil,false,errors.New("invalid Huangguo catalog")}
 if p.ai{
  if category==""{category="recommend"}
  path:="/"+category+"?page="+strconv.Itoa(page)
  body,err:=p.get(ctx,path);if err!=nil{return nil,false,err}
  pattern:=regexp.MustCompile(`(?is)href=["']/detail/([a-zA-Z0-9_-]{1,128})["'][^>]*>(.*?)</a>`)
  var out []core.Drama;seen:=map[string]bool{}
  for _,m:=range pattern.FindAllStringSubmatch(body,-1){
   if seen[m[1]]{continue}
   name:=core.CleanText(m[2]);if name==""{continue}
   seen[m[1]]=true
   out=append(out,core.Drama{ID:core.ProviderDramaID(p.source(),m[1]),Source:p.source(),SourceID:m[1],Title:name,Name:name})
  }
  return out,false,nil
 }
 if category==""{category="series"}
 path:="/"+category+"?page="+strconv.Itoa(page)
 body,err:=p.get(ctx,path);if err!=nil{return nil,false,err}
 return hgVideoItems(body,p.base()+path),false,nil
}
func (p *huangguo109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 if len(query)>256||page<1{return nil,false,errors.New("invalid search")}
 if p.ai{
  value,err:=p.aiAPI(ctx,"/api/search?"+url.Values{"q":{query},"page":{strconv.Itoa(page)},"size":{"20"}}.Encode())
  if err!=nil{return nil,false,err}
  m:=hjMap(value);var out []core.Drama
  for _,v:=range hjRows(m["items"]){
   item,err:=hgAIDrama(hjMap(v));if err==nil{out=append(out,item)}
  }
  pages,_:=hjInt(hjMap(m["pagination"])["pages"],100000)
  return out,page<pages,nil
 }
 body,err:=p.get(ctx,"/search?q="+url.QueryEscape(query)+"&page="+strconv.Itoa(page))
 if err!=nil{return nil,false,err}
 return hgVideoItems(body,p.base()+"/"),false,nil
}
func (p *huangguo109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if p.ai{
  if !hgIDRX.MatchString(id){return core.Drama{},nil,errors.New("invalid AI drama")}
  value,err:=p.aiAPI(ctx,"/api/videos/"+url.PathEscape(id))
  if err!=nil{return core.Drama{},nil,err}
  m:=hjMap(value);drama,err:=hgAIDrama(m)
  if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("AI detail mismatch")}
  var chapters []core.Chapter
  for _,v:=range hjRows(m["episodes"]){
   row:=hjMap(v);n,_:=hjInt(row["ep_num"],100000)
   if n<1{continue}
   title:=hjString(row["title"]);if title==""{title=fmt.Sprintf("第%d集",n)}
   chapters=append(chapters,core.Chapter{ID:core.ProviderChapterID(p.source(),id,strconv.Itoa(n)),
    Source:p.source(),Title:title,CurrentEpisode:core.RawEpisode(n),Referer:p.base()+"/"})
  }
  if len(chapters)==0{chapters=append(chapters,core.Chapter{ID:core.ProviderChapterID(p.source(),id,"1"),Source:p.source(),Title:"播放",CurrentEpisode:core.RawEpisode(1)})}
  core.SortProviderChapters(chapters);drama.TotalEpisode=len(chapters)
  return drama,chapters,nil
 }
 parts:=strings.Split(id,"/")
 if len(parts)!=2||(parts[0]!="series"&&parts[0]!="video")||!hgIDRX.MatchString(parts[1]){return core.Drama{},nil,errors.New("invalid video ID")}
 body,err:=p.get(ctx,"/"+id);if err!=nil{return core.Drama{},nil,err}
 title:=""
 if m:=hgTitleRX.FindStringSubmatch(body);len(m)>1{title=core.CleanText(m[1])}
 if title==""{return core.Drama{},nil,errors.New("video title unavailable")}
 var chapters []core.Chapter
 if parts[0]=="video"{
  chapters=[]core.Chapter{{ID:core.ProviderChapterID(p.source(),id,"1"),Source:p.source(),Title:"播放",CurrentEpisode:core.RawEpisode(1),PageURL:p.base()+"/"+id}}
 }else{
  seen:=map[string]bool{}
  for _,m:=range hgVideoLinkRX.FindAllStringSubmatch(body,-1){
   if len(m)<3||seen[m[2]]{continue}
   link:=core.ResolveProviderURL(p.base()+"/"+id,m[1])
   if !strings.HasPrefix(link,p.base()+"/video/"){continue}
   seen[m[2]]=true
   chapters=append(chapters,core.Chapter{ID:core.ProviderChapterID(p.source(),id,m[2]),Source:p.source(),
    Title:fmt.Sprintf("第%d集",len(chapters)+1),CurrentEpisode:core.RawEpisode(len(chapters)+1),PageURL:link})
  }
 }
 return core.Drama{ID:core.ProviderDramaID(p.source(),id),Source:p.source(),SourceID:id,Title:title,Name:title,TotalEpisode:len(chapters)},chapters,nil
}
func (p *huangguo109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!=p.source(){return core.ProviderMedia{},errors.New("wrong Huangguo source")}
 prefix:=core.ProviderChapterID(p.source(),id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("wrong Huangguo chapter")}
 episode:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if p.ai{
  n,err:=strconv.Atoi(episode);if err!=nil||n<1{return core.ProviderMedia{},errors.New("invalid AI episode")}
  value,err:=p.aiAPI(ctx,fmt.Sprintf("/api/videos/%s/play?ep=%d&page_texts=0",url.PathEscape(id),n))
  if err!=nil{return core.ProviderMedia{},err}
  address:=core.NormalizeProviderMediaURL(p.base()+"/",hjString(hjMap(value)["video_url"]))
  if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("no AI publicly issued media URL")}
  return p.host.PrepareWebMedia(ctx,core.ProviderMedia{URL:address,Referer:p.base()+"/"},"黃果AI")
 }
 page:=task.Chapter.PageURL
 if !strings.HasPrefix(page,p.base()+"/video/"){return core.ProviderMedia{},errors.New("invalid video page")}
 body,err:=p.get(ctx,strings.TrimPrefix(page,p.base()))
 if err!=nil{return core.ProviderMedia{},err}
 hls:=hgDataHLSRX.FindStringSubmatch(body)
 if len(hls)<2{return core.ProviderMedia{},errors.New("video page does not expose direct HLS")}
 address:=core.NormalizeProviderMediaURL(page,hls[1])
 if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("invalid video media URL")}
 return p.host.PrepareWebMedia(ctx,core.ProviderMedia{URL:address,Referer:page},"黃果視頻")
}
