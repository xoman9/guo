package legacy109
import (
 "bytes"
 "crypto/aes"
 "crypto/cipher"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "errors"
 "strings"
)
// These two routines are 109 legacy API protocol transport encoding, not
// encrypted video/DRM decryption. They only decode returned API JSON.
func legacy109Pad(b []byte,size int)[]byte{
 n:=size-len(b)%size;return append(b,bytes.Repeat([]byte{byte(n)},n)...)
}
func legacy109Unpad(b []byte,size int)([]byte,error){
 if len(b)==0||len(b)%size!=0{return nil,errors.New("invalid transport padding length")}
 n:=int(b[len(b)-1])
 if n<1||n>size||n>len(b){return nil,errors.New("invalid padding")}
 for _,v:=range b[len(b)-n:]{if int(v)!=n{return nil,errors.New("invalid padding bytes")}}
 return b[:len(b)-n],nil
}
func legacy109Encrypt(v any,key,iv []byte)(string,error){
 raw,err:=json.Marshal(v);if err!=nil{return "",err}
 block,err:=aes.NewCipher(key);if err!=nil{return "",err}
 if len(iv)!=aes.BlockSize{return "",errors.New("invalid transport IV size")}
 payload:=legacy109Pad(raw,block.BlockSize())
 out:=make([]byte,len(payload))
 cipher.NewCBCEncrypter(block,iv).CryptBlocks(out,payload)
 return base64.StdEncoding.EncodeToString(out),nil
}
func legacy109Decrypt(s string,interfaceKey []byte)([]byte,error){
 encoded,err:=base64.StdEncoding.DecodeString(strings.TrimSpace(s))
 if err!=nil{return nil,err}
 if len(encoded)<12{return nil,errors.New("short API ciphertext")}
 seed:=encoded[:12]
 combined:=append(append([]byte{},interfaceKey...),seed...)
 half:=len(combined)>>1
 hash:=sha256.Sum256(combined)
 middle:=hash[8:24]
 left:=append(append([]byte{},middle...),combined[:half]...)
 right:=append(append([]byte{},combined[half:]...),middle...)
 lh:=sha256.Sum256(left);rh:=sha256.Sum256(right)
 key:=append(append(append([]byte{},lh[:8]...),rh[8:24]...),lh[24:]...)
 iv:=append(append(append([]byte{},rh[:4]...),lh[12:20]...),rh[28:]...)
 encrypted:=encoded[12:]
 block,err:=aes.NewCipher(key);if err!=nil{return nil,err}
 if len(encrypted)%block.BlockSize()!=0{return nil,errors.New("bad API ciphertext length")}
 out:=make([]byte,len(encrypted))
 cipher.NewCBCDecrypter(block,iv).CryptBlocks(out,encrypted)
 return legacy109Unpad(out,block.BlockSize())
}
func legacy109HashEnabled(v any)bool{
 if v==nil{return false}
 switch x:=v.(type){
 case bool:return x
 case string:return x!=""&&x!="false"&&x!="0"&&x!="null"
 case json.Number:return x!="0"
 }
 return true
}
