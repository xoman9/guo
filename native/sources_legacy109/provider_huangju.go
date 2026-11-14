package legacy109

import (
 "context"
 "crypto/rand"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "sync"
 "time"

 "duanjuapp/native/core"
)

// 109 huangju routes (/auth/guest, /categories, /dramas, /play)
// ported to 110's exported Host.Fetch. No private Downloader dependency.
const hjSite = "https://huangju.net"
const hjAPI = "https://api.huangju.net"
const hjUserAgent = "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0 Mobile Safari/537.36"

type huangju109 struct {
 host core.Host
 mu sync.Mutex
 token string
 device string
}

var _ core.Provider = (*huangju109)(nil)
var _ core.Searcher = (*huangju109)(nil)

var huangju109Descriptor = core.SourceDescriptor{
 ID:"huangju",Name:"劇果",BaseURL:hjSite,DefaultOff:false,
 Description:"109 劇果 API 相容實驗版；110 Registry，未經線上驗證",
 NoHostRotation:true,
 ValidCategory:func(s string)bool{return validHuangju109Category(s)},
}
func init(){core.Register(huangju109Descriptor,func(h core.Host)core.Provider{return &huangju109{host:h}})}
func (p *huangju109) Descriptor()core.SourceDescriptor{return huangju109Descriptor}
func (p *huangju109) BaseURL()string{return hjSite}

