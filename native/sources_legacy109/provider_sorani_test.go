package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "net/url"
 "strings"
 "testing"

 "duanjuapp/native/core"
)

type sorani109Host struct {
 core.Host
 requests []core.HTTPRequest
 pages map[string]string
 state []byte
}
func (f *sorani109Host) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 f.requests=append(f.requests,r)
 if body,ok:=f.pages[r.URL];ok{return core.HTTPResponse{Status:200,Body:[]byte(body)},nil}
 return core.HTTPResponse{},errors.New("not mocked: "+r.URL)
}
func (f *sorani109Host) LoadState(name string)([]byte,error){
 if name!="sorani-session.json"{return nil,errors.New("unexpected state")}
 return f.state,nil
}
func TestSorani109CategoriesAndSearchRoutes(t *testing.T){
 f:=&sorani109Host{pages:map[string]string{
  sorani109API+"/video?page=1&size=24":`{"data":{"records":[{"id":123,"title":"範例動畫","episodeCount":2}],"total":1}}`,
  sorani109API+"/video/search?keyword=abc&page=1&size=24":`{"data":{"records":[{"id":123,"title":"範例動畫","episodeCount":2}],"total":1}}`,
 }}
 p:=&sorani109{host:f}
 if validSorani109Category("../bad")||!validSorani109Category("2"){t.Fatal("bad category rule")}
 rows,more,err:=p.CatalogPage(context.Background(),1,"")
 if err!=nil||more||len(rows)!=1||rows[0].SourceID!="123"{t.Fatalf("catalog: %+v %t %v",rows,more,err)}
 matches,_,err:=p.Search(context.Background(),1,"abc")
 if err!=nil||len(matches)!=1{t.Fatalf("search: %+v %v",matches,err)}
 if _,_,err=p.Search(context.Background(),1,strings.Repeat("q",257));err==nil{t.Fatal("oversized query accepted")}
}
func TestSorani109DetailAndNoCrossEpisodePlayback(t *testing.T){
 f:=&sorani109Host{pages:map[string]string{
  sorani109API+"/video/123":`{"data":{"id":123,"title":"動畫","episodes":[{"episodeId":456,"episodeOrder":1},{"episodeId":457,"episodeOrder":2}]}}`,
  sorani109API+"/video/episode/456/play?lineCode=anime_jp_m3u8":`{"data":{"canPlay":true,"url":"https://cdn.example/movie.mp4"}}`,
 }}
 p:=&sorani109{host:f}
 drama,chapters,err:=p.Detail(context.Background(),"123")
 if err!=nil||drama.SourceID!="123"||len(chapters)!=2{t.Fatalf("detail: %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||media.URL!="https://cdn.example/movie.mp4"{t.Fatalf("media: %+v %v",media,err)}
 for _,r:=range f.requests{
  if strings.Contains(r.URL,"/play")&&r.Header["Authorization"]!=""{t.Fatal("unauthorized bearer header")}
 }
 wrong:=chapters[1]
 wrong.ID=core.ProviderChapterID("sorani","999","456")
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:wrong})
 if err==nil{t.Fatal("cross-drama episode accepted")}
}
func TestSorani109HonorsAccountAuthorization(t *testing.T){
 address:=sorani109API+"/video/episode/456/play?lineCode=anime_jp_m3u8"
 f:=&sorani109Host{pages:map[string]string{
  address:`{"data":{"canPlay":false,"message":"请登录"}}`,
 }}
 p:=&sorani109{host:f}
 task:=core.Task{DramaID:core.ProviderDramaID("sorani","123"),
  Chapter:core.Chapter{ID:core.ProviderChapterID("sorani","123","456")}}
 if _,err:=p.ResolveMedia(context.Background(),task);!errors.Is(err,core.ErrSourceLogin){
  t.Fatalf("expected login required, got %v",err)
 }
 f.state=[]byte(`{"accessToken":"valid-authorized-token"}`)
 if _,err:=p.ResolveMedia(context.Background(),task);err==nil||errors.Is(err,core.ErrSourceLogin){
  t.Fatalf("denied permission should never be bypassed: %v",err)
 }
 last:=f.requests[len(f.requests)-1]
 if last.Header["Authorization"]!="Bearer valid-authorized-token"{t.Fatal("saved account token not used")}
}
func TestSorani109AuthenticationPayloadPreservesJSONNumbers(t *testing.T){
 var parsed any
 decoder:=json.NewDecoder(strings.NewReader(`{"data":{"id":12345678901234567890,"title":"Test"}}`))
 decoder.UseNumber()
 if err:=decoder.Decode(&parsed);err!=nil{t.Fatal(err)}
 if hgField(hjMap(hjPayload(parsed)),"id")!="12345678901234567890"{t.Fatal("big numeric ID corrupted")}
 if url.PathEscape("456")!="456"{t.Fatal("invalid URL path escape")}
}
