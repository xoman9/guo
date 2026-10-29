package core

import (
 "bytes"
 "crypto/aes"
 "crypto/cipher"
 "encoding/json"
 "fmt"
 "strings"
 "testing"

 nethtml "golang.org/x/net/html"
)

func TestChengguoEncryptedOriginalPreferredOverBrokenFallback(t *testing.T) {
 // Nuxt's compact payload stores values by array index. The encrypted
 // original must survive SSR parsing even when fallback_url looks valid.
 nodes:=[]any{
  map[string]any{"slug":1,"cover":2,"published_episode_count":7},
  "dj-a123",
  map[string]any{"url":3,"fallback_url":4,"encrypted":5},
  "https://pic.nihmcl.cn/upload_01/cover/20261023/2026102317112641016.jpeg",
  "/static/posters/missing.jpg",
  true,
  "unused",
  10,
 }
 raw,_:=json.Marshal(nodes)
 htmlText:=fmt.Sprintf("<html><body><article data-xpch=\"card-drama\"><h3>測試短劇</h3><a href=\"/drama/dj-a123\">詳情</a><img src=\"data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///\"></article><script id=\"__NUXT_DATA__\" type=\"application/json\">%s</script></body></html>",raw)
 doc,err:=nethtml.Parse(strings.NewReader(htmlText));if err!=nil{t.Fatal(err)}
 cards:=chengguoCards(doc)
 if len(cards)!=1 {t.Fatalf("wrong card count: %+v",cards)}
 if cards[0].Cover!="https://pic.nihmcl.cn/upload_01/cover/20261023/2026102317112641016.jpeg" {
  t.Fatalf("encrypted CDN original was discarded: %q",cards[0].Cover)
 }
}
func TestChengguoAESCBCCoverRoundtrip(t *testing.T) {
 plain:=append([]byte{0xff,0xd8,0xff,0xdb},bytes.Repeat([]byte{0x12},200)...)
 padding:=aes.BlockSize-len(plain)%aes.BlockSize
 plain=append(plain,bytes.Repeat([]byte{byte(padding)},padding)...)
 encrypted:=make([]byte,len(plain))
 block,err:=aes.NewCipher([]byte("f5d965df75336270"));if err!=nil{t.Fatal(err)}
 cipher.NewCBCEncrypter(block,[]byte("97b60394abc2fbe1")).CryptBlocks(encrypted,plain)
 result:=nativeDecodeCover(encrypted)
 if !bytes.Equal(result,plain[:len(plain)-padding]){
  t.Fatal("encrypted Chengguo cover was not decoded to original JPEG bytes")
 }
}
