package legacy109
import(
 "context"
 "crypto/aes"
 "crypto/cipher"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "sort"
 "strconv"
 "strings"
 "time"
 "duanjuapp/native/core"
)
const bbBase109="https://bh4066.top"
const bbKey109="f6322fa1c064370ea40c8cdc20649f8e"
var bbHosts109=[]string{bbBase109,"https://beabox.live"}
type beabox109 struct{host core.Host}
var bbDesc109=core.SourceDescriptor{
 ID:"beabox",Name:"筆盒 +",BaseURL:bbBase109,NoHostRotation:true,DefaultOff:true,
 Description:"109 Beabox 公開 API 與 AES 內容解碼、分集及播放器解析相容版；未經線上驗證",
 ValidCategory:bbCategory109,
}
var _ core.Provider=(*beabox109)(nil)
var _ core.Searcher=(*beabox109)(nil)
func init(){core.Register(bbDesc109,func(h core.Host)core.Provider{return &beabox109{host:h}})}
func (p *beabox109) Descriptor()core.SourceDescriptor{return bbDesc109}
func (p *beabox109) BaseURL()string{return bbBase109}
func bbID109(id string)bool{
 if len(id)==0||len(id)>128{return false}
 for _,r:=range id{if r=='/'||r=='\\'||r=='?'||r=='#'||r<32{return false}}
 return true
}
func bbCategory109(cat string)bool{
 if cat==""{return true}
 if !strings.HasPrefix(cat,"tag:"){return false}
 name:=strings.TrimSpace(strings.TrimPrefix(cat,"tag:"))
 return name!=""&&len([]rune(name))<=24&&!strings.ContainsAny(name,"|/\\\r\n")
}
func bbDecode109(v any)(any,error){
 str,ok:=v.(string)
 if !ok{return v,nil}
 if len(str)<64||len(str)%2!=0{return v,nil}
 for _,r:=range str{if !strings.ContainsRune("0123456789abcdefABCDEF",r){return v,nil}}
 raw,err:=hex.DecodeString(str)
 if err!=nil||len(raw)<32||(len(raw)-16)%16!=0{return nil,errors.New("invalid Beabox ciphertext")}
 key,_:=hex.DecodeString(bbKey109)
 block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
 decrypted:=make([]byte,len(raw)-16)
 cipher.NewCBCDecrypter(block,raw[:16]).CryptBlocks(decrypted,raw[16:])
 decrypted,err=legacy109Unpad(decrypted,16)
 if err!=nil{return nil,err}
 var result any
 decoder:=json.NewDecoder(strings.NewReader(string(decrypted)));decoder.UseNumber()
 if err:=decoder.Decode(&result);err!=nil{return nil,err}
 return result,nil
}
func (p *beabox109) call(ctx context.Context,path string,q url.Values,payload any)(any,string,error){
 if !strings.HasPrefix(path,"/")||strings.Contains(path,".."){return nil,"",errors.New("invalid Beabox API route")}
 method:=http.MethodGet
 var body []byte
 if payload!=nil{method=http.MethodPost;body,_=json.Marshal(payload)}
 if len(q)>0{path+="?"+q.Encode()}
 var lastErr error
 for _,site:=range bbHosts109{
  headers:=map[string]string{"Accept":"application/json","User-Agent":ua,"Origin":site,"Referer":site+"/home"}
  if body!=nil{headers["Content-Type"]="application/json"}
  r,err:=p.host.Fetch(ctx,core.HTTPRequest{
   Label:"筆盒",Method:method,URL:site+"/api"+path,Header:headers,
   Body:body,Timeout:15*time.Second,Limit:8<<20,Direct:true,
  })
  if err!=nil{lastErr=err;continue}
  var env map[string]any
  dec:=json.NewDecoder(strings.NewReader(string(r.Body)));dec.UseNumber()
  if dec.Decode(&env)!=nil{lastErr=errors.New("invalid Beabox API envelope");continue}
  if hjString(env["code"])!="200"{return nil,site,errors.New("Beabox API rejected request")}
  data,err:=bbDecode109(env["data"]);return data,site,err
 }
 if lastErr==nil{lastErr=errors.New("Beabox unavailable")}
 return nil,"",lastErr
}
func (p *beabox109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 v,_,err:=p.call(ctx,"/configs",url.Values{"platformType":{"h5"}},nil)
 if err!=nil{return nil,err}
 tags:=hjString(hjMap(v)["home_tags"])
 out:=[]core.Category{{ID:"",Name:"最新"}}
 seen:=map[string]bool{}
 for _,s:=range strings.Split(tags,","){
  name:=strings.TrimSpace(s)
  if !bbCategory109("tag:"+name)||seen[name]{continue}
  out=append(out,core.Category{ID:"tag:"+name,Name:name});seen[name]=true
 }
 return out,nil
}
func bbDrama109(v any,site string)(core.Drama,error){
 m:=hjMap(v)
 id,title:=hgField(m,"vodId"),hgField(m,"vodName")
 if !bbID109(id)||title==""{return core.Drama{},errors.New("invalid Beabox drama")}
 return core.Drama{
  ID:core.ProviderDramaID("beabox",id),Source:"beabox",SourceID:id,
  Title:title,Name:title,Cover:core.ResolveProviderURL(site+"/",hgField(m,"vodPic")),
  ChannelName:"筆盒",
 },nil
}
func (p *beabox109) catalog(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>1000000||!bbCategory109(category)||len(query)>256{return nil,false,errors.New("invalid Beabox catalog")}
 q:=url.Values{"page":{strconv.Itoa(page)},"limit":{"20"}}
 route:="/vod/latest"
 if query!=""||category!=""{
  route="/vod/search"
  keyword:=query;if keyword==""{keyword=strings.TrimPrefix(category,"tag:")}
  q.Set("keyword",keyword);q.Set("sort","latest")
 }
 v,site,err:=p.call(ctx,route,q,nil)
 if err!=nil{return nil,false,err}
 m:=hjMap(v);rows:=hjRows(m["data"])
 if rows==nil{return nil,false,errors.New("invalid Beabox content list")}
 seen:=map[string]bool{};var items []core.Drama
 for _,raw:=range rows{
  drama,err:=bbDrama109(raw,site);if err!=nil||seen[drama.SourceID]{continue}
  seen[drama.SourceID]=true;items=append(items,drama)
 }
 pages,_:=hjInt(m["totalPages"],1000000)
 return items,pages>page||pages==0&&len(rows)>=20,nil
}
func (p *beabox109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 return p.catalog(ctx,page,category,"")
}
func (p *beabox109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 return p.catalog(ctx,page,"",query)
}
func bbGroups109(raw any)map[string][]any{
 switch v:=raw.(type){
 case map[string]any:
  out:=map[string][]any{}
  for name,entries:=range v{if rows,ok:=entries.([]any);ok{out[name]=rows}}
  return out
 case []any:return map[string][]any{"":v}
 case string:return map[string][]any{"":[]any{map[string]any{"url":v}}}
 }
 return nil
}
func bbEntries109(row map[string]any)map[string][]any{
 if row["vodPlaySource"]!=nil{return bbGroups109(row["vodPlaySource"])}
 return bbGroups109(row["vodPlay"])
}
func bbChapters109(row map[string]any,id string)[]core.Chapter{
 groups:=bbEntries109(row)
 names:=make([]string,0,len(groups))
 for name:=range groups{names=append(names,name)}
 sort.Strings(names)
 bestName:="";bestScore:=-1
 for _,name:=range names{
  count:=0
  for _,e:=range groups[name]{
   if hgField(hjMap(e),"url","playUrl")!=""{count++}
  }
  if count>bestScore||count==bestScore&&name=="beabox-app"{bestName,bestScore=name,count}
 }
 var out []core.Chapter
 for index,value:=range groups[bestName]{
  e:=hjMap(value)
  if hgField(e,"url","playUrl")==""{continue}
  title:=hgField(e,"flag");if title==""{title=fmt.Sprintf("第%d集",len(out)+1)}
  out=append(out,core.Chapter{
   ID:core.ProviderChapterID("beabox",id,fmt.Sprintf("%s:%d",bestName,index)),
   Source:"beabox",Title:title,CurrentEpisode:core.RawEpisode(len(out)+1),
  })
 }
 return out
}
func (p *beabox109) detail(ctx context.Context,id string)(map[string]any,string,error){
 if !bbID109(id){return nil,"",errors.New("invalid Beabox ID")}
 v,site,err:=p.call(ctx,"/vod/detail/"+url.PathEscape(id),nil,nil)
 if err!=nil{return nil,site,err}
 m:=hjMap(v)
 if hgField(m,"vodId")!=id{return nil,site,errors.New("Beabox detail mismatch")}
 return m,site,nil
}
func (p *beabox109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 m,site,err:=p.detail(ctx,id);if err!=nil{return core.Drama{},nil,err}
 drama,err:=bbDrama109(m,site);if err!=nil{return core.Drama{},nil,err}
 chapters:=bbChapters109(m,id)
 if len(chapters)==0{return core.Drama{},nil,errors.New("no playable Beabox episodes")}
 drama.TotalEpisode=len(chapters)
 return drama,chapters,nil
}
func (p *beabox109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="beabox"||!bbID109(id){return core.ProviderMedia{},errors.New("invalid Beabox drama")}
 prefix:=core.ProviderChapterID("beabox",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid Beabox episode")}
 suffix:=strings.TrimPrefix(task.Chapter.ID,prefix)
 group,indexText,found:=strings.Cut(suffix,":")
 if !found{return core.ProviderMedia{},errors.New("invalid episode key")}
 index,err:=strconv.Atoi(indexText)
 if err!=nil||index<0{return core.ProviderMedia{},errors.New("invalid episode index")}
 row,site,err:=p.detail(ctx,id);if err!=nil{return core.ProviderMedia{},err}
 entries:=bbEntries109(row)[group]
 if index>=len(entries){return core.ProviderMedia{},errors.New("episode changed")}
 item:=hjMap(entries[index])
 address:=hgField(item,"url","playUrl")
 if address==""{return core.ProviderMedia{},errors.New("episode no longer playable")}
 needsParse:=hjString(item["needParse"])=="1"||item["needParse"]==true||row["playNeedParse"]==true
 if needsParse{
  value,_,err:=p.call(ctx,"/vod/parse-url",nil,map[string]any{
   "playerFrom":hjString(item["from"]),"playUrl":hjString(item["playUrl"]),
  })
  if err!=nil{return core.ProviderMedia{},err}
  address=hjString(hjMap(value)["url"])
 }
 if !core.IsProviderHTTPMediaURL(address){return core.ProviderMedia{},errors.New("invalid Beabox media URL")}
 return p.host.PrepareWebMedia(ctx,core.ProviderMedia{URL:address,Referer:site+"/"},"筆盒")
}