func validHuangju109Category(s string)bool{
 if s==""||s=="@new"{return true}
 if len(s)>80 || strings.ContainsAny(s,"/:?#\\ %"){return false}
 for _,r:=range s {if r<32{return false}}
 return true
}
func huangju109ID(s string)bool{
 if s==""||len(s)>450||strings.TrimSpace(s)!=s||s=="."||s==".."{return false}
 for _,r:=range s {if r<32||strings.ContainsRune("/\\:?#%|\"<>",r){return false}}
 return true
}
func hjString(v any)string{
 switch v:=v.(type){
 case string:return strings.TrimSpace(v)
 case json.Number:return string(v)
 case float64:return strconv.FormatFloat(v,'f',-1,64)
 }
 return ""
}
func hjMap(v any)map[string]any{r,_:=v.(map[string]any);return r}
func hjRows(v any)[]any{r,_:=v.([]any);return r}
func hjInt(v any,max int)(int,bool){
 s:=hjString(v);n,err:=strconv.Atoi(s)
 return n,err==nil&&n>=0&&n<=max
}
func hjPayload(v any)any{
 // Old API response sometimes wraps the payload in a "data" envelope.
 if m:=hjMap(v);m!=nil{
  if x,ok:=m["data"];ok{return x}
 }
 return v
}
func (p *huangju109) deviceID()string{
 p.mu.Lock(); defer p.mu.Unlock()
 if p.device==""{
  var id [16]byte
  if _,err:=rand.Read(id[:]);err==nil{
   id[6]=id[6]&0x0f|0x40
   id[8]=id[8]&0x3f|0x80
   p.device=fmt.Sprintf("%x-%x-%x-%x-%x",id[:4],id[4:6],id[6:8],id[8:10],id[10:])
  }
 }
 return p.device
}
func (p *huangju109) call(ctx context.Context,method,route string,query url.Values,body []byte,token string)(core.HTTPResponse,error){
 if !strings.HasPrefix(route,"/") || strings.Contains(route,"..") {return core.HTTPResponse{},errors.New("invalid huangju API route")}
 addr:=hjAPI+route
 if len(query)>0{addr+="?"+query.Encode()}
 headers:=map[string]string{"User-Agent":hjUserAgent,"Accept":"application/json, text/plain, */*","Origin":hjSite,"Referer":hjSite+"/"}
 if len(body)>0{headers["Content-Type"]="application/json"}
 if token!=""{headers["Authorization"]="Bearer "+token}
 return p.host.Fetch(ctx,core.HTTPRequest{
  Label:"劇果",Method:method,URL:addr,Header:headers,Body:body,
  Limit:8<<20,Timeout:15*time.Second,Direct:true,
 })
}
func (p *huangju109) guest(ctx context.Context)(string,error){
 p.mu.Lock();token:=p.token;p.mu.Unlock()
 if token!=""{return token,nil}
 device:=p.deviceID()
 if device==""{return "",errors.New("cannot generate guest device ID")}
 data,_:=json.Marshal(map[string]string{"deviceId":device})
 response,err:=p.call(ctx,http.MethodPost,"/auth/guest",nil,data,"")
 if err!=nil{return "",err}
 var value any
 if err=json.Unmarshal(response.Body,&value);err!=nil{return "",err}
 row:=hjMap(hjPayload(value))
 token=hjString(row["token"])
 if token==""||len(token)>8192||strings.ContainsAny(token,"\r\n\t "){
  return "",errors.New("guest token unavailable")
 }
 p.mu.Lock();p.token=token;p.mu.Unlock()
 return token,nil
}
func (p *huangju109) get(ctx context.Context,route string,query url.Values)(any,core.HTTPResponse,error){
 token,err:=p.guest(ctx)
 if err!=nil{return nil,core.HTTPResponse{},err}
 response,err:=p.call(ctx,http.MethodGet,route,query,nil,token)
 if response.Status==http.StatusUnauthorized||response.Status==http.StatusForbidden{
  p.mu.Lock();if p.token==token{p.token=""};p.mu.Unlock()
  token,err=p.guest(ctx)
  if err!=nil{return nil,response,err}
  response,err=p.call(ctx,http.MethodGet,route,query,nil,token)
 }
 if err!=nil{return nil,response,err}
 var value any
 if err=json.Unmarshal(response.Body,&value);err!=nil{return nil,response,err}
 return hjPayload(value),response,nil
}
func (p *huangju109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 raw,_,err:=p.get(ctx,"/categories",nil)
 if err!=nil{return nil,err}
 rows:=hjRows(raw)
 if len(rows)>300{return nil,errors.New("too many categories")}
 categories:=[]core.Category{{ID:"@new",Name:"最新"}}
 seen:=map[string]bool{"@new":true}
 for _,row:=range rows{
  m:=hjMap(row);id,name:=hjString(m["slug"]),hjString(m["name"])
  if id==""||id=="hot"||name==""||!validHuangju109Category(id)||seen[id]{continue}
  seen[id]=true
  categories=append(categories,core.Category{ID:id,Name:name})
 }
 return categories,nil
}
func hjDrama(row map[string]any,expect string)(core.Drama,error){
 id,slug,title:=hjString(row["id"]),hjString(row["slug"]),hjString(row["title"])
 if !huangju109ID(id)||!huangju109ID(slug)||title==""{return core.Drama{},errors.New("invalid drama metadata")}
 sourceID:=slug+"-"+id
 if expect!=""&&expect!=sourceID{return core.Drama{},errors.New("drama ID mismatch")}
 cover:=core.ResolveProviderURL(hjSite+"/",hjString(row["coverUrl"]))
 total,_:=hjInt(row["totalEpisodes"],100000)
 desc:=hjString(row["description"])
 return core.Drama{
  ID:core.ProviderDramaID("huangju",sourceID),Source:"huangju",SourceID:sourceID,
  Title:title,Name:title,Cover:cover,CoverURL:cover,
  TotalEpisode:total,EpisodeCount:total,Desc:desc,Intro:desc,ChannelName:"劇果",
 },nil
}
func (p *huangju109) catalog(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>1000000||len(query)>1024||!validHuangju109Category(category){return nil,false,errors.New("invalid catalog request")}
 values:=url.Values{"page":{strconv.Itoa(page)}}
 switch {
 case query!="":values.Set("q",query)
 case category=="@new":values.Set("sort","new")
 case category=="":values.Set("sort","hot")
 default:values.Set("category",category)
 }
 raw,_,err:=p.get(ctx,"/dramas",values)
 if err!=nil{return nil,false,err}
 meta:=hjMap(raw);rows:=hjRows(meta["items"])
 if len(rows)>500{return nil,false,errors.New("too many dramas")}
 var items []core.Drama
 seen:=map[string]bool{}
 for _,v:=range rows{
  item,err:=hjDrama(hjMap(v),"")
  if err!=nil{return nil,false,err}
  if seen[item.ID]{return nil,false,errors.New("duplicate drama")}
  seen[item.ID]=true;items=append(items,item)
 }
 pageSize:=20
 if v,ok:=hjInt(meta["pageSize"],500);ok&&v>0{pageSize=v}
 hasMore:=len(rows)>=pageSize
 if total,ok:=hjInt(meta["total"],100000000);ok{
  hasMore=page<(total+pageSize-1)/pageSize
 }
 return items,hasMore,nil
}
func (p *huangju109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 return p.catalog(ctx,page,category,"")
}
func (p *huangju109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 return p.catalog(ctx,page,"",query)
}
func (p *huangju109) Detail(ctx context.Context,sourceID string)(core.Drama,[]core.Chapter,error){
 if !huangju109ID(sourceID){return core.Drama{},nil,errors.New("invalid drama ID")}
 raw,_,err:=p.get(ctx,"/dramas/"+url.PathEscape(sourceID),nil)
 if err!=nil{return core.Drama{},nil,err}
 row:=hjMap(raw)
 drama,err:=hjDrama(row,sourceID)
 if err!=nil{return core.Drama{},nil,err}
 episodes:=hjRows(row["episodes"])
 if len(episodes)>10000{return core.Drama{},nil,errors.New("too many episodes")}
 seen:=map[string]bool{}
 var chapters []core.Chapter
 for i,raw:=range episodes{
  entry:=hjMap(raw)
  if allowed,ok:=entry["playable"].(bool);ok&&!allowed{continue}
  id:=hjString(entry["id"])
  if !huangju109ID(id)||seen[id]{return core.Drama{},nil,errors.New("invalid episode ID")}
  seen[id]=true
  n:=i+1
  if value,ok:=hjInt(entry["epNo"],100000);ok&&value>0{n=value}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("huangju",sourceID,id),Source:"huangju",
   Title:fmt.Sprintf("第 %d 集",n),CurrentEpisode:core.RawEpisode(n),
   PageURL:hjSite+"/dramas/"+url.PathEscape(sourceID),Referer:hjSite+"/",
  })
 }
 drama.TotalEpisode=len(chapters)
 return drama,chapters,nil
}
func hjCredentials(headers map[string][]string,mediaURL string)(*core.MediaCredentials,error){
 parsed,err:=url.Parse(mediaURL)
 if err!=nil||parsed.Scheme!="https"||parsed.Host==""{return nil,errors.New("invalid media URL")}
 required:=[]string{"CloudFront-Policy","CloudFront-Signature","CloudFront-Key-Pair-Id"}
 cookies:=map[string]string{}
 for name,values:=range headers{
  if !strings.EqualFold(name,"Set-Cookie"){continue}
  for _,item:=range values{
   cookie:=strings.SplitN(item,";",2)[0]
   name,value,ok:=strings.Cut(strings.TrimSpace(cookie),"=")
   if !ok||value==""{continue}
   for _,key:=range required{if name==key{cookies[key]=value}}
  }
 }
 var pairs []string
 for _,key:=range required{
  v:=cookies[key]
  if v==""||len(v)>8192||strings.ContainsAny(v,";\r\n"){return nil,errors.New("signed media cookie missing")}
  pairs=append(pairs,key+"="+v)
 }
 return &core.MediaCredentials{
  Cookie:strings.Join(pairs,"; "),Origin:core.ProviderMediaOrigin(parsed),
  Referer:hjSite+"/",UserAgent:hjUserAgent,
 },nil
}
func (p *huangju109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="huangju"||!huangju109ID(id){return core.ProviderMedia{},errors.New("invalid drama")}
 prefix:=core.ProviderChapterID("huangju",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){
  return core.ProviderMedia{},errors.New("invalid episode")
 }
 episode:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !huangju109ID(episode){return core.ProviderMedia{},errors.New("invalid episode ID")}
 raw,response,err:=p.get(ctx,"/play/"+url.PathEscape(episode),nil)
 if err!=nil{return core.ProviderMedia{},err}
 row:=hjMap(raw)
 address:=hjString(row["url"])
 credentials,err:=hjCredentials(response.Header,address)
 if err!=nil{return core.ProviderMedia{},err}
 media:=core.ProviderMedia{URL:address,Referer:hjSite+"/",Credentials:credentials}
 if strings.Contains(strings.ToLower(address),".m3u8"){
  media.Playlist,media.URL,err=p.host.FetchPlaylist(ctx,address,media.Referer,credentials)
  if err!=nil{return core.ProviderMedia{},err}
 }
 return media,nil
}
