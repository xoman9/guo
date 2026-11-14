package legacy109
import (
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type yg109Fixture struct{
 core.Host
 request []core.HTTPRequest
 playSortOverride int
}
func (f *yg109Fixture) FetchText(_ context.Context,r core.TextRequest)(core.TextResponse,error){
 switch r.URL{
 case yg109Site+"/":
  return core.TextResponse{Body:`<html><script type="module" src="/_nuxt/main.js"></script></html>`},nil
 case yg109Site+"/_nuxt/main.js":
  return core.TextResponse{Body:`const api="https://api.example.org/api.php";const cfg={version:"v0",mode:"CBC",padding:"Pkcs7",key:"1234567890abcdef",iv:"1234567890abcdef",sign_key:"public-sign"};`},nil
 }
 return core.TextResponse{},errors.New("unexpected Yeguo frontend route: "+r.URL)
}
func (f *yg109Fixture) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 f.request=append(f.request,r)
 var data any
 switch{
 case strings.Contains(r.URL,"/api/home/contentOptions"):
  data=map[string]any{"video_filter":map[string]any{"theme":map[string]any{"list":[]any{map[string]any{"value":1,"name":"動作"}}}}}
 case strings.Contains(r.URL,"/api/theater/exploreList"):
  data=map[string]any{"list":[]any{map[string]any{"video_id":101,"title":"作品","episode_count":2}},"page":1,"total":1}
 case strings.Contains(r.URL,"/api/playlet/detail"):
  data=map[string]any{"video_id":101,"title":"作品","episodes":[]any{map[string]any{"id":201,"sort":1},map[string]any{"id":202,"sort":2}}}
 case strings.Contains(r.URL,"/api/playlet/play"):
  sort:=1
  if f.playSortOverride>0{sort=f.playSortOverride}
  data=map[string]any{"playlet_id":101,"id":201,"episode_sort":sort,"video_url":"https://video.example/e201.mp4"}
 default:return core.HTTPResponse{},errors.New("unknown Yeguo API route: "+r.URL)
 }
 raw,_:=json.Marshal(map[string]any{"status":1,"data":data})
 return core.HTTPResponse{Status:200,Body:raw},nil
}
func (f *yg109Fixture) PrepareWebMedia(_ context.Context,v core.ProviderMedia,_ string)(core.ProviderMedia,error){return v,nil}
func TestYeguo109PublicConfigAndEpisodeParsing(t *testing.T){
 h:=&yg109Fixture{}
 p:=&yeguo109{host:h}
 cats,err:=p.Categories(context.Background(),false)
 if err!=nil||len(cats)!=1||cats[0].ID!="theme:1"{t.Fatalf("cats %+v %v",cats,err)}
 rows,more,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||more||len(rows)!=1||rows[0].SourceID!="101"{t.Fatalf("catalog %+v %v",rows,err)}
 drama,chapters,err:=p.Detail(context.Background(),"101")
 if err!=nil||len(chapters)!=2{t.Fatalf("detail %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||media.URL!="https://video.example/e201.mp4"{t.Fatalf("media %+v %v",media,err)}
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:core.Chapter{ID:core.ProviderChapterID("yeguo","102","201")}})
 if err==nil{t.Fatal("cross-episode accepted")}
}
func TestYeguo109RejectsTamperedSignedPayload(t *testing.T){
 access:=&ygAccess109{key:[]byte("1234567890abcdef"),iv:[]byte("1234567890abcdef"),signKey:[]byte("sign-key")}
 if _,err:=yg109Decode([]byte(`{"status":1,"data":{"list":[]},"sign":"tampered"}`),access);err==nil{t.Fatal("unsigned/tampered response accepted")}
 if _,_,_,err:=yg109PublicConfig(`version:"v0",mode:"CBC",padding:"Pkcs7",key:"1234567890abcdef",iv:"1234567890abcdef",sign_key:"sign-key"`);err!=nil{t.Fatal(err)}
 if yg109ValidCat("theme:../../")||!yg109ValidCat("theme:123"){t.Fatal("invalid Yeguo category rules")}
}

func TestYeguo109RejectsForeignEpisodeBeforePlayRequest(t *testing.T){
 host:=&yg109Fixture{}
 p:=&yeguo109{host:host}
 _,chapters,err:=p.Detail(context.Background(),"101")
 if err!=nil||len(chapters)!=2{t.Fatalf("fixture detail unavailable: %v",err)}
 before:=0
 for _,req:=range host.request{if strings.Contains(req.URL,"/api/playlet/play"){before++}}
 foreign:=core.Task{DramaID:core.ProviderDramaID("yeguo","101"),
  Chapter:core.Chapter{
   ID:core.ProviderChapterID("yeguo","101","999"),
   PageURL:yg109Site+"/drama/video/101/ep-3/",
  }}
 _,err=p.ResolveMedia(context.Background(),foreign)
 if err==nil||!strings.Contains(err.Error(),"does not belong"){
  t.Fatalf("foreign episode accepted: %v",err)
 }
 wrongURL:=core.Task{DramaID:core.ProviderDramaID("yeguo","101"),Chapter:chapters[0]}
 wrongURL.Chapter.PageURL=yg109Site+"/drama/video/999/ep-1/"
 if _,err=p.ResolveMedia(context.Background(),wrongURL);err==nil||!strings.Contains(err.Error(),"does not match"){
  t.Fatalf("foreign episode page accepted: %v",err)
 }
 after:=0
 for _,req:=range host.request{if strings.Contains(req.URL,"/api/playlet/play"){after++}}
 if after!=before{t.Fatal("foreign episode triggered a playback API call")}
}

func TestYeguo109RejectsWrongPlaybackOrder(t *testing.T){
 host:=&yg109Fixture{playSortOverride:2}
 p:=&yeguo109{host:host}
 drama,chapters,err:=p.Detail(context.Background(),"101")
 if err!=nil||len(chapters)==0{t.Fatalf("missing detail: %v",err)}
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err==nil||!strings.Contains(err.Error(),"order does not match"){
  t.Fatalf("wrong episode order accepted: %v",err)
 }
}
