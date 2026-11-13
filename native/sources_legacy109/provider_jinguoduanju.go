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
const jgdjBase="https://jinguoduanju.app"
const jgdjAPI=jgdjBase+"/api/v1"
type jgdj109 struct{
 host core.Host
 mu sync.Mutex
 device,token string
}
var jgdjDesc=core.SourceDescriptor{
 ID:"jinguoduanju",Name:"禁果短劇",BaseURL:jgdjBase,NoHostRotation:true,
 Description:"109 合法裝置會話及對應分集的 API 播放移植至 110；尚未線上驗證",
 ValidCategory:jgdjValidCategory,
}
var _ core.Provider=(*jgdj109)(nil)
var _ core.Searcher=(*jgdj109)(nil)
func init(){core.Register(jgdjDesc,func(h core.Host)core.Provider{return &jgdj109{host:h}})}
func (p *jgdj109) Descriptor()core.SourceDescriptor{return jgdjDesc}
func (p *jgdj109) BaseURL()string{return jgdjBase}
func jgdjValidCategory(s string)bool{
 if s==""||s=="home"||s=="heat"{return true}
 kind,id,ok:=strings.Cut(s,":")
 return ok&&(kind=="nav"||kind=="tag")&&dsd109Num(id)
}
func jgdjUUID()string{
 var b [16]byte
 if _,err:=rand.Read(b[:]);err!=nil{return ""}
 b[6]=b[6]&0x0f|0x40;b[8]=b[8]&0x3f|0x80
 return fmt.Sprintf("%x-%x-%x-%x-%x",b[:4],b[4:6],b[6:8],b[8:10],b[10:])
}
func (p *jgdj109) request(ctx context.Context,method,path string,query url.Values,body []byte,token string)(any,error){
 if !strings.HasPrefix(path,"/")||strings.Contains(path,".."){return nil,errors.New("invalid jinguoduanju route")}
 address:=jgdjAPI+path
 if len(query)>0{address+="?"+query.Encode()}
 headers:=map[string]string{"Accept":"application/json","Origin":jgdjBase,"Referer":jgdjBase+"/","User-Agent":ua}
 if len(body)>0{headers["Content-Type"]="application/json"}
 if token!=""{headers["Authorization"]="Bearer "+token}
 response,err:=p.host.Fetch(ctx,core.HTTPRequest{
  Label:"禁果短劇",Method:method,URL:address,Header:headers,
  Body:body,Timeout:15*time.Second,Limit:8<<20,Direct:true,
 })
 if err!=nil{return nil,err}
 decoder:=json.NewDecoder(strings.NewReader(string(response.Body)));decoder.UseNumber()
 var value any
 if err:=decoder.Decode(&value);err!=nil{return nil,err}
 return value,nil
}
func (p *jgdj109) guest(ctx context.Context)(string,error){
 p.mu.Lock()
 token:=p.token
 device:=p.device
 if device==""{device=jgdjUUID();p.device=device}
 p.mu.Unlock()
 if token!=""{return token,nil}
 if device==""{return "",errors.New("device ID unavailable")}
 body,_:=json.Marshal(map[string]string{"deviceId":"h5:"+device,"channelCode":"","attributionToken":""})
 value,err:=p.request(ctx,http.MethodPost,"/auth/device",nil,body,"")
 if err!=nil{return "",err}
 token=hjString(hjMap(value)["token"])
 if token==""||len(token)>8192||strings.ContainsAny(token,"\r\n\t "){return "",errors.New("device authorization unavailable")}
 p.mu.Lock();p.token=token;p.mu.Unlock()
 return token,nil
}
func (p *jgdj109) get(ctx context.Context,path string,q url.Values)(any,error){
 token,err:=p.guest(ctx)
 if err!=nil{return nil,err}
 return p.request(ctx,http.MethodGet,path,q,nil,token)
}
func (p *jgdj109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 result:=[]core.Category{{ID:"home",Name:"推薦"},{ID:"heat",Name:"熱度"}}
 value,err:=p.get(ctx,"/top-navigation",nil)
 if err!=nil{return result,err}
 seen:=map[string]bool{}
 for _,v:=range hjRows(value){
  row:=hjMap(v);id:=hjString(row["id"]);name:=hjString(row["name"])
  if !dsd109Num(id)||name==""||seen[id]{continue}
  seen[id]=true
  result=append(result,core.Category{ID:"nav:"+id,Name:name})
  if len(result)>202{return nil,errors.New("too many categories")}
 }
 return result,nil
}
func jgdjDrama(m map[string]any)(core.Drama,error){
 id,title:=hjString(m["id"]),hjString(m["title"])
 if !dsd109Num(id)||title==""{return core.Drama{},errors.New("invalid Jinguoduanju drama")}
 count,_:=hjInt(m["episodeCount"],100000)
 cover:=core.ResolveProviderURL(jgdjBase+"/",hgField(m,"coverUrl","cover"))
 return core.Drama{
  ID:core.ProviderDramaID("jinguoduanju",id),Source:"jinguoduanju",
  SourceID:id,Title:title,Name:title,Cover:cover,
  TotalEpisode:count,Desc:hgField(m,"description","subtitle"),ChannelName:"禁果短劇",
 },nil
}
func jgdjValues(category,query string)(url.Values,error){
 if !jgdjValidCategory(category)||len(query)>256{return nil,errors.New("invalid category or search")}
 q:=url.Values{"sort":{"latest"}}
 if query!=""{q.Set("q",query);return q,nil}
 switch category{
 case "home":q.Set("displayArea","HOME_FEED")
 case "heat":q.Set("sort","heat")
 default:
  kind,id,ok:=strings.Cut(category,":")
  if ok{if kind=="nav"{q.Set("navigationId",id)}else{q.Set("tags",id)}}
 }
 return q,nil
}
func (p *jgdj109) listing(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>100000{return nil,false,errors.New("invalid page")}
 q,err:=jgdjValues(category,query);if err!=nil{return nil,false,err}
 data,err:=p.get(ctx,"/dramas",q)
 if err!=nil{return nil,false,err}
 rows:=hjRows(data)
 if rows==nil{
  m:=hjMap(data)
  for _,name:=range []string{"dramas","items","list","records"}{
   if result:=hjRows(m[name]);result!=nil{rows=result;break}
  }
 }
 if rows==nil||len(rows)>20000{return nil,false,errors.New("invalid drama list")}
 start:=(page-1)*24
 if start>=len(rows){return nil,false,nil}
 end:=start+24;if end>len(rows){end=len(rows)}
 var out []core.Drama
 for _,v:=range rows[start:end]{
  item,err:=jgdjDrama(hjMap(v));if err==nil{out=append(out,item)}
 }
 return out,end<len(rows),nil
}
func (p *jgdj109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 return p.listing(ctx,page,category,"")
}
func (p *jgdj109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query);if query==""{return nil,false,nil}
 return p.listing(ctx,page,"",query)
}
func (p *jgdj109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !dsd109Num(id){return core.Drama{},nil,errors.New("invalid drama")}
 data,err:=p.get(ctx,"/dramas/"+url.PathEscape(id),nil)
 if err!=nil{return core.Drama{},nil,err}
 obj:=hjMap(data);row:=hjMap(obj["drama"]);if row==nil{row=obj}
 drama,err:=jgdjDrama(row)
 if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("drama mismatch")}
 seen:=map[string]bool{};numSeen:=map[int]bool{}
 var episodes []core.Chapter
 for _,v:=range hjRows(obj["episodes"]){
  m:=hjMap(v);ep:=hjString(m["id"])
  n,ok:=hjInt(m["number"],100000)
  if !dsd109Num(ep)||!ok||n<1||seen[ep]||numSeen[n]{return core.Drama{},nil,errors.New("invalid or duplicate episode")}
  seen[ep],numSeen[n]=true,true
  title:=hjString(m["title"]);if title==""{title=fmt.Sprintf("第%d集",n)}
  locked,_:=m["locked"].(bool)
  episodes=append(episodes,core.Chapter{
   ID:core.ProviderChapterID("jinguoduanju",id,ep),
   Source:"jinguoduanju",Title:title,VIP:locked,
   CurrentEpisode:core.RawEpisode(n),Referer:jgdjBase+"/",
  })
 }
 core.SortProviderChapters(episodes)
 drama.TotalEpisode=len(episodes)
 return drama,episodes,nil
}
func (p *jgdj109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!="jinguoduanju"||!dsd109Num(id){return core.ProviderMedia{},errors.New("invalid drama")}
 prefix:=core.ProviderChapterID("jinguoduanju",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid episode")}
 episode:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !dsd109Num(episode){return core.ProviderMedia{},errors.New("invalid episode ID")}
 data,err:=p.get(ctx,"/episodes/"+url.PathEscape(episode)+"/play",nil)
 if err!=nil{return core.ProviderMedia{},err}
 m:=hjMap(data)
 if v:=hjString(m["episodeId"]);v!=""&&v!=episode{return core.ProviderMedia{},errors.New("wrong episode returned")}
 if allowed,ok:=hjMap(m["allowance"])["allowed"].(bool);ok&&!allowed{return core.ProviderMedia{},errors.New("episode requires authorized purchase")}
 address:=core.NormalizeProviderMediaURL(jgdjBase+"/",hjString(m["videoUrl"]))
 if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("no playable media URL")}
 return p.host.PrepareWebMedia(ctx,core.ProviderMedia{URL:address,Referer:jgdjBase+"/"},"禁果短劇")
}
