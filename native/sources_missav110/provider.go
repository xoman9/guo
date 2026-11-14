package missavrecovery

// Port of the TVBox csp_Miss Java spider's category, thumbnail, search and
// detail rules to the existing Go Provider Registry. The Java spider's
// Result.get().parse() playback service is NOT present in this application.
import (
 "context"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"

 "duanjuapp/native/core"
 nethtml "golang.org/x/net/html"
)

const sourceID = "missav"
const baseURL = "https://missav.com"
var mediaSlug = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)
var categorySlug = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)
var nextPage = regexp.MustCompile(`(?is)(?:rel\s*=\s*["']next["']|aria-label\s*=\s*["']next["'])`)
var directVideoAttribute = regexp.MustCompile(`(?is)<(?:video|source)\b[^>]*\b(?:src|data-src|data-hls)\s*=\s*["']([^"']+)["']`)
var directPlayerJSON = regexp.MustCompile(`(?is)"(?:video_url|play_url|src)"\s*:\s*"([^"]+)"`)

var descriptor = core.SourceDescriptor{
 ID:sourceID,Name:"MissAV",DefaultOff:true,BaseURL:baseURL,
 NoHostRotation:true,BrowserFingerprint:true,
 Description:"依 TVBox csp_Miss Java 解析規則提供分類、影片、搜尋及詳情",
 ValidCategory:validCategory,
}
type provider struct{host core.Host}
var _ core.Provider=(*provider)(nil)
var _ core.Searcher=(*provider)(nil)
func init(){core.Register(descriptor,func(h core.Host)core.Provider{return &provider{host:h}})}
func validCategory(s string)bool{return s==""||categorySlug.MatchString(s)}
func (p *provider)Descriptor()core.SourceDescriptor{return descriptor}
func (p *provider)BaseURL()string{return baseURL}

