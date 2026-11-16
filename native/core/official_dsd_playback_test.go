package core

import (
  "context"
  "fmt"
  "strings"
  "strconv"
  "github.com/dop251/goja"
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

func TestDSDHostUTF8LengthInterop(t *testing.T) {
 p:=officialModules["dsd"].provider(&dsdOfflineHost{}).(*jsProvider)
 inst,err:=p.newInstance()
 if err!=nil {t.Fatal(err)}
 h:=p.hostObject(inst)
 if err:=inst.vm.Set("__dsdhost",h);err!=nil {t.Fatal(err)}
 v,err:=inst.vm.RunString("({kind: typeof __dsdhost.utf8Encode('中').length, total: __dsdhost.utf8Encode('中').length + 1})")
 if err!=nil {t.Fatalf("JS UTF8 length produces unexpected BigInt: %v",err)}
 o:=v.ToObject(inst.vm)
 if k:=o.Get("kind").String();k!="number" {t.Errorf("typed array length kind=%s, expected number",k)}
 if n:=o.Get("total").ToInteger();n!=4 {t.Errorf("unexpected UTF8 length+1=%d",n)}
}

type dsdSignedHost struct{ dsdOfflineHost }
func (*dsdSignedHost) FetchText(_ context.Context, r TextRequest)(TextResponse,error){
 if strings.Contains(r.URL,"/addons/vplayer/") {
   signed := "var vPath = '/signed/episode1.m3u8';"
   const marker = "(ﾟДﾟ) ['_'] ( (ﾟДﾟ) ['_']"
   const separator = "(ﾟДﾟ)[ﾟεﾟ]+"
   const tail = "+ (ﾟДﾟ)[ﾟoﾟ]"
   var encoded strings.Builder
   encoded.WriteString(marker)
   for _, ch := range signed {
     encoded.WriteString(separator)
     encoded.WriteString(strconv.FormatInt(int64(ch),8))
   }
   encoded.WriteString(tail)
   return TextResponse{URL:r.URL,Body:"<script>"+encoded.String()+"</script>"},nil
 }
 return (&dsdOfflineHost{}).FetchText(context.Background(),r)
}
func TestDSDJSDecodeSignedAAEncodingOffline(t *testing.T) {
 p:=officialModules["dsd"].provider(&dsdSignedHost{})
 task:=Task{DramaID:"dsd:123",Chapter:Chapter{ID:"dsd:123:1-1",Source:"dsd",Title:"正片"},Index:1}
 media,err:=p.ResolveMedia(context.Background(),task)
 if err!=nil{t.Fatalf("DSD signed AA playback failed: %v",err)}
 if media.URL!="https://www.dsd.com.se/signed/episode1.m3u8" {
    t.Fatalf("incorrect signed DSD URL: %q",media.URL)
 }
}

func TestDSDBigIntErrorReportsJSFrame(t *testing.T) {
 vm:=goja.New()
 _,err:=vm.RunString("function dsdPlayback() { throw new TypeError('Cannot mix BigInt and other types, use explicit conversions'); } dsdPlayback();")
 if err==nil {t.Fatal("expected JS error")}
 reported:=moduleError(context.Background(),err).Error()
 if !strings.Contains(reported,"Cannot mix BigInt") || !strings.Contains(reported,"dsdPlayback") {
   t.Fatalf("error lacks actionable JS stack location: %s",reported)
 }
}
