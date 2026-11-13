package legacy109

import (
 "context"
 "encoding/json"
 "net/http"
 "strings"
 "testing"

 "duanjuapp/native/core"
)

func Test109HuangjuCategoryAndMetadata(t *testing.T){
 for _,v:=range []string{"","@new","drama","short-show"}{
  if !validHuangju109Category(v){t.Fatalf("valid category rejected: %s",v)}
 }
 for _,v:=range []string{"../escape","url/a","foo?bar","a b"}{
  if validHuangju109Category(v){t.Fatalf("invalid category accepted: %s",v)}
 }
 value:=map[string]any{"id":"101","slug":"sample","title":"範例","coverUrl":"/image.jpg","totalEpisodes":float64(8)}
 item,err:=hjDrama(value,"sample-101")
 if err!=nil||item.SourceID!="sample-101"||item.Source!="huangju"||item.Title!="範例"{
  t.Fatalf("invalid drama map: %+v %v",item,err)
 }
 if _,err:=hjDrama(value,"unexpected-101");err==nil {t.Fatal("detail did not check drama identity")}
}
func Test109HuangjuSignedMediaRequiresAllCookies(t *testing.T){
 url:="https://media.example/video.mp4"
 headers:=map[string][]string{"Set-Cookie":{
  "CloudFront-Policy=policy1; Path=/; Secure",
  "CloudFront-Signature=signature1; Path=/",
  "CloudFront-Key-Pair-Id=key1; Path=/",
 }}
 credentials,err:=hjCredentials(headers,url)
 if err!=nil||credentials==nil||credentials.Origin!="https://media.example"{
  t.Fatalf("credentials error %+v %v",credentials,err)
 }
 if !strings.Contains(credentials.Cookie,"CloudFront-Policy=policy1") {t.Fatal("cookie missing")}
 delete(headers,"Set-Cookie")
 if _,err=hjCredentials(headers,url);err==nil{t.Fatal("missing media cookies accepted")}
}
type fakeHuangjuHost struct{
 core.Host
 calls []core.HTTPRequest
}
func (f *fakeHuangjuHost) Fetch(_ context.Context,request core.HTTPRequest)(core.HTTPResponse,error){
 f.calls=append(f.calls,request)
 if request.URL=="https://api.huangju.net/auth/guest"{
  if request.Method!=http.MethodPost||!strings.Contains(string(request.Body),"deviceId"){
   return core.HTTPResponse{Status:400},nil
  }
  return core.HTTPResponse{Status:200,Body:[]byte(`{"data":{"token":"guest-token"}}`)},nil
 }
 if request.Header["Authorization"]!="Bearer guest-token"{
  return core.HTTPResponse{Status:403},nil
 }
 if strings.HasPrefix(request.URL,"https://api.huangju.net/dramas?page=1"){
  return core.HTTPResponse{Status:200,Body:[]byte(`{"data":{"items":[{"id":"101","slug":"sample","title":"範例","totalEpisodes":2}],"pageSize":20,"total":1}}`)},nil
 }
 if request.URL=="https://api.huangju.net/dramas/sample-101"{
  return core.HTTPResponse{Status:200,Body:[]byte(`{"data":{"id":"101","slug":"sample","title":"範例","totalEpisodes":2,"episodes":[{"id":"ep1","epNo":1,"playable":true},{"id":"ep2","epNo":2,"playable":true}]}}`)},nil
 }
 if request.URL=="https://api.huangju.net/play/ep1"{
  return core.HTTPResponse{
   Status:200,Body:[]byte(`{"data":{"url":"https://media.example/episode.mp4"}}`),
   Header:map[string][]string{"Set-Cookie":{
    "CloudFront-Policy=policy1; Path=/",
    "CloudFront-Signature=signature1; Path=/",
    "CloudFront-Key-Pair-Id=key1; Path=/",
   }},
  },nil
 }
 return core.HTTPResponse{Status:404,Body:[]byte("{}")},nil
}
func Test109HuangjuGuestCatalogDetailAndSignedPlayback(t *testing.T){
 host:=&fakeHuangjuHost{}
 provider:=&huangju109{host:host}
 ctx:=context.Background()
 rows,more,err:=provider.CatalogPage(ctx,1,"")
 if err!=nil||more||len(rows)!=1{t.Fatalf("catalog: %+v %t %v",rows,more,err)}
 detail,chapters,err:=provider.Detail(ctx,"sample-101")
 if err!=nil||detail.SourceID!="sample-101"||len(chapters)!=2{
  t.Fatalf("detail: %+v %+v %v",detail,chapters,err)
 }
 media,err:=provider.ResolveMedia(ctx,core.Task{
  DramaID:core.ProviderDramaID("huangju","sample-101"),
  Chapter:chapters[0],
 })
 if err!=nil||media.Credentials==nil||media.URL!="https://media.example/episode.mp4"{
  t.Fatalf("playback: %+v %v",media,err)
 }
 guestCalls:=0
 for _,c:=range host.calls{if strings.Contains(c.URL,"/auth/guest"){guestCalls++}}
 if guestCalls!=1{t.Fatalf("guest auth must be reused; calls=%d",guestCalls)}
 // Ensure the signed media credentials are not serialized as a static URL.
 raw,_:=json.Marshal(media.URL)
 if strings.Contains(string(raw),"CloudFront-"){t.Fatal("signed cookie leaked into URL")}
}
