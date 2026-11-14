package core

import (
 "strings"
 "testing"
 nethtml "golang.org/x/net/html"
)

func TestTVBoxType1RegistryMetadata(t *testing.T) {
 expected:=map[string]string{
  "tvbox-laoguo":"老果", "tvbox-155":"一五五", "tvbox-jkun":"jkun",
  "tvbox-senlin":"森林", "tvbox-jingpinx":"精品下",
 }
 for _,site:=range tvboxType1Specs {
  if expected[site.id]!=site.name{t.Errorf("mismatched name for %s: %s",site.id,site.name)}
  if !strings.Contains(site.endpoint,"/api.php/provide/vod/"){t.Errorf("invalid API endpoint %s",site.endpoint)}
  d,ok:=registeredDescriptor(site.id)
  if !ok||!d.DefaultOff{t.Errorf("optional TVBox source %s must be registered default-off",site.id)}
  delete(expected,site.id)
 }
 if len(expected)!=0{t.Fatalf("unregistered sites: %v",expected)}
}
func TestTVBoxType1IdentityAndPoster(t *testing.T){
 p:=&tvboxType1Provider{spec:tvboxType1Spec{id:"tvbox-155",name:"一五五"}}
 a,ok:=p.drama(map[string]any{"vod_id":float64(123),"vod_name":"影片 A","vod_pic":"https://images.example.test/123.jpg"})
 if !ok||a.ID!="tvbox-155:123"||a.Title!="影片 A"||a.Cover!="https://images.example.test/123.jpg" {t.Fatalf("bad video: %+v",a)}
 b,ok:=p.drama(map[string]any{"vod_id":"124","vod_name":"影片 B","vod_pic":"javascript:alert(1)"})
 if !ok||b.Cover!=""||b.ID==a.ID{t.Fatalf("unsafe or mixed poster: %+v",b)}
}
func TestTVBoxJableScopedCards(t *testing.T){
 html:=`<div><a href="/videos/video-one/" title="第一片"><img data-src="https://img.example.test/one.jpg"></a><a href="/videos/video-two/" title="第二片"><img src="https://img.example.test/two.jpg"></a></div>`
 doc,e:=nethtml.Parse(strings.NewReader(html));if e!=nil{t.Fatal(e)}
 items:=tvboxJableCards(doc)
 if len(items)!=2||items[0].SourceID!="video-one"||items[1].SourceID!="video-two"||items[0].Cover==items[1].Cover{
  t.Fatalf("Jable cover/title must stay within a video anchor: %+v",items)
 }
}
