package core

import (
 "context"
 "crypto/aes"
 "crypto/cipher"
 "crypto/ecdh"
 "crypto/rand"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "strings"
 "testing"
 "time"
 "io"

 "golang.org/x/crypto/hkdf"
)

type offlineAIPlaybackHost struct {
 Host
 server *ecdh.PrivateKey
 shared []byte
 salt []byte
 requests int
 matched bool
}
func (h *offlineAIPlaybackHost) PreferredSite(string) string { return "https://huangguoai.com" }
func (h *offlineAIPlaybackHost) FinishMedia(_ context.Context, media ProviderMedia, _ string) (ProviderMedia,error) { return media,nil }
func (h *offlineAIPlaybackHost) Fetch(_ context.Context, req HTTPRequest) (HTTPResponse,error) {
 if req.Method!="POST" || req.URL!="https://huangguoai.com/api/playback/request" {
   return HTTPResponse{},fmt.Errorf("unexpected request %s %s",req.Method,req.URL)
 }
 h.requests++
 var packet struct {
   Type string `json:"type"`
   ClientPub string `json:"client_pub"`
   Kid string `json:"kid"`
   Sid string `json:"sid"`
   Seq int `json:"seq"`
   TS int64 `json:"ts"`
   Ciphertext string `json:"ciphertext"`
 }
 if err:=json.Unmarshal(req.Body,&packet);err!=nil{return HTTPResponse{},err}
 if packet.Type=="hello" {
   key,err:=ecdh.P256().GenerateKey(rand.Reader)
   if err!=nil{return HTTPResponse{},err}
   h.server=key
   clientBytes,err:=base64.RawURLEncoding.DecodeString(packet.ClientPub)
   if err!=nil{return HTTPResponse{},err}
   clientPub,err:=ecdh.P256().NewPublicKey(clientBytes)
   if err!=nil{return HTTPResponse{},err}
   h.shared,err=key.ECDH(clientPub)
   if err!=nil{return HTTPResponse{},err}
   h.salt=make([]byte,32)
   for i:=range h.salt {h.salt[i]=byte(i)}
   now:=time.Now().Unix()
   reply,_:=json.Marshal(map[string]any{
      "type":"hello","v":1,"kid":"unit-kid","sid":"unit-sid",
      "site":"huangguoai.com","server_pub":base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
      "salt":base64.RawURLEncoding.EncodeToString(h.salt),
      "server_time":now,"expires_at":now+300,
   })
   return HTTPResponse{Status:200,Body:reply},nil
 }
 if packet.Type!="call" || packet.Kid!="unit-kid" || packet.Sid!="unit-sid" || packet.Seq!=1 || h.server==nil {
   return HTTPResponse{},errors.New("bad playback session call")
 }
 info:=strings.Join([]string{"dx150-edge-v1","playback","huangguoai.com","unit-kid","unit-sid"},"\n")
 nonce:=make([]byte,12);nonce[11]=1
 derive:=func(dir string) []byte {
   key:=make([]byte,32)
   _,err:=io.ReadFull(hkdf.New(sha256.New,h.shared,h.salt,[]byte(info+"\n"+dir)),key)
   if err!=nil{panic(err)}
   return key
 }
 open:=func(key []byte, data []byte, aad string)([]byte,error){
   block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
   gcm,err:=cipher.NewGCM(block);if err!=nil{return nil,err}
   return gcm.Open(nil,nonce,data,[]byte(aad))
 }
 aad:=func(dir string)string{return strings.Join([]string{info,dir,"1",fmt.Sprint(packet.TS)},"\n")}
 bytes,err:=base64.RawURLEncoding.DecodeString(packet.Ciphertext)
 if err!=nil{return HTTPResponse{},err}
 plaintext,err:=open(derive("req"),bytes,aad("req"))
 if err!=nil{return HTTPResponse{},err}
 var in struct {
    Request struct{
      Method string `json:"method"`
      Path string `json:"path"`
      Body string `json:"body"`
      ContentType string `json:"content_type"`
    } `json:"request"`
 }
 if err:=json.Unmarshal(plaintext,&in);err!=nil{return HTTPResponse{},err}
 if in.Request.Method!="POST" || in.Request.Path!="/api/videos/abc/episodes/2/playback" ||
    in.Request.ContentType!="application/json" {
    return HTTPResponse{},fmt.Errorf("wrong playback request: %+v",in.Request)
 }
 payload,_:=base64.StdEncoding.DecodeString(in.Request.Body)
 if string(payload)!="{}" {return HTTPResponse{},fmt.Errorf("unexpected playback body %q",payload)}
 h.matched=true
 innerAPI:=`{"status":1,"data":{"streams":{"main":"https://cdn.example.com/episode2/master.m3u8?auth_key=test"}}}`
 inner,_:=json.Marshal(map[string]any{"status":200,"body":base64.StdEncoding.EncodeToString([]byte(innerAPI))})
 block,_:=aes.NewCipher(derive("res"))
 gcm,_:=cipher.NewGCM(block)
 sealed:=gcm.Seal(nil,nonce,inner,[]byte(aad("res")))
 reply,_:=json.Marshal(map[string]any{
    "type":"result","sid":"unit-sid","seq":1,
    "ciphertext":base64.RawURLEncoding.EncodeToString(sealed),
 })
 return HTTPResponse{Status:200,Body:reply},nil
}

func TestOfficialHuangguoAIEncryptedPlaybackOffline(t *testing.T) {
 host:=&offlineAIPlaybackHost{}
 module:=officialModules["huangguoai"]
 p:=module.provider(host)
 task:=Task{DramaID:"huangguoai:abc",Index:2,Chapter:Chapter{CurrentEpisode:RawEpisode(2)}}
 media,err:=p.ResolveMedia(context.Background(),task)
 if err!=nil{t.Fatalf("encrypted playback failed: %v",err)}
 want:="https://cdn.example.com/episode2/master.m3u8?auth_key=test"
 if media.URL!=want || !host.matched || host.requests!=2 {
   t.Fatalf("URL=%q, matched=%v, requests=%d",media.URL,host.matched,host.requests)
 }
}
