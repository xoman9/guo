package legacy109
import(
 "bytes"
 "compress/gzip"
 "context"
 "crypto/aes"
 "crypto/cipher"
 "crypto/hmac"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "time"
 "duanjuapp/native/core"
)
const hd109Base="https://ferncider.cc"
const hd109Platform="7961beb44246e3012ce228d6b5ced05a"
var hd109Hosts=[]string{hd109Base,"https://onyxripple.cc","https://reeflumen.cc","https://niniharbor.top","https://halomeadow.top","https://larksummit.icu"}
type huangdou109 struct{host core.Host; session,device string}
var hd109Desc=core.SourceDescriptor{ID:"huangdou",Name:"黃豆",BaseURL:hd109Base,NoHostRotation:true,
 Description:"109 黃豆簽章 API、資料壓縮與標準加密傳輸的 110 相容版；尚未線上驗證",
 ValidCategory:func(s string)bool{return s==""},
}
var _ core.Provider=(*huangdou109)(nil)
func init(){core.Register(hd109Desc,func(h core.Host)core.Provider{return &huangdou109{host:h}})}
func (p *huangdou109) Descriptor()core.SourceDescriptor{return hd109Desc}
func (p *huangdou109) BaseURL()string{return hd109Base}
func hd109UUID()string{
 var b [16]byte
 if _,err:=rand.Read(b[:]);err!=nil{return ""}
 return fmt.Sprintf("%x-%x-%x-%x-%x",b[:4],b[4:6],b[6:8],b[8:10],b[10:])
}
func hd109Key(rid string)([]byte,error){
 data,err:=hex.DecodeString(strings.ReplaceAll(rid,"-",""))
 if err!=nil{return nil,err}
 mac:=hmac.New(sha256.New,[]byte(hd109Platform))
 _,_ =mac.Write(data)
 return mac.Sum(nil),nil
}
func hd109Encrypt(data,key,iv []byte)([]byte,error){
 block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
 if len(iv)!=aes.BlockSize{return nil,errors.New("bad AES IV")}
 padded:=legacy109Pad(data,block.BlockSize())
 out:=make([]byte,len(padded));cipher.NewCBCEncrypter(block,iv).CryptBlocks(out,padded)
 return out,nil
}
func hd109Decrypt(data,key,iv []byte)([]byte,error){
 block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
 if len(iv)!=aes.BlockSize||len(data)==0||len(data)%aes.BlockSize!=0{return nil,errors.New("invalid encrypted response")}
 out:=make([]byte,len(data));cipher.NewCBCDecrypter(block,iv).CryptBlocks(out,data)
 return legacy109Unpad(out,block.BlockSize())
}
func hd109Decode(blob,key []byte)(any,error){
 if len(blob)>20<<20{return nil,errors.New("response too large")}
 var raw []byte
 if json.Valid(blob){raw=blob}else{
  if len(blob)<aes.BlockSize*2{return nil,errors.New("short encrypted response")}
  var err error
  raw,err=hd109Decrypt(blob[aes.BlockSize:],key,blob[:aes.BlockSize])
  if err!=nil{return nil,err}
  if len(raw)>1&&raw[0]==0x1f&&raw[1]==0x8b{
   reader,err:=gzip.NewReader(bytes.NewReader(raw));if err!=nil{return nil,err}
   raw,err=io.ReadAll(io.LimitReader(reader,20<<20+1))
   _=reader.Close()
   if err!=nil||len(raw)>20<<20{return nil,errors.New("decompressed response too large")}
  }
 }
 if !json.Valid(raw){return nil,errors.New("invalid Huangdou JSON")}
 var decoded any
 d:=json.NewDecoder(bytes.NewReader(raw));d.UseNumber()
 if err:=d.Decode(&decoded);err!=nil{return nil,err}
 return decoded,nil
}
func (p *huangdou109) init(){
 if p.session!=""{return}
 p.session=hd109UUID()
 p.device=p.session
}
func (p *huangdou109) call(ctx context.Context,route string,data map[string]any)(any,error){
 if route==""||!strings.HasPrefix(route,"/drama/"){return nil,errors.New("invalid Huangdou API route")}
 p.init()
 if p.session==""{return nil,errors.New("no device ID")}
 rid:=hd109UUID()
 key,err:=hd109Key(rid);if err!=nil{return nil,err}
 payload,_:=json.Marshal(map[string]any{"token":"","deviceId":p.device,"data":data})
 var compressed bytes.Buffer
 writer:=gzip.NewWriter(&compressed)
 if _,err=writer.Write(payload);err!=nil{return nil,err}
 if err=writer.Close();err!=nil{return nil,err}
 iv:=make([]byte,aes.BlockSize)
 if _,err=rand.Read(iv);err!=nil{return nil,err}
 cipherText,err:=hd109Encrypt(compressed.Bytes(),key,iv)
 if err!=nil{return nil,err}
 body:=append(append([]byte{},iv...),cipherText...)
 ts:=strconv.FormatInt(time.Now().Unix(),10)
 rawSignature:="Dart|"+p.session+"|"+rid+"|"+ts+"|"+route
 digest:=sha256.Sum256([]byte(rawSignature))
 headers:=map[string]string{
  "Content-Type":"application/octet-stream","Accept":"*/*",
  "version":"2.0.0","deviceType":"web",
  "requestId":rid,"sessionId":p.session,
  "time":ts,"sign":hex.EncodeToString(digest[:])+"-"+ts,
 }
 var lastErr error
 for _,host:=range hd109Hosts{
  headers["Origin"]=host;headers["Referer"]=host+"/home"
  response,err:=p.host.Fetch(ctx,core.HTTPRequest{
   Label:"黃豆",Method:http.MethodPost,URL:host+"/api"+route,
   Body:body,Header:headers,Timeout:12*time.Second,Limit:20<<20,Direct:true,
  })
  if err!=nil{lastErr=err;continue}
  if response.Status>=400{
   return nil,fmt.Errorf("Huangdou API HTTP %d; access restrictions respected",response.Status)
  }
  decoded,err:=hd109Decode(response.Body,key)
  if err!=nil{return nil,err}
  m:=hjMap(decoded)
  if status:=hjString(m["status"]);status!=""&&status!="y"{
   return nil,fmt.Errorf("Huangdou API denied request: %s",hgField(m,"errorCode","code"))
  }
  return decoded,nil
 }
 if lastErr==nil{lastErr=errors.New("no accessible Huangdou site")}
 return nil,lastErr
}
func hd109Data(v any)map[string]any{
 m:=hjMap(v)
 if nested:=hjMap(m["data"]);nested!=nil{return nested}
 return m
}
func hd109Rows(v any)[]any{
 if rows:=hjRows(v);rows!=nil{return rows}
 m:=hjMap(v)
 for _,key:=range []string{"list","items","data"}{
  if rows:=hjRows(m[key]);rows!=nil{return rows}
 }
 return nil
}
func hd109Drama(m map[string]any)(core.Drama,error){
 id:=strings.TrimPrefix(hgField(m,"id","drama_id"),"rp_")
 name:=hgField(m,"name","title","t")
 if !jgID(id)||name==""{return core.Drama{},errors.New("invalid Huangdou drama")}
 cover:=hgField(m,"img_y","img_x","img","cover","pic")
 count,_:=hjInt(m["episode_count"],10000)
 return core.Drama{ID:core.ProviderDramaID("huangdou",id),Source:"huangdou",SourceID:id,
  Title:name,Name:name,Cover:core.ResolveProviderURL(hd109Base+"/",cover),
  TotalEpisode:count,Desc:hgField(m,"description","summary"),ChannelName:"黃豆"},nil
}
func (p *huangdou109) Categories(context.Context,bool)([]core.Category,error){
 return []core.Category{{ID:"",Name:"全部"}},nil
}
func (p *huangdou109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1||page>100000||category!=""{return nil,false,errors.New("invalid Huangdou page/category")}
 response,err:=p.call(ctx,"/drama/list",map[string]any{
  "page":strconv.Itoa(page),"page_size":"30",
 })
 if err!=nil{return nil,false,err}
 rows:=hd109Rows(response)
 var out []core.Drama
 for _,v:=range rows{
  drama,err:=hd109Drama(hjMap(v))
  if err==nil{out=append(out,drama)}
 }
 return out,len(rows)>=30,nil
}
func (p *huangdou109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !jgID(id){return core.Drama{},nil,errors.New("invalid Huangdou drama ID")}
 data,err:=p.call(ctx,"/drama/detail",map[string]any{"id":id})
 if err!=nil{return core.Drama{},nil,err}
 m:=hd109Data(data)
 drama,err:=hd109Drama(m)
 if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("Huangdou detail mismatch")}
 episodes:=hd109Rows(m["episodes"])
 count:=drama.TotalEpisode.(int)
 if count==0{count=len(episodes)}
 if count<1||count>10000{return core.Drama{},nil,errors.New("invalid episode count")}
 var chapters []core.Chapter
 vipSeen:=map[string]bool{}
 for _,v:=range hd109Rows(m["vip_episodes"]){vipSeen[hjString(v)]=true}
 for i:=1;i<=count;i++{
  item:=map[string]any{}
  if i<=len(episodes){item=hjMap(episodes[i-1])}
  seq:=i
  if value,ok:=hjInt(item["seq"],10000);ok&&value>0{seq=value}
  title:=hgField(item,"name","title")
  if title==""{title=fmt.Sprintf("第%d集",seq)}
  locked:=vipSeen[strconv.Itoa(seq)]
  if paid,ok:=item["is_vip"].(bool);ok&&paid{locked=true}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("huangdou",id,strconv.Itoa(seq)),Source:"huangdou",
   Title:title,CurrentEpisode:core.RawEpisode(seq),VIP:locked,
  })
 }
 core.SortProviderChapters(chapters)
 return drama,chapters,nil
}
func (p *huangdou109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="huangdou"||!jgID(id){return core.ProviderMedia{},errors.New("invalid Huangdou drama")}
 prefix:=core.ProviderChapterID("huangdou",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid Huangdou chapter")}
 seq,err:=strconv.Atoi(strings.TrimPrefix(task.Chapter.ID,prefix))
 if err!=nil||seq<1||seq>10000{return core.ProviderMedia{},errors.New("invalid episode")}
 response,err:=p.call(ctx,"/drama/play",map[string]any{"id":id,"seq":strconv.Itoa(seq)})
 if err!=nil{return core.ProviderMedia{},err}
 m:=hd109Data(response)
 if preview,ok:=m["is_preview"].(bool);ok&&preview{return core.ProviderMedia{},errors.New("Huangdou only returned a preview")}
 address:=hgField(m,"m3u8","url","play_url","playUrl")
 if preview:=hgField(m,"preview_m3u8","preview_url");preview!=""&&(address==""||address==preview){
  return core.ProviderMedia{},errors.New("preview is not the requested episode")
 }
 if address==""{return core.ProviderMedia{},errors.New("Huangdou did not issue an authorized media URL")}
 mediaURL:=core.NormalizeProviderMediaURL(hd109Base+"/",address)
 if !core.IsProviderHTTPMediaURL(mediaURL){return core.ProviderMedia{},errors.New("invalid Huangdou media URL")}
 media:=core.ProviderMedia{URL:mediaURL,Referer:hd109Base+"/home"}
 if rawKey:=hjString(m["hls_key"]);rawKey!=""{
  key,err:=hex.DecodeString(rawKey)
  if err!=nil||len(key)!=aes.BlockSize{return core.ProviderMedia{},errors.New("invalid server-provided HLS key")}
  media.HLSKey=key
 }
 return p.host.PrepareWebMedia(ctx,media,"黃豆")
}
