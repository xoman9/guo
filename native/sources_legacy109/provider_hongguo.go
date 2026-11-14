package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "strconv"
 "strings"
 "duanjuapp/native/core"
)

// Reuses only the public 109 Hongguo web fallback routes.
// Does not claim to restore the separate App/CENC authenticated player.
const hgBase="https://hongguoduanju.com"
var hgCategories=[]core.Category{
 {ID:"real-drama",Name:"真人劇"},{ID:"comic-drama",Name:"漫劇"},
 {ID:"ai-drama",Name:"AI劇"},{ID:"comic",Name:"動漫"},
}
var hgDescriptor=core.SourceDescriptor{
 ID:"hongguo",Name:"紅果",BaseURL:hgBase,NoHostRotation:true,
 Description:"109 網頁備援路線移植 110；App/CENC 尚未恢復",
 ValidCategory:hgValidCategory,
}
type hongguo109 struct{host core.Host}
var _ core.Provider=(*hongguo109)(nil)
func init(){core.Register(hgDescriptor,func(h core.Host)core.Provider{return &hongguo109{host:h}})}
func (p *hongguo109) Descriptor()core.SourceDescriptor{return hgDescriptor}
func (p *hongguo109) BaseURL()string{return hgBase}
func hgNumericID(s string)bool{
 if len(s)==0||len(s)>24{return false}
 for _,r:=range s{if r<'0'||r>'9'{return false}}
 return true
}
func hgValidCategory(s string)bool{
 if s==""{return true}
 for _,c:=range hgCategories{if c.ID==s{return true}}
 return false
}
func (p *hongguo109) Categories(context.Context,bool)([]core.Category,error){
 return append([]core.Category(nil),hgCategories...),nil
}
func (p *hongguo109) fetch(ctx context.Context,path string)(string,error){
 if !strings.HasPrefix(path,"/category/")&&!strings.HasPrefix(path,"/detail?series_id=")&&!strings.HasPrefix(path,"/player/"){
  return "",errors.New("invalid Hongguo route")
 }
 response,err:=p.host.FetchText(ctx,core.TextRequest{URL:hgBase+path,Referer:hgBase+"/",UserAgent:ua})
 if err!=nil{return "",err}
 return response.Body,nil
}
func hgRouter(body string)map[string]any{
 marker:="_ROUTER_DATA"
 idx:=strings.Index(body,marker)
 if idx<0{return nil}
 rest:=body[idx+len(marker):]
 eq:=strings.Index(rest,"=")
 if eq<0||eq>16{return nil}
 decoder:=json.NewDecoder(strings.NewReader(rest[eq+1:]))
 decoder.UseNumber()
 var m map[string]any
 if decoder.Decode(&m)!=nil{return nil}
 return m
}
func hgLoader(data map[string]any,keys ...string)map[string]any{
 loader:=hjMap(data["loaderData"])
 for _,k:=range keys{if m:=hjMap(loader[k]);len(m)>0{return m}}
 for k,v:=range loader{
  for _,prefix:=range keys{
   if strings.HasSuffix(prefix,"_")&&strings.HasPrefix(k,prefix){
    if m:=hjMap(v);len(m)>0{return m}
   }
  }
 }
 return nil
}
func hgField(m map[string]any,keys ...string)string{
 for _,k:=range keys{if s:=hjString(m[k]);s!=""{return s}}
 return ""
}
func hgCards(v any)[]core.Drama{
 seen:=map[string]bool{}
 var out []core.Drama
 for _,item:=range hjRows(v){
  m:=hjMap(item);nested:=hjMap(m["video_data"])
  if len(nested)==0{nested=m}
  id:=hgField(nested,"series_id_str","series_id")
  if id==""{id=hgField(m,"series_id_str","series_id")}
  if !hgNumericID(id)||seen[id]{continue}
  title:=hgField(nested,"series_title","series_name","title")
  if title==""{title=hgField(m,"series_title","series_name","name")}
  if title==""{continue}
  cover:=hgField(nested,"series_cover","cover")
  if cover==""{cover=hgField(m,"series_cover")}
  episodeCount,_:=strconv.Atoi(hgField(nested,"episode_cnt"))
  seen[id]=true
  out=append(out,core.Drama{
   ID:core.ProviderDramaID("hongguo",id),Source:"hongguo",
   SourceID:id,Title:title,Name:title,ChannelName:"紅果",
   Cover:core.ResolveProviderURL(hgBase+"/",cover),TotalEpisode:episodeCount,
  })
 }
 return out
}
func (p *hongguo109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1||page>100000{return nil,false,errors.New("invalid page")}
 if category==""{category=hgCategories[0].ID}
 if !hgValidCategory(category){return nil,false,errors.New("invalid category")}
 path:=fmt.Sprintf("/category/%s?page=%d",url.PathEscape(category),page)
 body,err:=p.fetch(ctx,path)
 if err!=nil{return nil,false,err}
 data:=hgLoader(hgRouter(body),"category_page","category_")
 if len(data)==0||data["isSuccess"]==false{return nil,false,errors.New("no Hongguo category data")}
 pageInfo:=hjMap(data["pagination"])
 total,_:=strconv.Atoi(hgField(pageInfo,"totalPages"))
 return hgCards(data["recommendList"]),total>page,nil
}
func (p *hongguo109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !hgNumericID(id){return core.Drama{},nil,errors.New("invalid series ID")}
 body,err:=p.fetch(ctx,"/detail?series_id="+url.QueryEscape(id))
 if err!=nil{return core.Drama{},nil,err}
 row:=hjMap(hgLoader(hgRouter(body),"detail_page","detail_")["seriesDetail"])
 title:=hgField(row,"series_name","series_title","name")
 if title==""{return core.Drama{},nil,errors.New("detail title unavailable")}
 seen:=map[string]bool{}
 var chapters []core.Chapter
 for i,v:=range hjRows(row["vid_list"]){
  idVideo:=hjString(v)
  if !hgNumericID(idVideo)||seen[idVideo]{continue}
  seen[idVideo]=true
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("hongguo",id,idVideo),Source:"hongguo",
   Title:fmt.Sprintf("第%d集",i+1),CurrentEpisode:core.RawEpisode(i+1),
   PageURL:hgBase+"/player/"+url.PathEscape(id)+"/"+url.PathEscape(idVideo),
   Referer:hgBase+"/",
  })
 }
 if len(chapters)==0{return core.Drama{},nil,errors.New("no episodes")}
 return core.Drama{
  ID:core.ProviderDramaID("hongguo",id),Source:"hongguo",SourceID:id,
  Title:title,Name:title,TotalEpisode:len(chapters),ChannelName:"紅果",
 },chapters,nil
}
func (p *hongguo109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!="hongguo"||!hgNumericID(id){return core.ProviderMedia{},errors.New("invalid series")}
 prefix:=core.ProviderChapterID("hongguo",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid chapter")}
 vid:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !hgNumericID(vid){return core.ProviderMedia{},errors.New("invalid video ID")}
 path:="/player/"+url.PathEscape(id)+"/"+url.PathEscape(vid)
 body,err:=p.fetch(ctx,path)
 if err!=nil{return core.ProviderMedia{},err}
 data:=hgLoader(hgRouter(body),"player_page","player_")
 if hgField(data,"vid")!=vid||hgField(data,"series_id")!=id{
  return core.ProviderMedia{},errors.New("episode mismatch; refusing preview substitution")
 }
 info:=hjMap(data["video_player_info"])
 if info["drm"]!=nil||info["cenc"]!=nil||info["license_url"]!=nil{
  return core.ProviderMedia{},errors.New("encrypted player requires original authorization")
 }
 raw:=hgField(info,"main_url","play_url","url")
 if !core.IsProviderHTTPMediaURL(raw){
  return core.ProviderMedia{},errors.New("media URL not publicly present in requested episode")
 }
 media:=core.ProviderMedia{URL:raw,Referer:hgBase+"/"}
 if strings.Contains(strings.ToLower(raw),".m3u8"){return p.host.PrepareWebMedia(ctx,media,"紅果")}
 return media,nil
}
