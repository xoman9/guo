package core

import (
  "context"
  "fmt"
  "strings"
  "testing"
)

type dsdOfflineHost struct{Host}
func (*dsdOfflineHost) PreferredSite(string) string { return "https://www.dsd.com.se" }
func (*dsdOfflineHost) FetchText(_ context.Context, r TextRequest)(TextResponse,error){
  if strings.Contains(r.URL,"/addons/vplayer/"){
    return TextResponse{URL:r.URL,Body:"<html><script>var unrelated = 1;</script></html>"},nil
  }
  if !strings.Contains(r.URL,"/index.php/vod/play/id/123/sid/1/nid/1.html"){return TextResponse{},fmt.Errorf("unrecognized URL %q",r.URL)}
  return TextResponse{URL:r.URL,Body:`<html><body>
<div class="video-info"><h1>測試影片</h1></div>
<script>var player_aaaa = {"id":"123","sid":1,"nid":1,"url":"https://cdn.example.com/ok.mp4?token=123","encrypt":0,"trysee":0,"vod_data":{"vod_name":"測試影片"}};</script>
</body></html>`},nil
}
func (*dsdOfflineHost) PrepareWebMedia(_ context.Context,media ProviderMedia,_ string)(ProviderMedia,error){return media,nil}
func TestDSDJSOfflineResolveMediaWithoutBigIntTypeError(t *testing.T) {
  p:=officialModules["dsd"].provider(&dsdOfflineHost{})
  task:=Task{DramaID:"dsd:123",Chapter:Chapter{ID:"dsd:123:1-1",Source:"dsd",Title:"正片"},Index:1}
  result,err:=p.ResolveMedia(context.Background(),task)
  if err!=nil {t.Fatalf("dsd JS playback: %T: %v",err,err)}
  if result.URL!="https://cdn.example.com/ok.mp4?token=123" {t.Fatalf("wrong media URL: %q",result.URL)}
}
