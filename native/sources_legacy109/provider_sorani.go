package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "time"

 "duanjuapp/native/core"
)

// Sorani 109 API public catalog and account-aware playback, ported to the
// upstream 110 Provider/Host interface. Never forges a login or bypasses DRM.
const sorani109Site="https://www.sorani.net"
const sorani109API="https://api.sorani.cc/sorani-cms/api"
const sorani109Line="anime_jp_m3u8"

var sorani109Categories=[]core.Category{
 {ID:"1",Name:"TV番劇"},
 {ID:"2",Name:"劇場動畫"},
 {ID:"5",Name:"特攝劇場"},
}
var sorani109Descriptor=core.SourceDescriptor{
 ID:"sorani",Name:"青空次元",BaseURL:sorani109Site,
 Description:"109 青空次元公開分類與需授權播放的 110 相容版；未完成線上驗證",
 NoHostRotation:true,
 ValidCategory:validSorani109Category,
}
type sorani109 struct{host core.Host}
var _ core.Provider=(*sorani109)(nil)
var _ core.Searcher=(*sorani109)(nil)
func init(){core.Register(sorani109Descriptor,func(h core.Host)core.Provider{return &sorani109{host:h}})}
func (p *sorani109) Descriptor()core.SourceDescriptor{return sorani109Descriptor}
func (p *sorani109) BaseURL()string{return sorani109Site}
func validSorani109Category(v string)bool{
 if v==""{return true}
 for _,c:=range sorani109Categories{if v==c.ID{return true}}
 return false
}
func sorani109ID(v string)bool{
 if len(v)==0||len(v)>64{return false}
 for _,r:=range v{if r<'0'||r>'9'{return false}}
 return v!="0"
}
func (p *sorani109) Categories(context.Context,bool)([]core.Category,error){
 return append([]core.Category(nil),sorani109Categories...),nil
}
func sorani109Rows(v any)([]any,bool){
 if rows,ok:=v.([]any);ok{return rows,true}
 m:=hjMap(v)
 for _,k:=range []string{"records","list","items","rows"}{
  if rows,ok:=m[k].([]any);ok{return rows,true}
 }
 return nil,false
}
func sorani109Drama(row map[string]any)(core.Drama,error){
 id:=hgField(row,"id","videoId")
 title:=hgField(row,"title","name","videoTitle")
 if !sorani109ID(id)||title==""{return core.Drama{},errors.New("Sorani invalid drama metadata")}
 cover:=hgField(row,"cover","coverLarge","coverSmall","coverThumb","backgroundImage")
 count:=0
 for _,key:=range []string{"episodeCount","totalEpisode","total"}{
  value,err:=strconv.Atoi(hjString(row[key]))
  if err==nil&&value>count&&value<=100000{count=value}
 }
 return core.Drama{
  ID:core.ProviderDramaID("sorani",id),Source:"sorani",SourceID:id,
  Title:title,Name:title,Cover:core.ResolveProviderURL(sorani109Site+"/",cover),
  Desc:hgField(row,"summary","description","intro","desc"),
  TotalEpisode:count,ChannelName:"青空次元",
 },nil
}
func (p *sorani109) fetch(ctx context.Context,route string,query url.Values,bearer string)(any,error){
 if !strings.HasPrefix(route,"/video")||strings.Contains(route,".."){
  return nil,errors.New("invalid Sorani route")
 }
 address:=sorani109API+route
 if len(query)>0{address+="?"+query.Encode()}
 headers:=map[string]string{
  "User-Agent":"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36",
  "Accept":"application/json, text/plain, */*",
  "Origin":sorani109Site,"Referer":sorani109Site+"/",
 }
 if bearer!=""{headers["Authorization"]="Bearer "+bearer}
 response,err:=p.host.Fetch(ctx,core.HTTPRequest{
  Label:"青空次元",Method:http.MethodGet,URL:address,Header:headers,
  Direct:true,Timeout:15*time.Second,Limit:8<<20,
 })
 if err!=nil{return nil,err}
 decoder:=json.NewDecoder(strings.NewReader(string(response.Body)))
 decoder.UseNumber()
 var value any
 if err:=decoder.Decode(&value);err!=nil{return nil,err}
 m:=hjMap(value)
 code:=hjString(m["code"])
 if code!=""&&code!="200"{return nil,fmt.Errorf("Sorani API rejected request: %s",code)}
 return hjPayload(value),nil
}
func (p *sorani109) catalog(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>100000||!validSorani109Category(category)||len(query)>256{
  return nil,false,errors.New("invalid Sorani catalog request")
 }
 values:=url.Values{"page":{strconv.Itoa(page)},"size":{"24"}}
 route:="/video"
 if query!=""{route="/video/search";values.Set("keyword",query)
 }else if category!=""{values.Set("categoryId",category)}
 payload,err:=p.fetch(ctx,route,values,"")
 if err!=nil{return nil,false,err}
 rows,valid:=sorani109Rows(payload)
 if !valid||len(rows)>500{return nil,false,errors.New("invalid Sorani catalog")}
 seen:=map[string]bool{}
 out:=make([]core.Drama,0,len(rows))
 for _,row:=range rows{
  drama,err:=sorani109Drama(hjMap(row))
  if err!=nil{return nil,false,err}
  if seen[drama.SourceID]{return nil,false,errors.New("duplicate Sorani drama")}
  seen[drama.SourceID]=true
  out=append(out,drama)
 }
 hasMore:=len(rows)>=24
 if total,ok:=hjInt(hjMap(payload)["total"],100000000);ok{hasMore=page*24<total}
 return out,hasMore,nil
}
func (p *sorani109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 return p.catalog(ctx,page,category,"")
}
func (p *sorani109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 return p.catalog(ctx,page,"",query)
}
func (p *sorani109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !sorani109ID(id){return core.Drama{},nil,errors.New("invalid Sorani drama ID")}
 payload,err:=p.fetch(ctx,"/video/"+url.PathEscape(id),nil,"")
 if err!=nil{return core.Drama{},nil,err}
 m:=hjMap(payload)
 drama,err:=sorani109Drama(m)
 if err!=nil{return core.Drama{},nil,err}
 if drama.SourceID!=id{return core.Drama{},nil,errors.New("Sorani detail ID mismatch")}
 episodes,valid:=m["episodes"].([]any)
 if !valid||len(episodes)>10000{return core.Drama{},nil,errors.New("Sorani episodes missing")}
 seen:=map[string]bool{}
 orderSeen:=map[int]bool{}
 var chapters []core.Chapter
 for i,item:=range episodes{
  ep:=hjMap(item)
  if enabled,ok:=ep["enabled"].(bool);ok&&!enabled{continue}
  epID:=hgField(ep,"episodeId","id")
  if !sorani109ID(epID)||seen[epID]{return core.Drama{},nil,errors.New("invalid or duplicate Sorani episode")}
  number:=i+1
  if n,ok:=hjInt(ep["episodeOrder"],100000);ok&&n>0{number=n}
  if orderSeen[number]{return core.Drama{},nil,errors.New("duplicate Sorani episode order")}
  seen[epID],orderSeen[number]=true,true
  title:=hgField(ep,"episodeLabel","title","displayTitle")
  if title==""{title=fmt.Sprintf("第 %d 集",number)}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("sorani",id,epID),Source:"sorani",
   Title:title,CurrentEpisode:core.RawEpisode(number),
   Referer:sorani109Site+"/",
  })
 }
 drama.TotalEpisode=len(chapters)
 core.SortProviderChapters(chapters)
 return drama,chapters,nil
}
func (p *sorani109) authorizedToken()(string,error){
 // Reuse only a genuine, previously authorized app session; no invented
 // token, guest identity or bypass of account permissions.
 raw,err:=p.host.LoadState("sorani-session.json")
 if err!=nil{return "",err}
 if len(raw)==0{return "",nil}
 var session struct{AccessToken string `json:"accessToken"`}
 if err:=json.Unmarshal(raw,&session);err!=nil{return "",err}
 if len(session.AccessToken)>4096||strings.ContainsAny(session.AccessToken,"\r\n\t "){
  return "",errors.New("invalid saved Sorani access token")
 }
 return session.AccessToken,nil
}
func (p *sorani109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!="sorani"||!sorani109ID(id){return core.ProviderMedia{},errors.New("invalid Sorani drama")}
 prefix:=core.ProviderChapterID("sorani",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid Sorani chapter")}
 episode:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !sorani109ID(episode){return core.ProviderMedia{},errors.New("invalid Sorani episode")}
 token,err:=p.authorizedToken()
 if err!=nil{return core.ProviderMedia{},err}
 payload,err:=p.fetch(ctx,"/video/episode/"+url.PathEscape(episode)+"/play",url.Values{
  "lineCode":{sorani109Line},
 },token)
 if err!=nil{return core.ProviderMedia{},err}
 m:=hjMap(payload)
 if allowed,ok:=m["canPlay"].(bool);ok&&!allowed{
  if token==""{return core.ProviderMedia{},core.SourceLoginError("青空次元需要登入授權才能播放")}
  return core.ProviderMedia{},errors.New("青空次元帳號沒有此分集的播放權限")
 }
 address:=hgField(m,"playUrl","url")
 if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("Sorani playback URL not available")}
 media:=core.ProviderMedia{URL:address,Referer:sorani109Site+"/"}
 if strings.Contains(strings.ToLower(address),".m3u8"){
  return p.host.PrepareWebMedia(ctx,media,"青空次元")
 }
 return media,nil
}
