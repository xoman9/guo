package core

import (
 "strings"
 "testing"

 nethtml "golang.org/x/net/html"
)

func TestTVBoxJableRepairExtractsInlineHLS(t *testing.T) {
 html := `<html><head><title>Demo</title></head><body>
 <div class="video-img-box"><a href="/videos/demo-12/"><img data-src="/demo.jpg"></a><h6>Demo</h6></div>
 <script>const unused = 'hello'; var hlsUrl = "https:\/\/media.example.test\/demo.m3u8";</script>
 </body></html>`
 doc,err:=nethtml.Parse(strings.NewReader(html))
 if err!=nil{t.Fatal(err)}
 got:=tvboxJableHLSFromDocument(doc)
 want:="https://media.example.test/demo.m3u8"
 if got!=want{t.Fatalf("hlsUrl = %q, want %q",got,want)}
 cards:=tvboxJableCards(doc)
 if len(cards)!=1||cards[0].SourceID!="demo-12"||cards[0].Title!="Demo"{
  t.Fatalf("Jable repaired cards = %+v",cards)
 }
}

func TestTVBoxJableRepairNoFallbackURL(t *testing.T) {
 doc,err:=nethtml.Parse(strings.NewReader(`<html><body><script>var unrelatedUrl="https://example.test/1.m3u8";</script></body></html>`))
 if err!=nil{t.Fatal(err)}
 if got:=tvboxJableHLSFromDocument(doc);got!=""{t.Fatalf("unexpected HLS %q",got)}
}
