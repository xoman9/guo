package core

import (
 "context"
 "net/http"
 "testing"
)

func TestJiaoguoUsesVerifiedCatalogRoutes(t *testing.T) {
 hits:=[]string{}
 d:=sourceFixtureDownloader(t,func(r *http.Request)(*http.Response,error){
  hits=append(hits,r.URL.Path)
  if r.URL.Path!="/vodtype/1.html"{
   return sourceFixtureResponse(r,http.StatusNotFound,"wrong route"),nil
  }
  return sourceFixtureResponse(r,http.StatusOK,`<html><body><div class="module-item"><a href="/voddetail/337026.html" title="影片测试"><img src="/cover.jpg" alt="影片测试"></a></div></body></html>`),nil
 })
 items,_,err:=d.fetchApk391CatalogPage(context.Background(),"jiaoguo",1,"","")
 if err!=nil||len(items)!=1||items[0].SourceID!="337026"{
  t.Fatalf("角果取片失败: items=%+v err=%v routes=%v",items,err,hits)
 }
 if len(hits)==0||hits[0]!="/vodtype/1.html"{t.Fatalf("incorrect source route: %v",hits)}
}
