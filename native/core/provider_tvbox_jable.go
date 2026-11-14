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
 "fmt"
 nethtml "golang.org/x/net/html"
)
const tvboxJableBase = "https://jable.tv"
var tvboxJableSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,100}$`)
var tvboxJableMedia = regexp.MustCompile(`(?i)(?:hlsUrl|video_url|videoUrl)\s*[:=]\s*["']([^"']+?\.m3u8(?:\?[^"']*)?)["']`)

type tvboxJableProvider struct{ host Host }
func (p *tvboxJableProvider) Descriptor() SourceDescriptor {
 return SourceDescriptor{ID:"tvbox-jable",Name:"Jable",DefaultOff:true,
  Description:"依 TVBox csp_Jable 結構重建；仍需網站驗證",
  NoHostRotation:true,BaseURL:tvboxJableBase,
  ValidCategory:func(s string)bool{return s==""||tvboxJableSlug.MatchString(s)}}
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
 response,err:=VerifiedSiteText(ctx,p.host,"tvbox-jable",address,tvboxJableBase+"/")
 if err!=nil{return nil,"",err}
 doc,e:=nethtml.Parse(strings.NewReader(response.Body))
 return doc,response.URL,e
}
// The original TVBox csp_Jable spider derives its categories from
// /categories/ div.img-box and lists from div.video-img-box. Its category
// pagination uses the site's async get_block endpoint, NOT /hot/ or
// /latest-updates/. Keep the card's id, title and image inside one card.
func tvboxJableAttr(n *nethtml.Node,key string) string {
 return strings.TrimSpace(ProviderHTMLAttr(n,key))
}
func tvboxJableClass(n *nethtml.Node,name string) bool {
 for _,token:=range strings.Fields(tvboxJableAttr(n,"class")){
  if token==name{return true}
 }
 return false
}
func tvboxJableChildren(root *nethtml.Node,tag,cls string) []*nethtml.Node {
 return ProviderHTMLNodes(root,func(n *nethtml.Node)bool{
  if tag!=""&&n.Data!=tag{return false}
  return cls==""||tvboxJableClass(n,cls)
 })
}
func tvboxJableCategoryID(link string) string {
 address,err:=tvboxJableURL(link);if err!=nil{return ""}
 u,err:=url.Parse(address);if err!=nil{return ""}
 parts:=strings.Split(strings.Trim(u.Path,"/"),"/")
 if len(parts)!=2||parts[0]!="categories"||!tvboxJableSlug.MatchString(parts[1]){return ""}
 return parts[1]
}
func tvboxJableCategories(doc *nethtml.Node) []Category {
 seen:=map[string]bool{};out:=[]Category{}
 for _,box:=range tvboxJableChildren(doc,"div","img-box"){
  id:="";for _,a:=range tvboxJableChildren(box,"a",""){
   if value:=tvboxJableCategoryID(tvboxJableAttr(a,"href"));value!=""{id=value;break}
  }
  if id==""||seen[id]{continue}
  name:=""
  for _,n:=range tvboxJableChildren(box,"h4",""){
   if tvboxJableClass(n,"absolute-center"){continue}
   name=strings.TrimSpace(ProviderHTMLText(n));if name!=""{break}
  }
  if name==""{continue}
  seen[id]=true;out=append(out,Category{ID:id,Name:name})
 }
 return out
}
func (p *tvboxJableProvider) Categories(ctx context.Context,force bool)([]Category,error){
 doc,_,err:=p.page(ctx,"/categories/")
 if err!=nil{return nil,err}
 items:=tvboxJableCategories(doc)
 if len(items)==0{return nil,errors.New("Jable 分類頁沒有解析到 TVBox 分類")}
 return items,nil
}
func tvboxJableID(address string) string {
 u,e:=url.Parse(address);if e!=nil{return ""}
 parts:=strings.Split(strings.Trim(u.Path,"/"),"/")
 if len(parts)!=2||parts[0]!="videos"||!tvboxJableSlug.MatchString(parts[1]){return ""}
 return parts[1]
}
func tvboxJableCards(doc *nethtml.Node)[]Drama{
 seen:=map[string]bool{};out:=[]Drama{}
 for _,card:=range tvboxJableChildren(doc,"div","video-img-box"){
  id:=""
  for _,a:=range tvboxJableChildren(card,"a",""){
   absolute,e:=tvboxJableURL(tvboxJableAttr(a,"href"));if e!=nil{continue}
   id=tvboxJableID(absolute)
   if id!=""{break}
  }
  if id==""||seen[id]{continue}
  title:=""
  for _,n:=range tvboxJableChildren(card,"h6",""){
   title=strings.TrimSpace(ProviderHTMLText(n));if title!=""{break}
  }
  cover:=""
  for _,img:=range tvboxJableChildren(card,"img",""){
   raw:=tvboxJableAttr(img,"data-src")
   if raw==""{raw=tvboxJableAttr(img,"src")}
   if raw==""{continue}
   imgURL,e:=url.Parse(raw);if e!=nil||imgURL.User!=nil{continue}
   root,_:=url.Parse(tvboxJableBase+"/");imgURL=root.ResolveReference(imgURL)
   if (imgURL.Scheme=="https"||imgURL.Scheme=="http")&&imgURL.Hostname()!=""&&
     !strings.HasSuffix(strings.ToLower(imgURL.Path),".gif"){cover=imgURL.String();break}
  }
  if title==""||cover==""{continue}
  seen[id]=true
  out=append(out,Drama{ID:ProviderDramaID("tvbox-jable",id),Source:"tvbox-jable",SourceID:id,
    Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"Jable"})
 }
 return out
}
func (p *tvboxJableProvider) CatalogPage(ctx context.Context,page int,category string)([]Drama,bool,error){
 if page<1||page>500{return nil,false,errors.New("無效頁碼")}
 path:="/"
 if category!="" {
  if !tvboxJableSlug.MatchString(category){return nil,false,errors.New("無效的 Jable 分類")}
  params:=url.Values{}
  params.Set("mode","async");params.Set("function","get_block")
  params.Set("block_id","list_videos_common_videos_list")
  params.Set("sort_by","post_date")
  params.Set("from",fmt.Sprintf("%02d",page))
  path="/categories/"+category+"/?"+params.Encode()
 }else if page>1{
  // The TVBox spider does not paginate its home page. Keep this
  // predictable rather than inventing a category or returning wrong films.
  return []Drama{},false,nil
 }
 doc,_,err:=p.page(ctx,path);if err!=nil{return nil,false,err}
 items:=tvboxJableCards(doc)
 return items,len(items)>=20,nil
}
func (p *tvboxJableProvider) Search(ctx context.Context,page int,query string)([]Drama,bool,error){
 if page<1||page>100||len(query)>100{return nil,false,errors.New("無效搜尋參數")}
 if strings.TrimSpace(query)==""{return []Drama{},false,nil}
 if page>1{return []Drama{},false,nil}
 path:="/search/"+url.PathEscape(strings.TrimSpace(query))+"/"
 doc,_,err:=p.page(ctx,path);if err!=nil{return nil,false,err}
 return tvboxJableCards(doc),false,nil
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
 resp,err:=VerifiedSiteText(ctx,p.host,"tvbox-jable",page,tvboxJableBase+"/")
 if err!=nil{return ProviderMedia{},err}
 found:=tvboxJableMedia.FindStringSubmatch(resp.Body)
 if len(found)!=2{return ProviderMedia{},fmt.Errorf("Jable 當前頁面未提供可用 HLS；csp_Jable 專有 spider 尚未移植")}
 address:=strings.ReplaceAll(found[1],"\\/","/")
 u,e:=url.Parse(address)
 if e!=nil||u.Scheme!="https"&&u.Scheme!="http"||u.Hostname()==""||u.User!=nil{return ProviderMedia{},errors.New("Jable HLS 連結無效")}
 return p.host.PrepareWebMedia(ctx,ProviderMedia{URL:address,Referer:page,Credentials:VerifiedMediaCredentials(p.host,"tvbox-jable",page)},"Jable")
}
func init(){
 descriptor:=SourceDescriptor{ID:"tvbox-jable",Name:"Jable",DefaultOff:true,
  Description:"依 TVBox csp_Jable 分類與列表結構移植的 Go 解析器",NoHostRotation:true,
  BaseURL:tvboxJableBase,ValidCategory:func(s string)bool{return s==""||s=="new-release"||s=="hot"}}
 Register(descriptor,func(host Host)Provider{return &tvboxJableProvider{host:host}})
}
