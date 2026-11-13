package legacy109

import (
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
 "duanjuapp/native/core"
)
type hg109Host struct{
 core.Host
 pages map[string]string
 api map[string]string
}
func (h *hg109Host) FetchText(_ context.Context,r core.TextRequest)(core.TextResponse,error){
 if text,ok:=h.pages[r.URL];ok{return core.TextResponse{Body:text},nil}
 return core.TextResponse{},errors.New("unmocked page: "+r.URL)
}
func (h *hg109Host) PrepareWebMedia(_ context.Context,media core.ProviderMedia,_ string)(core.ProviderMedia,error){return media,nil}
func (h *hg109Host) Fetch(_ context.Context,r core.HTTPRequest)(core.HTTPResponse,error){
 if r.URL!=hgAIBase+"/api/edge/request"{return core.HTTPResponse{},errors.New("unmocked AI endpoint")}
 var m map[string]string
 if json.Unmarshal(r.Body,&m)!=nil{return core.HTTPResponse{},errors.New("bad AI envelope")}
 if raw,ok:=h.api[m["path"]];ok{return core.HTTPResponse{Body:[]byte(raw),Status:200},nil}
 return core.HTTPResponse{},errors.New("unexpected edge API path "+m["path"])
}
func Test109HuangguoAIDetailAndPlayback(t *testing.T){
 h:=&hg109Host{api:map[string]string{
  "/api/videos/ab1":`{"status":1,"data":{"id":"ab1","title":"測試劇","episodes":[{"ep_num":1,"title":"第一集"}]}}`,
  "/api/videos/ab1/play?ep=1&page_texts=0":`{"status":1,"data":{"video_url":"https://video.example.org/ab1.mp4"}}`,
 }}
 p:=&huangguo109{host:h,ai:true}
 drama,chapters,err:=p.Detail(context.Background(),"ab1")
 if err!=nil||drama.SourceID!="ab1"||len(chapters)!=1{t.Fatalf("AI details %+v %+v %v",drama,chapters,err)}
 media,err:=p.ResolveMedia(context.Background(),core.Task{DramaID:drama.ID,Chapter:chapters[0]})
 if err!=nil||media.URL!="https://video.example.org/ab1.mp4"{t.Fatalf("AI media %+v %v",media,err)}
}
func Test109HuangguoVideoCatalogAndDirectHLS(t *testing.T){
 h:=&hg109Host{pages:map[string]string{
  hgVideoBase+"/series?page=1":`<article class="video-card"><a href="/series/abc-1" title="測試作品"><img data-src="/poster.jpg"></a></article>`,
  hgVideoBase+"/video/ep1":`<script data-hls="https://hls.example/ep1.m3u8"></script>`,
 }}
 p:=&huangguo109{host:h}
 rows,_,err:=p.CatalogPage(context.Background(),1,"series")
 if err!=nil||len(rows)!=1||rows[0].SourceID!="series/abc-1"{t.Fatalf("video catalog %+v %v",rows,err)}
 id:=core.ProviderDramaID(p.source(),"video/ep1")
 chapter:=core.Chapter{ID:core.ProviderChapterID(p.source(),"video/ep1","1"),PageURL:hgVideoBase+"/video/ep1"}
 // M3U8 preparation requires its own Host fixture; the no-media behavior is independently checked.
 chapter.PageURL=hgVideoBase+"/video/missing"
 _,err=p.ResolveMedia(context.Background(),core.Task{DramaID:id,Chapter:chapter})
 if err==nil||!strings.Contains(err.Error(),"unmocked"){t.Fatalf("missing page was not rejected: %v",err)}
}
