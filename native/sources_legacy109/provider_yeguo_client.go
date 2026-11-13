package legacy109
import (
 "bytes"
 "context"
 "crypto/aes"
 "crypto/cipher"
 "crypto/md5"
 "crypto/rand"
 "crypto/sha256"
 "crypto/subtle"
 "encoding/base64"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "regexp"
 "sort"
 "strconv"
 "strings"
 "sync"
 "time"
 "duanjuapp/native/core"
)
const yg109Site="https://attract.vgbyziuz.cc"
var yg109Sites=[]string{yg109Site,"https://as.vgbyziuz.cc","https://anyway.vgbyziuz.cc","https://analyze.buxefaex.cc"}
var yg109Fields=[]string{"theme","setting","background","time","recommend"}
var yg109Script=regexp.MustCompile(`(?is)<script\b[^>]*type=["']module["'][^>]*src=["']([^"']+\.js)["']`)
var yg109Import=regexp.MustCompile(`(?is)\bimport\s*\{[^{};]+\}\s*from\s*["']([^"']+\.js)["']`)
var yg109Config=regexp.MustCompile(`\b(version|mode|padding|key|iv|sign_key)\s*:\s*["']([^"'\\\r\n]{1,512})["']`)
var yg109BaseRX=regexp.MustCompile(`https?://[A-Za-z0-9._:-]+(?:/api\.php)?`)
type ygAccess109 struct{
 base,site,identifier string
 key,iv,signKey []byte
}
type yeguo109 struct{
 host core.Host
 mu sync.Mutex
 access *ygAccess109
}
var yg109Desc=core.SourceDescriptor{
 ID:"yeguo",Name:"野果",BaseURL:yg109Site,NoHostRotation:true,
 Description:"109 野果前端公開協定、簽章 API、分類及分集移植 110；未經線上驗證",
 ValidCategory:yg109ValidCat,
}
var _ core.Provider=(*yeguo109)(nil)
var _ core.Searcher=(*yeguo109)(nil)
func init(){core.Register(yg109Desc,func(h core.Host)core.Provider{return &yeguo109{host:h}})}
func (p *yeguo109) Descriptor()core.SourceDescriptor{return yg109Desc}
func (p *yeguo109) BaseURL()string{return yg109Site}
func yg109ValidCat(v string)bool{
 if v==""{return true}
 name,value,valid:=strings.Cut(v,":")
 if !valid||!dsd109Num(value){return false}
 for _,x:=range yg109Fields{if name==x{return true}}
 return false
}
func yg109Bytes(s string)[]byte{
 if !strings.Contains(s,"_"){return []byte(s)}
 var bytes []byte
 for _,part:=range strings.Split(s,"_"){
  number,err:=strconv.ParseUint(part,10,8)
  if err!=nil{return nil}
  bytes=append(bytes,byte(number))
 }
 return bytes
}
func yg109PublicConfig(script string)([]byte,[]byte,[]byte,error){
 fields:=map[string]string{}
 for _,match:=range yg109Config.FindAllStringSubmatch(script,48){fields[match[1]]=match[2]}
 if fields["version"]!="v0"||fields["mode"]!="CBC"||fields["padding"]!="Pkcs7"{
  return nil,nil,nil,errors.New("Yeguo public cipher version is unsupported")
 }
 key,iv,sign:=yg109Bytes(fields["key"]),yg109Bytes(fields["iv"]),yg109Bytes(fields["sign_key"])
 if len(key)!=16&&len(key)!=24&&len(key)!=32{return nil,nil,nil,errors.New("bad Yeguo key length")}
 if len(iv)!=16||len(sign)==0||len(sign)>128{return nil,nil,nil,errors.New("invalid Yeguo public protocol")}
 return key,iv,sign,nil
}
func yg109ScriptURL(site,reference string)string{
 base,err:=url.Parse(site);if err!=nil{return ""}
 target,err:=url.Parse(reference);if err!=nil{return ""}
 target=base.ResolveReference(target)
 if target.Scheme!="https"||target.Hostname()!=base.Hostname()||
 !strings.HasPrefix(target.Path,"/_nuxt/")||!strings.HasSuffix(target.Path,".js"){return ""}
 return target.String()
}
func yg109APIBase(script,origin string)string{
 for _,s:=range yg109BaseRX.FindAllString(script,-1){
  u,err:=url.Parse(s);if err!=nil||u.Scheme!="https"||u.User!=nil{continue}
  if strings.HasSuffix(u.Path,"/api.php"){return strings.TrimRight(s,"/")}
 }
 if strings.Contains(script,"/api.php"){return origin+"/api.php"}
 return ""
}
func (p *yeguo109) discover(ctx context.Context)(*ygAccess109,error){
 p.mu.Lock();if p.access!=nil{value:=p.access;p.mu.Unlock();return value,nil};p.mu.Unlock()
 var lastErr error
 for _,site:=range yg109Sites{
  r,err:=p.host.FetchText(ctx,core.TextRequest{URL:site+"/",Referer:site+"/",UserAgent:ua})
  if err!=nil{lastErr=err;continue}
  script:=yg109Script.FindStringSubmatch(r.Body)
  if len(script)<2{lastErr=errors.New("Yeguo module script not found");continue}
  entry:=yg109ScriptURL(site+"/",script[1])
  if entry==""{continue}
  modules:=[]string{entry}
  seen:=map[string]bool{}
  var key,iv,signature []byte
  apiBase:=""
  for len(modules)>0&&len(seen)<25{
   address:=modules[0];modules=modules[1:]
   if address==""||seen[address]{continue}
   seen[address]=true
   body,err:=p.host.FetchText(ctx,core.TextRequest{URL:address,Referer:site+"/",UserAgent:ua})
   if err!=nil{continue}
   if apiBase==""{apiBase=yg109APIBase(body.Body,site)}
   if key==nil{
    var e error
    key,iv,signature,e=yg109PublicConfig(body.Body)
    if e!=nil{key=nil}
   }
   for _,m:=range yg109Import.FindAllStringSubmatch(body.Body,30){
    next:=yg109ScriptURL(address,m[1])
    if next!=""&&!seen[next]{modules=append(modules,next)}
   }
  }
  if apiBase==""||key==nil{lastErr=errors.New("Yeguo public 109 frontend protocol unavailable");continue}
  var id [16]byte
  if _,err:=rand.Read(id[:]);err!=nil{return nil,err}
  entryAccess:=&ygAccess109{base:apiBase,site:site,identifier:hex.EncodeToString(id[:]),key:key,iv:iv,signKey:signature}
  p.mu.Lock();p.access=entryAccess;p.mu.Unlock()
  return entryAccess,nil
 }
 if lastErr==nil{lastErr=errors.New("Yeguo discovery unavailable")}
 return nil,lastErr
}
func yg109Signature(data map[string]any,key []byte)(string,error){
 var names []string
 for n,v:=range data{if n!="sign"&&n!="_ver"&&v!=nil{names=append(names,n)}}
 sort.Strings(names)
 var vals []string
 for _,n:=range names{
  value:=data[n]
  var s string
  switch v:=value.(type){
  case string:s=v
  case json.Number:s=string(v)
  case bool:s=strconv.FormatBool(v)
  default:return "",errors.New("Yeguo unexpected signed payload type")
  }
  if n=="data"{s=strings.ReplaceAll(s," ","+")}
  vals=append(vals,n+"="+s)
 }
 digest:=sha256.Sum256(append([]byte(strings.Join(vals,"&")),key...))
 result:=md5.Sum([]byte(hex.EncodeToString(digest[:])))
 return hex.EncodeToString(result[:]),nil
}
func yg109Decode(body []byte,access *ygAccess109)(map[string]any,error){
 var envelope map[string]any
 d:=json.NewDecoder(bytes.NewReader(body));d.UseNumber()
 if err:=d.Decode(&envelope);err!=nil{return nil,err}
 if access==nil{return nil,errors.New("Yeguo access unavailable")}
 if signature:=hjString(envelope["sign"]);signature!=""{
  expected,err:=yg109Signature(envelope,access.signKey)
  if err!=nil||subtle.ConstantTimeCompare([]byte(expected),[]byte(strings.ToLower(signature)))!=1{
   return nil,errors.New("Yeguo response signature mismatch")
  }
 }
 if encoded,ok:=envelope["data"].(string);ok{
  blob,err:=base64.StdEncoding.DecodeString(strings.ReplaceAll(encoded," ","+"))
  if err!=nil||len(blob)==0||len(blob)%16!=0{return nil,errors.New("invalid encrypted Yeguo payload")}
  block,err:=aes.NewCipher(access.key);if err!=nil{return nil,err}
  out:=make([]byte,len(blob))
  cipher.NewCBCDecrypter(block,access.iv).CryptBlocks(out,blob)
  out,err=legacy109Unpad(out,16);if err!=nil{return nil,err}
  d:=json.NewDecoder(bytes.NewReader(out));d.UseNumber()
  var content map[string]any
  if err:=d.Decode(&content);err!=nil{return nil,err}
  return content,nil
 }
 return envelope,nil
}
func (p *yeguo109) call(ctx context.Context,route string,q url.Values)(map[string]any,error){
 if !strings.HasPrefix(route,"/api/")||strings.Contains(route,".."){return nil,errors.New("invalid Yeguo route")}
 access,err:=p.discover(ctx);if err!=nil{return nil,err}
 values:=url.Values{
  "bundleId":{"com.pwa.mater"},"version":{"1.3.2"},"oauth_type":{"web"},
  "language":{"zh"},"via":{"pwa"},"oauth_id":{access.identifier},"trace_id":{access.identifier},"token":{""},
 }
 for key,v:=range q{values[key]=append([]string(nil),v...)}
 method:=http.MethodPost
 address:=access.base+route
 var payload []byte
 if route=="/api/home/contentOptions"{method=http.MethodGet;address+="?"+values.Encode()
 }else{payload=[]byte(values.Encode())}
 headers:=map[string]string{"Accept":"application/json","Origin":access.site,"Referer":access.site+"/","User-Agent":ua}
 if method==http.MethodPost{headers["Content-Type"]="application/x-www-form-urlencoded"}
 response,err:=p.host.Fetch(ctx,core.HTTPRequest{
  Label:"野果",Method:method,URL:address,Body:payload,Header:headers,Timeout:25*time.Second,Limit:8<<20,Direct:true,
 })
 if err!=nil{return nil,err}
 decoded,err:=yg109Decode(response.Body,access)
 if err!=nil{return nil,err}
 code:=hjString(decoded["status"])
 if code=="-1"{return nil,errors.New("Yeguo upstream authorization required")}
 if code!="1"{return nil,errors.New("Yeguo request denied by upstream")}
 data:=hjMap(decoded["data"])
 if data==nil{return nil,errors.New("invalid Yeguo data")}
 return data,nil
}
