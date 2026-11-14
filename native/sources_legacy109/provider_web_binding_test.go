package legacy109

import (
 "context"
 "errors"
 "strings"
 "testing"

 "duanjuapp/native/core"
)

type legacy109BindingHost struct {
 core.Host
 pages map[string]string
}
func (h *legacy109BindingHost) FetchText(_ context.Context, req core.TextRequest)(core.TextResponse,error){
 if body,ok:=h.pages[req.URL];ok{return core.TextResponse{Body:body,URL:req.URL},nil}
 return core.TextResponse{},errors.New("unmocked page: "+req.URL)
}

func Test109WebProvidersRejectCrossDramaPlayback(t *testing.T) {
 tests:=[]struct{
  cfg spec
  dramaID string
  detailPath string
  detailHTML string
  playerURL string
 }{
  {
   cfg:korean,dramaID:"123",detailPath:"/hanxiaoquan/123.html",
   detailHTML:`<h1 class="page-title">示例韓劇</h1><a title="第一集" href="/play/123-1-1.html">播放</a>`,
   playerURL:"https://www.jennyhow.com/play/123-1-1.html",
  },
  {
   cfg:ghost,dramaID:"123",detailPath:"/nv/123.html",
   detailHTML:`<h1 class="media-title">示例鬼片</h1><li><a title='第一集' href='/play/101.html'>播放</a></li>`,
   playerURL:"https://guipianwu.com/play/101.html",
  },
 }
 for _,tc:=range tests{
  t.Run(tc.cfg.id,func(t *testing.T){
   host:=&legacy109BindingHost{pages:map[string]string{
    tc.cfg.base+tc.detailPath:tc.detailHTML,
    tc.playerURL:`<script>var now = "https://cdn.example.org/episode.mp4";</script>`,
    tc.cfg.base+"/play/other.html":`<script>var now = "https://cdn.example.org/wrong.mp4";</script>`,
   }}
   p:=&provider{host:host,cfg:tc.cfg}
   drama,chapters,err:=p.Detail(context.Background(),tc.dramaID)
   if err!=nil||len(chapters)!=1{t.Fatalf("fixture detail: %+v %+v %v",drama,chapters,err)}
   task:=core.Task{DramaID:drama.ID,Chapter:chapters[0]}
   media,err:=p.ResolveMedia(context.Background(),task)
   if err!=nil||media.URL!="https://cdn.example.org/episode.mp4"{
    t.Fatalf("valid episode rejected: %+v %v",media,err)
   }
   wrong:=task
   wrong.Chapter.PageURL=tc.cfg.base+"/play/other.html"
   if _,err:=p.ResolveMedia(context.Background(),wrong);err==nil||!strings.Contains(err.Error(),"not found"){
    t.Fatalf("foreign page accepted: %v",err)
   }
   wrong=task
   wrong.Chapter.ID=core.ProviderChapterID(tc.cfg.id,"999","1")
   if _,err:=p.ResolveMedia(context.Background(),wrong);err==nil{
    t.Fatal("another drama chapter accepted")
   }
   wrong=task
   wrong.DramaID=core.ProviderDramaID("different","123")
   if _,err:=p.ResolveMedia(context.Background(),wrong);err==nil{
    t.Fatal("another Provider accepted")
   }
  })
 }
}

func Test109CardCoverNeverBorrowsNextItem(t *testing.T) {
 cases:=[]struct{
  cfg spec
  html string
 }{
  {
   cfg:korean,
   html:`<a href="/hanxiaoquan/101.html" title="First" class="module-item-pic"></a>
<a href="/hanxiaoquan/102.html" title="Second" class="module-item-pic"><img data-src="/second.jpg"></a>`,
  },
  {
   cfg:ghost,
   html:`<a class="video-link" href="/nv/101.html" title="First"></a>
<a class="video-link" href="/nv/102.html" title="Second"><span data-background="/second.jpg"></span></a>`,
  },
 }
 for _,tc:=range cases{
  t.Run(tc.cfg.id,func(t *testing.T){
   p:=&provider{cfg:tc.cfg}
   cards:=p.parseCards(tc.html)
   if len(cards)!=2{t.Fatalf("expected two cards: %+v",cards)}
   if cards[0].Cover!=""{t.Fatalf("first card borrowed the second cover: %+v",cards[0])}
   if cards[1].Cover!=tc.cfg.base+"/second.jpg"{
    t.Fatalf("second cover missing/wrong: %+v",cards[1])
   }
  })
 }
}
