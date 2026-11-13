package legacy109
import(
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
const jgBase="https://d1wyuxwdrlykj1.cloudfront.net"
const jgBackup="https://hghts.rlkfus.cn"
const jgFrontend="https://d2pypzndaqisk.cloudfront.net"
const jgParamKey="BxJand%xf5h3sycH"
const jgInterfaceKey="65dc07d1b7915c6b2937432b091837a7"
type jinguo109 struct{host core.Host;mu sync.Mutex; token,device string}
var jgDesc=core.SourceDescriptor{
 ID:"jinguo",Name:"禁果",BaseURL:jgFrontend,NoHostRotation:true,
 Description:"109 加密 API 傳輸協定、裝置訪客授權及合法播放路線的 110 相容版；未經線上驗證",
 ValidCategory:jgValidCat,
}
var _ core.Provider=(*jinguo109)(nil)
var _ core.Searcher=(*jinguo109)(nil)
func init(){core.Register(jgDesc,func(h core.Host)core.Provider{return &jinguo109{host:h}})}
func (p *jinguo109) Descriptor()core.SourceDescriptor{return jgDesc}
func (p *jinguo109) BaseURL()string{return jgFrontend}
func jgID(s string)bool{
 if len(s)==0||len(s)>64{return false}
 for _,r:=range s {if r<'0'||r>'9'{if r<'A'||r>'Z'{if r<'a'||r>'z'{return false}}}}
 return true
}
func jgValidCat(v string)bool{
 if v==""||v=="choose"||v=="new"{return true}
 if strings.HasPrefix(v,"tab:"){return jgID(strings.TrimPrefix(v,"tab:"))}
 return false
}
func (p *jinguo109) deviceID()(string,error){
 p.mu.Lock();defer p.mu.Unlock()
 if p.device!=""{return p.device,nil}
 var raw [8]byte
 if _,err:=rand.Read(raw[:]);err!=nil{return "",err}
 p.device=strings.ToUpper(fmt.Sprintf("%x",raw[:]))+strconv.FormatInt(time.Now().UnixMilli(),10)
 return p.device,nil
}
func (p *jinguo109) api(ctx context.Context,method,path string,parameters any,token string)(any,error){
 if !strings.HasPrefix(path,"/api/app/")||strings.Contains(path,".."){return nil,errors.New("invalid Jinguo API path")}
 device,err:=p.deviceID();if err!=nil{return nil,err}
 var body []byte
 if parameters!=nil{
  encoded,err:=legacy109Encrypt(parameters,[]byte(jgParamKey),[]byte(jgParamKey))
  if err!=nil{return nil,err}
  if method==http.MethodGet{path+="?data="+url.QueryEscape(encoded)
  }else{body,_=json.Marshal(map[string]string{"data":encoded})}
 }
 headers:=map[string]string{
  "User-Agent":ua,"Accept":"application/json",
  "X-User-Agent":fmt.Sprintf("BuildID=com.hgaidj.ozThree;SysType=android;DevID=%s;Ver=1.0.8;DevType=Android;DeviceBrand=google;DeviceModel=Pixel;SystemName=Android;SystemVersion=14;Terminal=1;IsH5=1;Sid=00000000000000000000000000000",device),
  "temp":"test","Referer":jgFrontend+"/",
 }
 if token!=""{headers["Authorization"]=token}
 if len(body)>0{headers["Content-Type"]="application/json"}
 var response core.HTTPResponse
 var lastErr error
 for _,base:=range []string{jgBase,jgBackup}{
  response,lastErr=p.host.Fetch(ctx,core.HTTPRequest{
   Label:"禁果",Method:method,URL:base+path,Body:body,Header:headers,
   Timeout:15*time.Second,Limit:20<<20,Direct:true,
  })
  if lastErr==nil{break}
 }
 if lastErr!=nil{return nil,lastErr}
 decoder:=json.NewDecoder(strings.NewReader(string(response.Body)));decoder.UseNumber()
 var envelope struct{Code any `json:"code"`;Hash any `json:"hash"`;Msg string `json:"msg"`;Data json.RawMessage `json:"data"`}
 if err:=decoder.Decode(&envelope);err!=nil{return nil,err}
 if hjString(envelope.Code)!="200"{return nil,fmt.Errorf("Jinguo API error %s: %s",hjString(envelope.Code),envelope.Msg)}
 payload:=envelope.Data
 if legacy109HashEnabled(envelope.Hash){
  var encrypted string
  if err:=json.Unmarshal(payload,&encrypted);err!=nil{return nil,err}
  payload,err=legacy109Decrypt(encrypted,[]byte(jgInterfaceKey))
  if err!=nil{return nil,err}
 }
 if len(payload)==0{return nil,errors.New("empty Jinguo API payload")}
 var result any
 jsonDecoder:=json.NewDecoder(strings.NewReader(string(payload)));jsonDecoder.UseNumber()
 if err:=jsonDecoder.Decode(&result);err!=nil{return nil,err}
 return result,nil
}
func (p *jinguo109) tokenFor(ctx context.Context)(string,error){
 p.mu.Lock();current:=p.token;p.mu.Unlock()
 if current!=""{return current,nil}
 if raw,err:=p.host.LoadState("jinguo-session.json");err==nil&&len(raw)>0{
  var state struct{DeviceID string `json:"deviceId"`;Token string `json:"token"`}
  if json.Unmarshal(raw,&state)==nil&&state.Token!=""&&len(state.Token)<=8192{
   p.mu.Lock();p.device=state.DeviceID;p.token=state.Token;p.mu.Unlock()
   return state.Token,nil
  }
 }
 device,err:=p.deviceID();if err!=nil{return "",err}
 data,err:=p.api(ctx,http.MethodPost,"/api/app/mine/login/h5",map[string]any{
  "devID":device,"sysType":"ios","isAppStore":false,
 },"")
 if err!=nil{return "",err}
 token:=hjString(hjMap(data)["token"])
 if token==""||len(token)>8192||strings.ContainsAny(token,"\\r\\n\\t "){return "",errors.New("no authorized guest token")}
 p.mu.Lock();p.token=token;p.mu.Unlock()
 saved,_:=json.Marshal(map[string]any{"version":1,"deviceId":device,"token":token})
 _=p.host.SaveState("jinguo-session.json",saved)
 return token,nil
}
func (p *jinguo109) get(ctx context.Context,path string,params any)(any,error){
 token,err:=p.tokenFor(ctx);if err!=nil{return nil,err}
 return p.api(ctx,http.MethodGet,path,params,token)
}
func (p *jinguo109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 values,err:=p.get(ctx,"/api/app/playlet-tab/all",nil)
 if err!=nil{return nil,err}
 rows:=hjRows(values);if rows==nil{rows=hjRows(hjMap(values)["list"])}
 out:=[]core.Category{{ID:"choose",Name:"精選"},{ID:"new",Name:"最新"}}
 seen:=map[string]bool{}
 for _,v:=range rows{
  m:=hjMap(v);id,name:=hjString(m["id"]),hjString(m["name"])
  if !jgID(id)||name==""||seen[id]{continue}
  seen[id]=true;out=append(out,core.Category{ID:"tab:"+id,Name:name})
 }
 return out,nil
}
func jgDrama(m map[string]any)(core.Drama,error){
 id:=hjString(m["id"])
 name:=hgField(m,"title","name")
 if !jgID(id)||name==""{return core.Drama{},errors.New("invalid Jinguo drama")}
 cover:=core.ResolveProviderURL(jgFrontend+"/",hgField(m,"coverUrl","cover","imageUrl","image"))
 count,_:=hjInt(m["totalEpisode"],100000)
 return core.Drama{
  ID:core.ProviderDramaID("jinguo",id),Source:"jinguo",SourceID:id,
  Title:name,Name:name,Cover:cover,TotalEpisode:count,ChannelName:"禁果",
 },nil
}
func (p *jinguo109) catalog(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>100000||!jgValidCat(category)||len(query)>256{return nil,false,errors.New("invalid catalog request")}
 route:="/api/app/playlet-recommend/home/choose"
 params:=map[string]string{"pageNumber":strconv.Itoa(page),"pageSize":"24"}
 switch{
 case query!="":route="/api/app/playlet/keyword/search";params["keyword"]=query
 case category=="new":route="/api/app/playlet-recommend/home/new"
 case strings.HasPrefix(category,"tab:"):
  route="/api/app/playlet/home/tab/"+url.PathEscape(strings.TrimPrefix(category,"tab:"))
  params["tabSortType"]="1"
 }
 result,err:=p.get(ctx,route,params)
 if err!=nil{return nil,false,err}
 rows:=hjRows(hjMap(result)["list"]);if rows==nil{rows=hjRows(result)}
 var out []core.Drama
 for _,r:=range rows{item,err:=jgDrama(hjMap(r));if err==nil{out=append(out,item)}}
 return out,len(rows)>=24,nil
}
func (p *jinguo109) CatalogPage(ctx context.Context,page int,cat string)([]core.Drama,bool,error){return p.catalog(ctx,page,cat,"")}
func (p *jinguo109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query);if query==""{return nil,false,nil}
 return p.catalog(ctx,page,"",query)
}
func (p *jinguo109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !jgID(id){return core.Drama{},nil,errors.New("invalid Jinguo drama")}
 data,err:=p.get(ctx,"/api/app/playlet/detail/"+url.PathEscape(id),nil)
 if err!=nil{return core.Drama{},nil,err}
 m:=hjMap(data)
 drama,err:=jgDrama(m)
 if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("detail drama mismatch")}
 rows:=hjRows(m["chapters"])
 if len(rows)==0{
  list,err:=p.get(ctx,"/api/app/playlet-chapter/list/"+url.PathEscape(id),nil)
  if err!=nil{return core.Drama{},nil,err}
  rows=hjRows(list)
 }
 var chapters []core.Chapter
 seen:=map[string]bool{}
 for i,v:=range rows{
  entry:=hjMap(v);ep:=hjString(entry["id"])
  if ep==""{ep=strconv.Itoa(i+1)}
  if !jgID(ep)||seen[ep]{return core.Drama{},nil,errors.New("duplicate Jinguo episode")}
  seen[ep]=true
  video:=hgField(entry,"videoUrl","video_url")
  allowed:=video!=""
  if canPlay,ok:=entry["canPlay"].(bool);ok&&!canPlay{allowed=false}
  title:=hjString(entry["title"])
  if title==""{title=fmt.Sprintf("第%d集",i+1)}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("jinguo",id,ep),Source:"jinguo",Title:title,
   CurrentEpisode:core.RawEpisode(i+1),VIP:!allowed,VideoURL:video,
  })
 }
 core.SortProviderChapters(chapters)
 drama.TotalEpisode=len(chapters)
 return drama,chapters,nil
}
func (p *jinguo109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!="jinguo"||!jgID(id){return core.ProviderMedia{},errors.New("wrong drama")}
 prefix:=core.ProviderChapterID("jinguo",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("wrong episode")}
 ep:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !jgID(ep){return core.ProviderMedia{},errors.New("invalid episode")}
 video:=strings.Trim(task.Chapter.VideoURL,"/")
 if video==""{return core.ProviderMedia{},errors.New("episode requires paid authorization or unlock")}
 if strings.ContainsAny(video,"?#\\\\")||strings.Contains(video,"://")||strings.Contains(video,".."){
  return core.ProviderMedia{},errors.New("invalid encoded media path")
 }
 token,err:=p.tokenFor(ctx);if err!=nil{return core.ProviderMedia{},err}
 path:=strings.Join(strings.Split(video,"/"),"/")
 address:=jgBase+"/api/app/vid/h5/m3u8/"+path+"?"+url.Values{"token":{token}}.Encode()
 playlist,final,err:=p.host.FetchPlaylist(ctx,address,jgBase+"/",nil)
 if err!=nil{return core.ProviderMedia{},err}
 if !strings.HasPrefix(strings.TrimSpace(playlist),"#EXTM3U"){return core.ProviderMedia{},errors.New("no verified HLS from authorized API")}
 return core.ProviderMedia{URL:final,Playlist:playlist,Referer:jgBase+"/"},nil
}
