package legacy109
import(
 "context"
 "encoding/hex"
 "encoding/json"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type bb109Fixture struct{core.Host;routes map[string]string}
func (h *bb109Fixture) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 for part,payload:=range h.routes{if strings.Contains(r.URL,part){return core.HTTPResponse{Status:200,Body:[]byte(payload)},nil}}
 return core.HTTPResponse{Status:404,Body:[]byte("{}")},nil
}
func (h *bb109Fixture) PrepareWebMedia(_ context.Context,v core.ProviderMedia,_ string)(core.ProviderMedia,error){return v,nil}
func TestBeabox109CatalogDetailAndMedia(t *testing.T){
 h:=&bb109Fixture{routes:map[string]string{
  "/api/configs":`{"code":200,"data":{"home_tags":"短劇,動畫"}}`,
  "/api/vod/latest":`{"code":200,"data":{"data":[{"vodId":"s100","vodName":"樣本","vodPic":"/img.jpg"}],"totalPages":1}}`,
  "/api/vod/detail/s100":`{"code":200,"data":{"vodId":"s100","vodName":"樣本","vodPlay":{"beabox-app":[{"flag":"第一集","url":"https://cdn.example/first.mp4"},{"flag":"第二集","url":"https://cdn.example/second.mp4"}]}}}`,
 }}
 p:=&beabox109{host:h}
 cats,err:=p.Categories(context.Background(),false)
 if err!=nil||len(cats)!=3{t.Fatalf("categories %+v %v",cats,err)}
 items,more,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||more||len(items)!=1||items[0].SourceID!="s100"{t.Fatalf("catalog %+v %t %v",items,more,err)}
 drama,chapters,err:=p.Detail(context.Background(),"s100")
 if err!=nil||len(chapters)!=2{t.Fatalf("detail %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[1]})
 if err!=nil||media.URL!="https://cdn.example/second.mp4"{t.Fatalf("media %+v %v",media,err)}
}
func TestBeabox109RejectInvalidCipherAndEpisode(t *testing.T){
 if _,err:=bbDecode109(strings.Repeat("00",32));err==nil{t.Fatal("invalid encrypted response accepted")}
 if _,err:=bbDecode109("not ciphertext");err!=nil{t.Fatal("non-encrypted field should pass")}
 if bbCategory109("tag:../../")||!bbCategory109("tag:測試"){t.Fatal("category validation failure")}
 if !bbID109("s100")||bbID109("../bad"){t.Fatal("ID validation failure")}
 _,err:=(&beabox109{host:&bb109Fixture{routes:map[string]string{}}}).ResolveMedia(context.Background(),core.Task{
  DramaID:core.ProviderDramaID("beabox","s100"),
  Chapter:core.Chapter{ID:core.ProviderChapterID("beabox","other","beabox-app:0")},
 })
 if err==nil{t.Fatal("cross-video episode accepted")}
 key,_:=hex.DecodeString(bbKey109)
 if len(key)!=16{t.Fatalf("unexpected Beabox key length %d",len(key))}
 var m any
 if json.Unmarshal([]byte(`{"code":200}`),&m)!=nil{t.Fatal("JSON fixture invalid")}
}
