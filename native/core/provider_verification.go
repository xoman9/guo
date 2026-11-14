package core

// A user may complete a site's browser challenge manually in the Android
// WebView. We keep its first-party session information local to the app.
// This does not bypass a challenge or impersonate a successful verification:
// the remote website still decides whether each request is accepted.
import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net/url"
  "strings"
)

type VerifiedSiteSession struct {
 Cookie string `json:"cookie"`
 UserAgent string `json:"userAgent"`
}
const verifiedSiteMaxCookie = 8192
const verifiedSiteMaxUserAgent = 512

func verifiedSiteHost(source string) string {
 switch source {
 case "missav": return "missav.live"
 case "tvbox-jable": return "jable.tv"
 }
 return ""
}

func verifiedSiteFilename(source string) string {
 if verifiedSiteHost(source)=="" {return ""}
 return "verified-"+source+".json"
}

// SaveVerifiedSiteSession stores only first-party cookies obtained from a
// browser verification interaction. Do not print these secrets in logs.
func SaveVerifiedSiteSession(host Host, source string, session VerifiedSiteSession) error {
 file:=verifiedSiteFilename(source)
 if file=="" {return errors.New("不支援此站源的瀏覽器驗證")}
 if len(session.Cookie)>verifiedSiteMaxCookie || len(session.UserAgent)>verifiedSiteMaxUserAgent {
  return errors.New("瀏覽器驗證資訊超過限制")
 }
 if strings.ContainsAny(session.Cookie,"\r\n") || strings.ContainsAny(session.UserAgent,"\r\n") {
  return errors.New("瀏覽器驗證資訊格式無效")
 }
 if session.Cookie=="" {
  return errors.New("尚未取得網站 Cookie；請在瀏覽器完成站方要求的驗證")
 }
 bytes,err:=json.Marshal(session);if err!=nil{return errors.New("無法儲存驗證資訊")}
 return host.SaveState(file,bytes)
}
func LoadVerifiedSiteSession(host Host,source string) VerifiedSiteSession {
 file:=verifiedSiteFilename(source)
 if file=="" {return VerifiedSiteSession{}}
 body,err:=host.LoadState(file)
 if err!=nil || len(body)==0 || len(body)>verifiedSiteMaxCookie+verifiedSiteMaxUserAgent+256 {
  return VerifiedSiteSession{}
 }
 var session VerifiedSiteSession
 if json.Unmarshal(body,&session)!=nil || strings.ContainsAny(session.Cookie,"\r\n") ||
  strings.ContainsAny(session.UserAgent,"\r\n") || len(session.Cookie)>verifiedSiteMaxCookie ||
  len(session.UserAgent)>verifiedSiteMaxUserAgent {
  return VerifiedSiteSession{}
 }
 return session
}
func VerifiedSiteText(ctx context.Context,host Host,source,address,referer string)(TextResponse,error){
 domain:=verifiedSiteHost(source)
 if domain=="" {return TextResponse{},errors.New("未知的驗證站源")}
 u,err:=url.Parse(address)
 if err!=nil||u.Scheme!="https"||u.Hostname()!=domain||u.User!=nil {
  return TextResponse{},errors.New("站源請求地址不符")
 }
 session:=LoadVerifiedSiteSession(host,source)
 response,err:=host.FetchText(ctx,TextRequest{
  URL:address,Referer:referer,Cookie:session.Cookie,UserAgent:session.UserAgent,NoCache:true,
 })
 if err!=nil {
  lower:=strings.ToLower(err.Error())
  if strings.Contains(lower,"403") || strings.Contains(lower,"browser verification") ||
   strings.Contains(lower,"challenge") || strings.Contains(lower,"瀏覽器驗證") {
   return TextResponse{},fmt.Errorf("%s HTTP 403：請到站源管理，使用「驗證網站」完成站方要求的驗證，再刷新；若仍 403，可能不支援 WebView 或原生 HTTP 請求",source)
  }
 }
 return response,err
}
