package core

import (
  "context"
  "errors"
  "strings"
  "testing"
)

// Embed the interface so these tests only expose the session operations they
// need, without constructing network clients or leaking cookies in output.
type fakeVerifiedHost struct {
  Host
  state map[string][]byte
  last TextRequest
  err error
}
func (h *fakeVerifiedHost) SaveState(key string, content []byte) error {
  if h.state==nil {h.state=map[string][]byte{}}
  h.state[key]=append([]byte(nil),content...)
  return nil
}
func (h *fakeVerifiedHost) LoadState(key string)([]byte,error){
 return h.state[key],nil
}
func (h *fakeVerifiedHost) FetchText(_ context.Context,request TextRequest)(TextResponse,error){
 h.last=request
 if h.err!=nil{return TextResponse{},h.err}
 return TextResponse{URL:request.URL,Body:"OK"},nil
}
func TestVerifiedSiteSessionIsFirstPartyAndPrivate(t *testing.T){
 host:=&fakeVerifiedHost{}
 if err:=SaveVerifiedSiteSession(host,"missav",VerifiedSiteSession{Cookie:"verified=yes",UserAgent:"Android WebView"});err!=nil{t.Fatal(err)}
 if got:=LoadVerifiedSiteSession(host,"missav");got.Cookie!="verified=yes"||got.UserAgent!="Android WebView"{t.Fatalf("bad first party session: %+v",got)}
 if got:=LoadVerifiedSiteSession(host,"tvbox-jable");got.Cookie!=""{t.Fatal("MissAV cookie leaked to Jable")}
 if _,err:=VerifiedSiteText(context.Background(),host,"missav","https://missav.live/cn/new","https://missav.live/cn/");err!=nil{t.Fatal(err)}
 if host.last.Cookie!="verified=yes"||host.last.UserAgent!="Android WebView"{t.Fatalf("missing session headers: %+v",host.last)}
 if _,err:=VerifiedSiteText(context.Background(),host,"missav","https://evil.example.test/","https://missav.live/");err==nil{t.Fatal("cross-site request must be rejected")}
}
func TestVerifiedSiteSessionRejectsInvalidInputAndExplains403(t *testing.T){
 h:=&fakeVerifiedHost{err:errors.New("HTTP 403")}
 for _,cookie:=range []string{"", "a=b\r\nInjected: yes",strings.Repeat("x",verifiedSiteMaxCookie+1)} {
  if err:=SaveVerifiedSiteSession(h,"tvbox-jable",VerifiedSiteSession{Cookie:cookie});err==nil{t.Fatal("invalid cookie accepted")}
 }
 if err:=SaveVerifiedSiteSession(h,"tvbox-jable",VerifiedSiteSession{Cookie:"manual=yes"});err!=nil{t.Fatal(err)}
 _,err:=VerifiedSiteText(context.Background(),h,"tvbox-jable","https://jable.tv/videos/a/","https://jable.tv/")
 if err==nil||!strings.Contains(err.Error(),"驗證"){t.Fatalf("403 must explain user verification: %v",err)}
 if err:=SaveVerifiedSiteSession(h,"unknown",VerifiedSiteSession{Cookie:"x=1"});err==nil{t.Fatal("unapproved source accepted")}
}
