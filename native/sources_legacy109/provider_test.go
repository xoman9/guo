package legacy109

import "testing"

func TestRecover109HanxiaoquanCatalogCards(t *testing.T){
 provider:=&provider{cfg:korean}
 body:=`<a href="/hanxiaoquan/333.html" title="Sample" class="module-item-pic"><img data-src="/a.jpg"></a>`
 rows:=provider.parseCards(body)
 if len(rows)!=1 || rows[0].SourceID!="333" || rows[0].Title!="Sample"{
  t.Fatalf("bad 109 Korean card parsing: %+v",rows)
 }
}
func TestRecover109GuipianCatalogCards(t *testing.T){
 provider:=&provider{cfg:ghost}
 body:=`<a class="video-link" href="/nv/222.html" title="Sample" data-original="/cover.jpg"></a>`
 rows:=provider.parseCards(body)
 if len(rows)!=1||rows[0].SourceID!="222"{t.Fatalf("bad 109 Ghost card parsing: %+v",rows)}
}