func missAttr(n *nethtml.Node,key string)string{return strings.TrimSpace(core.ProviderHTMLAttr(n,key))}
func missNodes(root *nethtml.Node,tag,cls string)[]*nethtml.Node{
 return core.ProviderHTMLNodes(root,func(n *nethtml.Node)bool{
  return n.Type==nethtml.ElementNode&&(tag==""||n.Data==tag)&&(cls==""||core.ProviderHTMLClass(n,cls))
 })
}
func missFirst(root *nethtml.Node,tag,cls string)*nethtml.Node{
 nodes:=missNodes(root,tag,cls);if len(nodes)==0{return nil};return nodes[0]
}
func missURL(base,raw string)string{
 raw=strings.TrimSpace(raw);if raw==""{return ""}
 u,e:=url.Parse(raw);if e!=nil||u.User!=nil{return ""}
 if u.Scheme!=""&&u.Scheme!="https"&&u.Scheme!="http"{return ""}
 b,_:=url.Parse(base);absolute:=b.ResolveReference(u)
 if (absolute.Scheme!="https"&&absolute.Scheme!="http")||absolute.Hostname()==""||absolute.User!=nil{return ""}
 return absolute.String()
}
func missSitePath(raw string)string{
 resolved:=missURL(baseURL+"/",raw)
 u,e:=url.Parse(resolved);if e!=nil||!strings.EqualFold(u.Hostname(),"missav.com")||u.RawQuery!=""||u.Fragment!=""{return ""}
 path:=strings.Trim(u.Path,"/")
 if path==""{return ""}
 parts:=strings.Split(path,"/")
 if len(parts)==1&&mediaSlug.MatchString(parts[0]){return parts[0]}
 if len(parts)==2&&parts[0]=="cn"&&mediaSlug.MatchString(parts[1]){return path}
 return ""
}
func missCategoryPath(raw string)string{
 path:=missSitePath(raw)
 if path==""{return ""}
 token:=strings.TrimPrefix(path,"cn/")
 if strings.HasPrefix(strings.ToLower(token),"dm")||strings.Contains(strings.ToUpper(token),"VR"){return path}
 return ""
}
func missCategories(doc *nethtml.Node)[]core.Category{
 categories:=[]core.Category{{ID:"",Name:"全部"}}
 seen:=map[string]bool{}
 // Java: a.block.px-4.py-2.text-sm.leading-5.text-nord5.bg-nord3
 for _,a:=range missNodes(doc,"a","bg-nord3"){
  if !core.ProviderHTMLClass(a,"block"){continue}
  id:=missCategoryPath(missAttr(a,"href"))
  if id==""||seen[id]{continue}
  name:=core.ProviderHTMLText(a)
  if name==""{continue}
  seen[id]=true
  categories=append(categories,core.Category{ID:id,Name:name})
 }
 return categories
}
func missCards(doc *nethtml.Node)[]core.Drama{
 out:=[]core.Drama{}
 seen:=map[string]bool{}
 // Java: each div.thumbnail owns its own a.text-secondary and img.
 for _,box:=range missNodes(doc,"div","thumbnail"){
  a:=missFirst(box,"a","text-secondary")
  id:=missSitePath(missAttr(a,"href"))
  if id==""||seen[id]||id=="new"||id=="search"{continue}
  name:=core.ProviderHTMLText(a)
  if name==""{continue}
  cover:=""
  for _,img:=range missNodes(box,"img",""){
   pic:=missAttr(img,"data-src")
   if pic==""{pic=missAttr(img,"src")}
   if address:=missURL(baseURL+"/",pic);address!=""{cover=address;break}
  }
  seen[id]=true
  out=append(out,core.Drama{
   ID:core.ProviderDramaID(sourceID,id),Source:sourceID,SourceID:id,
   Title:name,Name:name,Cover:cover,CoverURL:cover,TotalEpisode:1,
  })
 }
 return out
}
func (p *provider)page(ctx context.Context,path string)(*nethtml.Node,error){
 if !strings.HasPrefix(path,"/"){return nil,errors.New("invalid MissAV route")}
 // FetchHostPage uses the app's browser-aware HTTP transport. Site challenges
 // remain enforced by the remote server; we do not bypass verification.
 target:=baseURL+path
 doc,_,err:=core.FetchHostPage(ctx,p.host,target,baseURL+"/",core.UserAgent)
 if err!=nil{return nil,err}
 for _,n:=range missNodes(doc,"title",""){
  title:=strings.ToLower(core.ProviderHTMLText(n))
  if strings.Contains(title,"just a moment")||strings.Contains(title,"verify you are human")||strings.Contains(title,"attention required"){
   return nil,errors.New("MissAV HTTP 403：網站要求額外驗證，目前連線未被接受")
  }
 }
 return doc,nil
}
func (p *provider)Categories(ctx context.Context,_ bool)([]core.Category,error){
 doc,err:=p.page(ctx,"/")
 if err!=nil{return nil,err}
 cats:=missCategories(doc)
 if len(cats)<2{return nil,errors.New("MissAV 未解析到 TVBox Java 分類")}
 return cats,nil
}
func hasNextPage(raw string,current int)bool{return current>0&&nextPage.MatchString(raw)}
func missNext(doc *nethtml.Node)bool{
 for _,a:=range missNodes(doc,"a",""){
  if strings.EqualFold(missAttr(a,"rel"),"next")||strings.EqualFold(missAttr(a,"aria-label"),"next"){return true}
 }
 return false
}
func (p *provider)CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if page<1||page>10000||!validCategory(category){return nil,false,errors.New("invalid MissAV page/category")}
 path:="/"
 if category!=""{path="/"+url.PathEscape(category)}
 if page>1{
  q:=url.Values{"page":{strconv.Itoa(page)}}
  path+="?"+q.Encode()
 }
 doc,err:=p.page(ctx,path)
 if err!=nil{return nil,false,err}
 return missCards(doc),missNext(doc),nil
}
func (p *provider)Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 if page<1||page>10000||len(query)>200{return nil,false,errors.New("invalid MissAV search")}
 path:="/search/"+url.PathEscape(query)+"?page="+strconv.Itoa(page)
 doc,err:=p.page(ctx,path)
 if err!=nil{return nil,false,err}
 return missCards(doc),missNext(doc),nil
}
func missMeta(doc *nethtml.Node,property string)string{
 for _,m:=range missNodes(doc,"meta",""){
  if missAttr(m,"property")==property{return missAttr(m,"content")}
 }
 return ""
}
func (p *provider)Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if missSitePath("/"+id)!=id{return core.Drama{},nil,errors.New("invalid MissAV media ID")}
 path:="/"+id
 doc,err:=p.page(ctx,path);if err!=nil{return core.Drama{},nil,err}
 name:=missMeta(doc,"og:title")
 if name==""{name=core.ProviderHTMLText(missFirst(doc,"h1",""))}
 if name==""{return core.Drama{},nil,errors.New("MissAV detail has no title")}
 pic:=missURL(baseURL+path,missMeta(doc,"og:image"))
 drama:=core.Drama{ID:core.ProviderDramaID(sourceID,id),Source:sourceID,SourceID:id,
  Title:name,Name:name,Cover:pic,CoverURL:pic,TotalEpisode:1}
 ch:=core.Chapter{ID:core.ProviderChapterID(sourceID,id,"1"),Source:sourceID,
  Title:"播放",CurrentEpisode:core.RawEpisode(1),PageURL:baseURL+path,Referer:baseURL+path}
 return drama,[]core.Chapter{ch},nil
}
func mediaPageForTask(task core.Task)(string,error){
 source,id,ok:=core.SplitProviderDramaID(task.DramaID)
 if !ok||source!=sourceID||missSitePath("/"+id)!=id{return "",errors.New("invalid MissAV drama ID")}
 if task.Chapter.ID!=core.ProviderChapterID(sourceID,id,"1"){return "",errors.New("MissAV chapter does not match drama")}
 expected:=baseURL+"/"+id
 if task.Chapter.PageURL!=expected{return "",errors.New("MissAV chapter page does not match drama")}
 return "/"+id,nil
}
// TVBox Java playerContent delegates to Result.get().parse(), a TVBox service.
// The Go app does not have that Java service. Only explicit HTTP media URLs
// exposed in the HTML can be played; no false claim of parser parity.
func directMediaURL(page,body string)(string,error){
 normalized:=strings.ReplaceAll(body,`\/`,"/")
 unique:=map[string]bool{}
 add:=func(raw string){
  u:=core.NormalizeProviderMediaURL(page,raw)
  if core.IsProviderHTTPMediaURL(u){unique[u]=true}
 }
 for _,m:=range directVideoAttribute.FindAllStringSubmatch(normalized,-1){if len(m)>1{add(m[1])}}
 for _,m:=range directPlayerJSON.FindAllStringSubmatch(normalized,-1){if len(m)>1{add(m[1])}}
 if len(unique)==0{return "",errors.New("MissAV Java 播放解析依賴 TVBox parse 服務；頁面未提供直接媒體連結")}
 if len(unique)>1{return "",errors.New("MissAV 播放頁提供多個不同媒體連結，無法安全選擇")}
 for u:=range unique{return u,nil}
 return "",errors.New("MissAV media unavailable")
}
func (p *provider)ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 path,err:=mediaPageForTask(task);if err!=nil{return core.ProviderMedia{},err}
 page:=baseURL+path
 // For media extraction retain the original HTML to detect publicly exposed
 // URLs rather than assuming TVBox's Java parsing service exists.
 resp,err:=p.host.FetchText(ctx,core.TextRequest{URL:page,Referer:baseURL+"/",UserAgent:core.UserAgent,NoCache:true})
 if err!=nil{return core.ProviderMedia{},fmt.Errorf("MissAV media page: %w",err)}
 media,err:=directMediaURL(page,resp.Body);if err!=nil{return core.ProviderMedia{},err}
 return core.ProviderMedia{URL:media,Referer:page,Credentials:core.VerifiedMediaCredentials(p.host,sourceID,page)},nil
}
