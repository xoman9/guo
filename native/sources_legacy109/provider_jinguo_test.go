package legacy109
import(
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type jg109Fixture struct{core.Host; calls []core.HTTPRequest}
func (*jg109Fixture) LoadState(string)([]byte,error){return nil,nil}
func (*jg109Fixture) SaveState(string,[]byte)error{return nil}
func (f *jg109Fixture) Fetch(_ context.Context,req core.HTTPRequest)(core.HTTPResponse,error){
 f.calls=append(f.calls,req)
 var data any
 switch{
 case strings.Contains(req.URL,"/mine/login/h5"):
  if req.Method!="POST"||!strings.Contains(string(req.Body),`"data"`){return core.HTTPResponse{},errors.New("Jinguo login must encrypt parameters")}
  data=map[string]string{"token":"guest-token"}
 case strings.Contains(req.URL,"/playlet/detail/A1"):
  data=map[string]any{"id":"A1","title":"測試禁果","chapters":[]any{
   map[string]any{"id":"C1","videoUrl":"playlist/A1/ep1"},
   map[string]any{"id":"C2","videoUrl":""},
  }}
 case strings.Contains(req.URL,"/playlet-recommend/home/choose"):
  data=map[string]any{"list":[]any{map[string]any{"id":"A1","title":"測試禁果"}}}
 default:return core.HTTPResponse{},errors.New("unmocked request "+req.URL)
 }
 body,_:=json.Marshal(map[string]any{"code":200,"hash":false,"data":data})
 return core.HTTPResponse{Status:200,Body:body},nil
}
func (f *jg109Fixture) FetchPlaylist(_ context.Context,address,referer string,_ *core.MediaCredentials)(string,string,error){
 if !strings.Contains(address,"/api/app/vid/h5/m3u8/playlist/A1/ep1?")||!strings.Contains(address,"token=guest-token"){
  return "","",errors.New("invalid authenticated playlist request")
 }
 return "#EXTM3U
#EXTINF:4.0,
seg1.ts
",address,nil
}
func TestJinguo109EncryptedGuestCatalogAndPlayback(t *testing.T){
 h:=&jg109Fixture{}
 p:=&jinguo109{host:h}
 rows,_,err:=p.CatalogPage(context.Background(),1,"choose")
 if err!=nil||len(rows)!=1||rows[0].SourceID!="A1"{t.Fatalf("catalog %+v %v",rows,err)}
 drama,chapters,err:=p.Detail(context.Background(),"A1")
 if err!=nil||len(chapters)!=2{t.Fatalf("detail %+v %+v",chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||!strings.HasPrefix(media.Playlist,"#EXTM3U"){t.Fatalf("authorized HLS %+v %v",media,err)}
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[1]})
 if err==nil||!strings.Contains(err.Error(),"unlock"){t.Fatalf("locked episode accepted %v",err)}
 count:=0;for _,r:=range h.calls{if strings.Contains(r.URL,"/mine/login/h5"){count++}}
 if count!=1{t.Fatalf("repeated guest auth %d",count)}
}
func TestJinguo109TransportCryptoAndInputValidation(t *testing.T){
 s,err:=legacy109Encrypt(map[string]string{"keyword":"test"},[]byte(jgParamKey),[]byte(jgParamKey))
 if err!=nil||s==""{t.Fatalf("encryption %s %v",s,err)}
 if _,err:=legacy109Decrypt("invalid",[]byte(jgInterfaceKey));err==nil{t.Fatal("invalid response accepted")}
 if !jgValidCat("tab:A1")||jgValidCat("tab:../"){t.Fatal("invalid category accepted")}
}
