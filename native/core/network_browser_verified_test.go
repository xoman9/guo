package core

import (
 "net/http"
 "testing"
)

func TestMissAVBrowserTransportPreservesApprovedWebViewUA(t *testing.T) {
 req,err:=http.NewRequest("GET","https://missav.ws/cn/test-123",nil)
 if err!=nil{t.Fatal(err)}
 req.Header.Set("Cookie","site_session=fixture")
 req.Header.Set("User-Agent","Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/135.0.0.0 Mobile Safari/537.36")
 headers:=huangguoBrowserHeaders(req)
 if got:=headers["user-agent"];len(got)!=1||got[0]!=req.Header.Get("User-Agent"){
  t.Fatalf("verified WebView UA was overwritten: %q",got)
 }
 for _,key:=range []string{"sec-ch-ua","sec-ch-ua-platform","sec-ch-ua-mobile"}{
  if _,exists:=headers[key];exists{t.Fatalf("contradictory desktop hint: %s",key)}
 }
 req2,err:=http.NewRequest("GET","https://other.example.test/",nil)
 if err!=nil{t.Fatal(err)}
 req2.Header.Set("Cookie","session=fixture")
 req2.Header.Set("User-Agent","test")
 defaultHeaders:=huangguoBrowserHeaders(req2)
 if defaultHeaders["user-agent"][0]=="test" {t.Fatal("other source profile should remain untouched")}
}
