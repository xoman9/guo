package legacy109
import(
 "bytes"
 "compress/gzip"
 "context"
 "crypto/aes"
 "crypto/rand"
 "encoding/json"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type hd109Host struct{core.Host;requests []core.HTTPRequest}
func (f *hd109Host) Fetch(_ context.Context,req core.HTTPRequest)(core.HTTPResponse,error){
 f.requests=append(f.requests,req)
 switch{
 case strings.Contains(req.URL,"/api/drama/list"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"status":"y","data":{"list":[{"id":"ab12","name":"樣本黃豆","episode_count":2}]}}`)},nil
 case strings.Contains(req.URL,"/api/drama/detail"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"status":"y","data":{"id":"ab12","name":"樣本黃豆","episode_count":2,"vip_episodes":[2],"episodes":[{"seq":1},{"seq":2}]}}`)},nil
 case strings.Contains(req.URL,"/api/drama/play"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"status":"y","data":{"is_preview":true,"url":"https://cdn.example.org/preview.mp4"}}`)},nil
 }
 return core.HTTPResponse{Status:404,Body:[]byte("{}")},nil
}
func (f *hd109Host) PrepareWebMedia(_ context.Context,m core.ProviderMedia,_ string)(core.ProviderMedia,error){return m,nil}
func TestHuangdou109CatalogAndPreventPreview(t *testing.T){
 f:=&hd109Host{}
 p:=&huangdou109{host:f}
 rows,more,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||more||len(rows)!=1||rows[0].SourceID!="ab12"{t.Fatalf("catalog %+v %t %v",rows,more,err)}
 drama,chapters,err:=p.Detail(context.Background(),"ab12")
 if err!=nil||drama.Title!="樣本黃豆"||len(chapters)!=2||!chapters[1].VIP{t.Fatalf("detail %+v %+v %v",drama,chapters,err)}
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err==nil||!strings.Contains(err.Error(),"preview"){t.Fatalf("preview was not blocked: %v",err)}
 if len(f.requests)!=3{t.Fatalf("unexpected number of HTTP requests %d",len(f.requests))}
 for _,req:=range f.requests{
  if req.Header["sign"]==""||req.Header["requestId"]==""||len(req.Body)<32{
   t.Fatal("Huangdou request missing protocol signature, IV or payload")
  }
 }
}
func TestHuangdou109EncryptedCompressedData(t *testing.T){
 rid:=hd109UUID()
 key,err:=hd109Key(rid)
 if err!=nil||len(key)!=32{t.Fatalf("invalid key: %v",err)}
 input:=[]byte(`{"status":"y","data":{"id":"ab12"}}`)
 var gz bytes.Buffer;writer:=gzip.NewWriter(&gz)
 _,_ =writer.Write(input);_ =writer.Close()
 iv:=make([]byte,aes.BlockSize);_,_=rand.Read(iv)
 cipherText,err:=hd109Encrypt(gz.Bytes(),key,iv);if err!=nil{t.Fatal(err)}
 parsed,err:=hd109Decode(append(iv,cipherText...),key)
 if err!=nil||hjString(hjMap(hd109Data(parsed))["id"])!="ab12"{t.Fatalf("decode %+v %v",parsed,err)}
 if _,err=hd109Decode([]byte("short"),key);err==nil{t.Fatal("truncated ciphertext accepted")}
 var payload any
 if json.Unmarshal(input,&payload)!=nil{t.Fatal("fixture invalid")}
}
