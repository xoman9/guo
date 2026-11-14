package legacy109
import ("context";"encoding/json";"errors";"net/url";"strings";"testing";"duanjuapp/native/core")
type jgdjMock struct{core.Host; requests []core.HTTPRequest}
func (f *jgdjMock) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 f.requests=append(f.requests,r)
 switch{
 case strings.HasSuffix(r.URL,"/auth/device"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"token":"valid-test-device-token"}`)},nil
 case strings.Contains(r.URL,"/dramas?"):
  return core.HTTPResponse{Status:200,Body:[]byte(`[{"id":123,"title":"測試短劇","episodeCount":2}]`)},nil
 case strings.HasSuffix(r.URL,"/dramas/123"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"drama":{"id":123,"title":"測試短劇"},"episodes":[{"id":456,"number":1},{"id":457,"number":2,"locked":true}]}`)},nil
 case strings.HasSuffix(r.URL,"/episodes/456/play"):
  return core.HTTPResponse{Status:200,Body:[]byte(`{"episodeId":456,"allowance":{"allowed":false},"videoUrl":"https://cdn.example/test.mp4"}`)},nil
 }
 return core.HTTPResponse{},errors.New("unmocked "+r.URL)
}
func TestJinguoduanju109AuthAndEntitlement(t *testing.T){
 p:=&jgdj109{host:&jgdjMock{}}
 rows,more,err:=p.CatalogPage(context.Background(),1,"home")
 if err!=nil||more||len(rows)!=1||rows[0].SourceID!="123"{t.Fatalf("catalog %+v %v",rows,err)}
 drama,chapters,err:=p.Detail(context.Background(),"123")
 if err!=nil||len(chapters)!=2{t.Fatalf("detail %+v %v",chapters,err)}
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err==nil||!strings.Contains(err.Error(),"purchase"){t.Fatalf("unauthorized media must be rejected: %v",err)}
 h:=p.host.(*jgdjMock); count:=0
 for _,r:=range h.requests{
  if strings.Contains(r.URL,"/auth/device"){count++;var m map[string]string
   if json.Unmarshal(r.Body,&m)!=nil||!strings.HasPrefix(m["deviceId"],"h5:"){t.Fatal("device registration invalid")}
  }
 }
 if count!=1{t.Fatalf("expected one guest handshake, got %d",count)}
 if _,err=jgdjValues("nav:../../","");err==nil{t.Fatal("invalid nav accepted")}
 if u:=url.PathEscape("456");u!="456"{t.Fatal("bad URL path")}
}
