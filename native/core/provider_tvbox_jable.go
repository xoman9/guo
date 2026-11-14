package core

// Native experimental Jable adapter. TVBox's "csp_Jable" references a
// Java spider class, NOT a network API. This does not execute that spider.
// It uses public Jable page metadata and supports direct HLS only.
import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"
 nethtml "golang.org/x/net/html"
)
const tvboxJableBase = "https://jable.tv"
var tvboxJableSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,100}$`)
var tvboxJableMedia = regexp.MustCompile(`(?i)(?:hlsUrl|video_url|videoUrl)\s*[:=]\s*["']([^"']+?\.m3u8(?:\?[^"']*)?)["']`)

type tvboxJableProvider struct{ host Host }
func (p *tvboxJableProvider) Descriptor() SourceDescriptor {
 return SourceDescriptor{ID:"tvbox-jable",Name:"Jable",DefaultOff:true,
  Description:"csp_Jable 的獨立原生實驗解析器（不支援執行 TVBox Java spider）",
  NoHostRotation:true,BaseURL:tvboxJableBase,
  ValidCategory:func(s string)bool{return s==""||s=="new-release"||s=="hot"}}
}
func tvboxJableURL(path string) (string,error){
 root,_:=url.Parse(tvboxJableBase)
 ref,err:=url.Parse(path);if err!=nil||ref.User!=nil{return "",errors.New("Jable 地址無效")}
 target:=root.ResolveReference(ref)
 if target.Scheme!="https"||target.Hostname()!="jable.tv" {return "",errors.New("Jable 站外地址無效")}
 return target.String(),nil
}
func (p *tvboxJableProvider) page(ctx context.Context,path string)(*nethtml.Node,string,error){
 address,err:=tvboxJableURL(path);if err!=nil{return nil,"",err}
 return FetchHostPage(ctx,p.host,address,tvboxJableBase+"/","Mozilla/5.0 (Android 11; Mobile) AppleWebKit/537.36 Chrome/131.0 Safari/537.36")
}
func (p *tvboxJableProvider) Categories(ctx context.Context,force bool)([]Category,error){
 return []Category{{ID:"new-release",Name:"最新"},{ID:"hot",Name:"熱門"}},nil
}
func tvboxJableID(address string) string {
 u,e:=url.Parse(address);if e!=nil{return ""}
 parts:=strings.Split(strings.Trim(u.Path,"/"),"/")
 if len(parts)!=2||parts[0]!="videos"||!tvboxJableSlug.MatchString(parts[1]){return ""}
 return parts[1]
}
func tvboxJableCards(doc *nethtml.Node)[]Drama{
 seen:=map[string]bool{};out:=[]Drama{}
 for _,a:=range ProviderHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
  link:=ProviderHTMLAttr(a,"href");absolute,e:=tvboxJableURL(link);if e!=nil{continue}
  id:=tvboxJableID(absolute);if id==""||seen[id]{continue}
  title:=strings.TrimSpace(ProviderHTMLAttr(a,"title"))
  cover:=""
  for _,img:=range ProviderHTMLNodes(a,func(n *nethtml.Node)bool{return n.Data=="img"}){
   if title==""{title=strings.TrimSpace(ProviderHTMLAttr(img,"alt"))}
   raw:=ProviderHTMLAttr(img,"data-src")
   if raw==""{raw=ProviderHTMLAttr(img,"src")}
   imgURL,e:=url.Parse(raw);if e==nil&&imgURL.User==nil{
    u,_:=url.Parse(absolute);imgURL=u.ResolveReference(imgURL)
    if imgURL.Scheme=="https"||imgURL.Scheme=="http"{cover=imgURL.String()}
   }
   if cover!=""{break}
  }
  if title==""||cover==""{continue}
  seen[id]=true
  out=append(out,Drama{ID:ProviderDramaID("tvbox-jable",id),Source:"tvbox-jable",SourceID:id,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"Jable"})
 }
 return out
}
func (p *tvboxJableProvider) CatalogPage(ctx context.Context,page int,category string)([]Drama,bool,error){
 if page<1||page>500{return nil,false,errors.New("無效頁碼")}
 path:="/"
 if category=="new-release"{path="/latest-updates/"}
 if category=="hot"{path="/hot/"}
 if page>1{path=strings.TrimRight(path,"/")+"/"+strconv.Itoa(page)+"/"}
 doc,_,err:=p.page(ctx,path);if err!=nil{return nil,false,err}
 items:=tvboxJableCards(doc)
 return items,len(items)>0,nil
}
func (p *tvboxJableProvider) Search(ctx context.Context,page int,query string)([]Drama,bool,error){
 if page<1||page>100||len(query)>100{return nil,false,errors.New("無效搜尋參數")}
 path:="/search/"+url.PathEscape(strings.TrimSpace(query))+"/"
 if page>1{path+=strconv.Itoa(page)+"/"}
 doc,_,err:=p.page(ctx,path);if err!=nil{return nil,false,err}
 items:=tvboxJableCards(doc);return items,len(items)>0,nil
}
func (p *tvboxJableProvider) Detail(ctx context.Context,id string)(Drama,[]Chapter,error){
 if !tvboxJableSlug.MatchString(id){return Drama{},nil,errors.New("Jable 影片 ID 無效")}
 path:="/videos/"+id+"/";doc,actual,err:=p.page(ctx,path);if err!=nil{return Drama{},nil,err}
 title:="";cover:=""
 for _,node:=range ProviderHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="meta"}){
  key:=ProviderHTMLAttr(node,"property")
  if key=="og:title"{title=ProviderHTMLAttr(node,"content")}
  if key=="og:image"{cover=ProviderHTMLAttr(node,"content")}
 }
 if title==""{for _,h:=range ProviderHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="h1"}){title=ProviderHTMLText(h);if title!=""{break}}}
 if title==""{return Drama{},nil,errors.New("Jable 詳情缺少可信影片標題")}
 if cover!="" {if u,e:=url.Parse(cover);e!=nil||u.User!=nil||u.Scheme!="https"&&u.Scheme!="http"{cover=""}}
 drama:=Drama{ID:ProviderDramaID("tvbox-jable",id),Source:"tvbox-jable",SourceID:id,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"Jable"}
 return drama,[]Chapter{{ID:ProviderChapterID("tvbox-jable",id,"1"),Source:"tvbox-jable",Title:"播放",CurrentEpisode:json.RawMessage("1"),PageURL:actual,VideoURL:actual,Referer:tvboxJableBase+"/"}},nil
}
func (p *tvboxJableProvider) ResolveMedia(ctx context.Context,task Task)(ProviderMedia,error){
 source,id,ok:=SplitProviderDramaID(task.DramaID)
 if !ok||source!="tvbox-jable"||!tvboxJableSlug.MatchString(id)||task.Chapter.ID!=ProviderChapterID(source,id,"1"){
  return ProviderMedia{},errors.New("Jable 影片分集不一致")
 }
 page,err:=tvboxJableURL("/videos/"+id+"/");if err!=nil{return ProviderMedia{},err}
 resp,err:=p.host.FetchText(ctx,TextRequest{URL:page,Referer:tvboxJableBase+"/"})
 if err!=nil{return ProviderMedia{},err}
 found:=tvboxJableMedia.FindStringSubmatch(resp.Body)
 if len(found)!=2{return ProviderMedia{},fmt.Errorf("Jable 當前頁面未提供可用 HLS；csp_Jable 專有 spider 尚未移植")}
 address:=strings.ReplaceAll(found[1],"\\/","/")
 u,e:=url.Parse(address)
 if e!=nil||u.Scheme!="https"&&u.Scheme!="http"||u.Hostname()==""||u.User!=nil{return ProviderMedia{},errors.New("Jable HLS 連結無效")}
 return p.host.PrepareWebMedia(ctx,ProviderMedia{URL:address,Referer:page},"Jable")
}
func init(){
 descriptor:=SourceDescriptor{ID:"tvbox-jable",Name:"Jable",DefaultOff:true,
  Description:"csp_Jable 相容性實驗解析器（並非 TVBox 原 spider）",NoHostRotation:true,
  BaseURL:tvboxJableBase,ValidCategory:func(s string)bool{return s==""||s=="new-release"||s=="hot"}}
 Register(descriptor,func(host Host)Provider{return &tvboxJableProvider{host:host}})
}
