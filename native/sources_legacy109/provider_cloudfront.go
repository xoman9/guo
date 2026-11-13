package legacy109
import (
 "context"
 "crypto/rand"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "regexp"
 "strconv"
 "strings"
 "sync"
 "time"
 "duanjuapp/native/core"
)
const cf109Base="https://dr6skssi3nxbk.cloudfront.net"
const cf109Site="https://d2pypzndaqisk.cloudfront.net"
const cf109CDN="https://sjljsla.lkkwip.cn"
type cfProtocol109 struct{InterfaceKey string `json:"interfaceKey"`;ParamKey string `json:"paramKey"`;ParamIV string `json:"paramIV"`}
type cf109Provider struct{host core.Host;mu sync.Mutex;device,token string;protocol cfProtocol109}
var cf109Desc=core.SourceDescriptor{ID:"cloudfront",Name:"黃果舊版",BaseURL:cf109Site,NoHostRotation:true,Description:"109 官方舊版 HTTP API 和合法訪客 HLS 協定相容實驗版",ValidCategory:func(s string)bool{return s==""||jgID(s)}}
var _ core.Provider=(*cf109Provider)(nil)
var _ core.Searcher=(*cf109Provider)(nil)
func init(){core.Register(cf109Desc,func(h core.Host)core.Provider{return &cf109Provider{host:h}})}
func (p *cf109Provider) Descriptor()core.SourceDescriptor{return cf109Desc}
func (p *cf109Provider) BaseURL()string{return cf109Site}
func cfValidProtocol109(p cfProtocol109)bool{return p.InterfaceKey!=""&&len(p.InterfaceKey)<=512&&len(p.ParamKey)==16&&len(p.ParamIV)==16}
func cfExtractProtocol109(script string)(cfProtocol109,error){
 read:=func(k string)string{
  exp:=regexp.MustCompile("(?i)[\\\"']"+regexp.QuoteMeta(k)+"[\\\"']\\s*[:,]\\s*[\\\"']([^\\\"'\\\\\\r\\n]+)[\\\"']")
  m:=exp.FindStringSubmatch(script);if len(m)>1{return m[1]};return ""
 }
 p:=cfProtocol109{read("interfaceKey"),read("parameterKey"),read("parameterIv")}
 if !cfValidProtocol109(p){return p,errors.New("109 public API protocol not found")}
 return p,nil
}
func (p *cf109Provider) authenticate(ctx context.Context)(cfProtocol109,string,error){
 p.mu.Lock()
 if p.token!=""&&cfValidProtocol109(p.protocol){protocol,token:=p.protocol,p.token;p.mu.Unlock();return protocol,token,nil}
 p.mu.Unlock()
 var state struct{DeviceID string `json:"deviceId"`;Token string `json:"token"`;Protocol cfProtocol109 `json:"protocol"`}
 raw,err:=p.host.LoadState("huangguo-session.json")
 if err==nil&&len(raw)>0{_ =json.Unmarshal(raw,&state)}
 if !cfValidProtocol109(state.Protocol){
  // Only the actual site-provided protocol, never a fabricated API key.
  page,err:=p.host.FetchText(ctx,core.TextRequest{URL:cf109Site+"/",Referer:cf109Site+"/",UserAgent:ua})
  if err!=nil{return cfProtocol109{},"",err}
  state.Protocol,err=cfExtractProtocol109(page.Body)
  if err!=nil{return cfProtocol109{},"",err}
 }
 if state.DeviceID==""{
  var seed [8]byte
  if _,err:=rand.Read(seed[:]);err!=nil{return cfProtocol109{},"",err}
  state.DeviceID=strings.ToUpper(fmt.Sprintf("%x",seed[:]))+strconv.FormatInt(time.Now().UnixMilli(),10)
 }
 p.mu.Lock();p.device=state.DeviceID;p.protocol=state.Protocol;p.mu.Unlock()
 if state.Token==""{
  value,err:=p.api(ctx,http.MethodPost,"/api/app/mine/login/h5",
   map[string]any{"devID":state.DeviceID,"sysType":"ios","isAppStore":false},state.Protocol,"")
  if err!=nil{return cfProtocol109{},"",err}
  state.Token=hjString(hjMap(value)["token"])
  if state.Token==""||len(state.Token)>4096||strings.ContainsAny(state.Token,"\r\n\t "){return cfProtocol109{},"",errors.New("CloudFront guest login unavailable")}
  saved,_:=json.Marshal(state)
  _=p.host.SaveState("huangguo-session.json",saved)
 }
 p.mu.Lock();p.token=state.Token;p.mu.Unlock()
 return state.Protocol,state.Token,nil
}
func (p *cf109Provider) api(ctx context.Context,method,path string,params any,protocol cfProtocol109,token string)(any,error){
 if !strings.HasPrefix(path,"/api/app/")||strings.Contains(path,"..")||!cfValidProtocol109(protocol){return nil,errors.New("invalid legacy API request")}
 var body []byte
 if params!=nil{
  encrypted,err:=legacy109Encrypt(params,[]byte(protocol.ParamKey),[]byte(protocol.ParamIV))
  if err!=nil{return nil,err}
  if method==http.MethodGet{path+="?data="+url.QueryEscape(encrypted)}else{body,_=json.Marshal(map[string]string{"data":encrypted})}
 }
 p.mu.Lock();device:=p.device;p.mu.Unlock()
 headers:=map[string]string{
  "Accept":"application/json","Origin":cf109Site,"Referer":cf109Site+"/","User-Agent":ua,
  "X-User-Agent":fmt.Sprintf("BuildID=com.abc.Butterfly;SysType=ios;DevID=%s;Ver=1.0.0;DevType=iPhone;DeviceBrand=APPLE;DeviceModel=iPhone;SystemName=iOS;SystemVersion=18.7;Terminal=1;IsH5=1;Sid=00000000000000000000000000000000",device),
 }
 if token!=""{headers["Authorization"]=token}
 if len(body)>0{headers["Content-Type"]="application/json"}
 response,err:=p.host.Fetch(ctx,core.HTTPRequest{Label:"黃果舊版",Method:method,URL:cf109Base+path,Header:headers,Body:body,Limit:20<<20,Timeout:15*time.Second,Direct:true})
 if err!=nil{return nil,err}
 decoder:=json.NewDecoder(strings.NewReader(string(response.Body)));decoder.UseNumber()
 var env struct{Code any `json:"code"`;Msg string `json:"msg"`;Hash any `json:"hash"`;Data json.RawMessage `json:"data"`}
 if err:=decoder.Decode(&env);err!=nil{return nil,err}
 if hjString(env.Code)!="200"{return nil,fmt.Errorf("109 legacy API error: %s",hjString(env.Code))}
 payload:=env.Data
 if legacy109HashEnabled(env.Hash){
  var blob string
  if err:=json.Unmarshal(payload,&blob);err!=nil{return nil,err}
  payload,err=legacy109Decrypt(blob,[]byte(protocol.InterfaceKey));if err!=nil{return nil,err}
 }
 var result any
 dec:=json.NewDecoder(strings.NewReader(string(payload)));dec.UseNumber()
 if err:=dec.Decode(&result);err!=nil{return nil,err}
 return result,nil
}
func (p *cf109Provider) get(ctx context.Context,path string,params any)(any,error){
 protocol,token,err:=p.authenticate(ctx);if err!=nil{return nil,err}
 return p.api(ctx,http.MethodGet,path,params,protocol,token)
}
func (p *cf109Provider) Categories(ctx context.Context,_ bool)([]core.Category,error){
 value,err:=p.get(ctx,"/api/app/playlet-tab/all",nil);if err!=nil{return nil,err}
 rows:=hjRows(value);if rows==nil{rows=hjRows(hjMap(value)["list"])}
 var out []core.Category
 for _,v:=range rows{
  m:=hjMap(v);id,name:=hjString(m["id"]),hjString(m["name"])
  if jgID(id)&&name!=""{out=append(out,core.Category{ID:id,Name:name})}
 }
 if len(out)==0{return nil,errors.New("legacy tabs unavailable")}
 return out,nil
}
func cfDrama109(m map[string]any)(core.Drama,error){
 id,name:=hjString(m["id"]),hgField(m,"title","name")
 if !jgID(id)||name==""{return core.Drama{},errors.New("invalid legacy drama")}
 cover:=core.ResolveProviderURL(cf109Site+"/",hgField(m,"coverUrl","cover","imageUrl","image"))
 return core.Drama{ID:core.ProviderDramaID("cloudfront",id),Source:"cloudfront",SourceID:id,Title:name,Name:name,Cover:cover,ChannelName:"黃果舊版"},nil
}
func (p *cf109Provider) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1||page>100000||(category!=""&&!jgID(category)){return nil,false,errors.New("invalid category")}
 cats:=[]core.Category{}
 if category==""{var err error;cats,err=p.Categories(ctx,false);if err!=nil{return nil,false,err}
 }else{cats=[]core.Category{{ID:category}}}
 seen:=map[string]bool{};var all []core.Drama;more:=false
 for _,cat:=range cats{
  value,err:=p.get(ctx,"/api/app/playlet/home/tab/"+url.PathEscape(cat.ID),map[string]string{"pageNumber":strconv.Itoa(page),"pageSize":"24","tabSortType":"1"})
  if err!=nil{return nil,false,err}
  rows:=hjRows(hjMap(value)["list"]);more=more||len(rows)>=24
  for _,v:=range rows{item,err:=cfDrama109(hjMap(v));if err==nil&&!seen[item.SourceID]{seen[item.SourceID]=true;all=append(all,item)}}
 }
 return all,more,nil
}
func (p *cf109Provider) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 if strings.TrimSpace(query)==""{return nil,false,nil}
 if len(query)>256||page<1{return nil,false,errors.New("invalid search")}
 value,err:=p.get(ctx,"/api/app/playlet/keyword/search",map[string]string{"keyword":query,"pageNumber":strconv.Itoa(page),"pageSize":"24"})
 if err!=nil{return nil,false,err}
 rows:=hjRows(hjMap(value)["list"]);var out []core.Drama
 for _,v:=range rows{item,err:=cfDrama109(hjMap(v));if err==nil{out=append(out,item)}}
 return out,len(rows)>=24,nil
}
func (p *cf109Provider) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !jgID(id){return core.Drama{},nil,errors.New("invalid legacy ID")}
 value,err:=p.get(ctx,"/api/app/playlet/detail/"+url.PathEscape(id),nil)
 if err!=nil{return core.Drama{},nil,err}
 m:=hjMap(value);drama,err:=cfDrama109(m)
 if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("legacy drama mismatch")}
 rows:=hjRows(m["chapters"])
 if len(rows)==0{
  value,err=p.get(ctx,"/api/app/playlet-chapter/list/"+url.PathEscape(id),nil)
  if err!=nil{return core.Drama{},nil,err}
  rows=hjRows(value)
 }
 var chapters []core.Chapter
 for i,v:=range rows{
  row:=hjMap(v);eid:=hjString(row["id"]);if eid==""{eid=strconv.Itoa(i+1)}
  if !jgID(eid){continue}
  name:=hjString(row["title"]);if name==""{name=fmt.Sprintf("第%d集",i+1)}
  video:=hgField(row,"videoUrl","video_url")
  chapters=append(chapters,core.Chapter{ID:core.ProviderChapterID("cloudfront",id,eid),Source:"cloudfront",
   Title:name,CurrentEpisode:core.RawEpisode(i+1),VideoURL:video,VIP:video=="",Referer:cf109Site+"/"})
 }
 core.SortProviderChapters(chapters);drama.TotalEpisode=len(chapters)
 return drama,chapters,nil
}
func (p *cf109Provider) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="cloudfront"||!jgID(id){return core.ProviderMedia{},errors.New("invalid legacy drama")}
 prefix:=core.ProviderChapterID("cloudfront",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid legacy chapter")}
 video:=strings.Trim(task.Chapter.VideoURL,"/")
 if video==""{return core.ProviderMedia{},errors.New("locked or unavailable episode")}
 _,token,err:=p.authenticate(ctx);if err!=nil{return core.ProviderMedia{},err}
 address:=task.Chapter.VideoURL
 if !core.IsProviderHTTPMediaURL(address){
  if strings.Contains(video,"://")||strings.ContainsAny(video,"?#\\\\")||strings.Contains(video,".."){return core.ProviderMedia{},errors.New("invalid playlist identifier")}
  address=cf109Base+"/api/app/vid/h5/m3u8/"+url.PathEscape(video)+"?"+url.Values{"token":{token},"c":{cf109CDN}}.Encode()
 }
 playlist,final,err:=p.host.FetchPlaylist(ctx,address,cf109Site+"/",nil)
 if err!=nil{return core.ProviderMedia{},err}
 if !strings.HasPrefix(strings.TrimSpace(playlist),"#EXTM3U"){return core.ProviderMedia{},errors.New("no authorized HLS playlist")}
 return core.ProviderMedia{URL:final,Playlist:playlist,Referer:cf109Site+"/"},nil
}
