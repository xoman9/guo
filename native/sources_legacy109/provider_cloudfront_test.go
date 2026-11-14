package legacy109
import (
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type cf109Host struct{core.Host; state []byte;request []core.HTTPRequest}
func (h *cf109Host) LoadState(name string)([]byte,error){
 if name!="huangguo-session.json"{return nil,errors.New("unexpected session")}
 return h.state,nil
}
func (h *cf109Host) SaveState(name string,content []byte)error{h.state=append([]byte(nil),content...);return nil}
func (h *cf109Host) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 h.request=append(h.request,r)
 var data any
 switch{
 case strings.Contains(r.URL,"/api/app/playlet-tab/all"):
  data=[]any{map[string]string{"id":"1","name":"最新"}}
 case strings.Contains(r.URL,"/api/app/playlet/home/tab/1"):
  data=map[string]any{"list":[]any{map[string]string{"id":"A123","title":"影片一"}}}
 case strings.Contains(r.URL,"/api/app/playlet/detail/A123"):
  data=map[string]any{"id":"A123","title":"影片一","chapters":[]any{map[string]any{"id":"E1","videoUrl":"files/A123/1"}}}
 default:return core.HTTPResponse{},errors.New("unmocked CloudFront API: "+r.URL)
 }
 b,_:=json.Marshal(map[string]any{"code":200,"data":data,"hash":false})
 return core.HTTPResponse{Status:200,Body:b},nil
}
func (h *cf109Host) FetchPlaylist(_ context.Context,address,_ string,_ *core.MediaCredentials)(string,string,error){
 if !strings.Contains(address,"/api/app/vid/h5/m3u8/files%2FA123%2F1?")||!strings.Contains(address,"token=authorized-fixture"){
  return "","",errors.New("wrong CloudFront playback URI")
 }
 return "#EXTM3U\n",address,nil
}
func TestCloudFront109AuthorizedSessionAndCatalog(t *testing.T){
 proto:=cfProtocol109{InterfaceKey:"0123456789abcdef",ParamKey:"1234567890abcdef",ParamIV:"1234567890abcdef"}
 state,_:=json.Marshal(map[string]any{"protocol":proto,"token":"authorized-fixture","deviceId":"1234"})
 h:=&cf109Host{state:state}
 p:=&cf109Provider{host:h}
 categories,err:=p.Categories(context.Background(),false)
 if err!=nil||len(categories)!=1{t.Fatalf("categories %+v %v",categories,err)}
 rows,_,err:=p.CatalogPage(context.Background(),1,"1")
 if err!=nil||len(rows)!=1{t.Fatalf("catalog %+v %v",rows,err)}
 drama,chapters,err:=p.Detail(context.Background(),"A123")
 if err!=nil||len(chapters)!=1{t.Fatalf("detail %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||!strings.HasPrefix(media.Playlist,"#EXTM3U"){t.Fatalf("legacy media %+v %v",media,err)}
 for _,r:=range h.request{
  if r.Header["Authorization"]!="authorized-fixture"{t.Fatal("authorized token not forwarded")}
 }
}
func TestCloudFront109RejectsMissingProtocol(t *testing.T){
 if _,err:=cfExtractProtocol109("not a real script");err==nil{t.Fatal("fabricated protocol accepted")}
 found,err:=cfExtractProtocol109(`const p={"interfaceKey":"0123456789abcdef","parameterKey":"1234567890abcdef","parameterIv":"1234567890abcdef"}`)
 if err!=nil||!cfValidProtocol109(found){t.Fatalf("public protocol parsing failed: %+v %v",found,err)}
}
