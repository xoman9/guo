package core

import (
 "context"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"

 nethtml "golang.org/x/net/html"
)

// Experimental shared adapter only for sites whose APK registry Kind=maccms.
// Specialty APIs retain explicit unsupported responses until their original
// request/response protocol is reconstructed. No source is pre-cached.
var (
 apk391IDPattern = regexp.MustCompile("(?i)/(?:voddetail|vodplay|detail|play|id)/([0-9]{1,9})(?:[-./?]|$)")
 apk391SlashID = regexp.MustCompile("(?i)/(?:vod/)?(?:detail|play)/id/([0-9]{1,9})(?:[-./?]|$)")
 apk391CatPattern=regexp.MustCompile("(?i)/(?:vodtype|vodshow|type|list)(?:/id)?[/_-]([0-9]{1,8})(?:[-./?]|$)")
 apk391Numeric=regexp.MustCompile("^[0-9]{1,9}$")
)

func apk391SameSite(source, target string) string {
 base:=apk391BaseURL(source)
 if base==""{return ""}
 b,be:=url.Parse(base);u,ue:=url.Parse(strings.TrimSpace(target))
 if be!=nil||ue!=nil||u.User!=nil{return ""}
 a:=b.ResolveReference(u)
 if a.User!=nil || (a.Scheme!="https"&&a.Scheme!="http") ||
  !strings.EqualFold(strings.TrimPrefix(a.Hostname(),"www."),strings.TrimPrefix(b.Hostname(),"www.")){return ""}
 return a.String()
}

func apk391VideoID(target string) string {
 u,err:=url.Parse(target);if err!=nil{return ""}
 for _,rx:=range []*regexp.Regexp{apk391SlashID,apk391IDPattern}{
  if m:=rx.FindStringSubmatch(u.Path);len(m)>1{return m[1]}
 }
 return ""
}

func apk391ValidCategory(source,category string)bool {
 if category==""{return true}
 for _,cat:=range apk391StaticCategoryFallback[source]{
  if cat.ID==category{return true}
 }
 if len(apk391StaticCategoryFallback[source])>0{return false}
 return apk391Numeric.MatchString(category)
}

func (d *Downloader) apk391Page(ctx context.Context,source,path string)(*nethtml.Node,string,error){
 link:=apk391SameSite(source,path)
 if link==""{return nil,"",errors.New("非法或跨站页面")}
 base:=apk391BaseURL(source)
 doc,actual,err:=d.fetchProviderPage(ctx,link,base+"/",lujj31UserAgent)
 if err!=nil{return nil,"",err}
 if len(providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="body"&&n.FirstChild!=nil}))==0{
  return nil,"",errors.New("网页未返回有效正文")
 }
 return doc,actual,nil
}

func (d *Downloader) fetchApk391Categories(ctx context.Context,source string)([]nativeCategory,error){
 spec,ok:=apk391Spec(source)
 if !ok{return nil,errors.New("未知站源")}
 // Extracted native category tables are useful even when a custom
 // website/API adapter is not yet available for live catalog playback.
 if choices:=apk391StaticCategoryFallback[source];len(choices)>0{
  return append([]nativeCategory(nil),choices...),nil
 }
 if spec.Kind!="maccms"{return nil,fmt.Errorf("%s 需要独立 %s 分类协议，尚未完成移植",spec.Name,spec.Kind)}
 doc,_,err:=d.apk391Page(ctx,source,"/")
 if err!=nil{return nil,err}
 seen:=map[string]bool{};items:=[]nativeCategory{}
 for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
  link:=apk391SameSite(source,providerHTMLAttr(a,"href"))
  if link==""{continue}
  m:=apk391CatPattern.FindStringSubmatch(link);if len(m)<2||seen[m[1]]{continue}
  name:=strings.TrimSpace(providerHTMLText(a))
  if name==""||name=="全部"||len([]rune(name))>28{continue}
  seen[m[1]]=true
  items=append(items,nativeCategory{ID:m[1],Name:name})
  if len(items)>=48{break}
 }
 if len(items)==0{return nil,errors.New("网页没有取得有效分类")}
 return items,nil
}

// A category/nav item can link to a video-shaped path, so a URL match
// alone is not evidence of a drama card. This prevents category labels from
// appearing as fake films before a detail-page refresh.
func apk391CategoryLabel(source, value string) bool {
 value = strings.Join(strings.Fields(value), "")
 if value == "" { return true }
 switch value {
 case "全部", "首页", "首頁", "分类", "分類", "更多", "最新", "最新上傳", "最新上传":
  return true
 }
 for _,category := range apk391StaticCategoryFallback[source] {
  if strings.Join(strings.Fields(category.Name),"") == value { return true }
 }
 return false
}

func apk391InsideNavigation(node *nethtml.Node) bool {
 for p:=node.Parent;p!=nil;p=p.Parent {
  switch p.Data {
  case "nav","header","footer": return true
  }
  if p.Data=="body" {break}
  for _,class:=range []string{"module-play-list-content","playlist","play-list","module-play-list","module-play-list-group","breadcrumb","module-class","module-type","category-list","type-list"} {
   if lujj31HasClass(p,class){return true}
  }
 }
 return false
}

func apk391CardTitle(anchor *nethtml.Node, source string) string {
 // Image alt/title takes priority over text from adjacent taxonomy badges.
 for _,img:=range providerHTMLNodes(anchor,func(n *nethtml.Node)bool{return n.Data=="img"}) {
  for _,attribute:=range []string{"alt","title"} {
   title:=strings.TrimSpace(providerHTMLAttr(img,attribute))
   if title!="" && !apk391CategoryLabel(source,title){return truncate(title,512)}
  }
 }
 for _,attribute:=range []string{"title","data-title"}{
  title:=strings.TrimSpace(providerHTMLAttr(anchor,attribute))
  if title!=""&&!apk391CategoryLabel(source,title){return truncate(title,512)}
 }
 for _,root:=range []*nethtml.Node{anchor,anchor.Parent} {
  if root==nil {continue}
  for _,n:=range providerHTMLNodes(root,func(n *nethtml.Node)bool{
   return n.Data=="h2"||n.Data=="h3"||n.Data=="h4"||
    lujj31HasClass(n,"module-poster-item-title")||lujj31HasClass(n,"video-title")
  }){
   title:=strings.TrimSpace(providerHTMLText(n))
   if title!=""&&!apk391CategoryLabel(source,title){return truncate(title,512)}
  }
 }
 title:=strings.TrimSpace(providerHTMLText(anchor))
 if apk391CategoryLabel(source,title){return ""}
 return truncate(title,512)
}

func apk391Cards(doc *nethtml.Node,source string)[]Drama{
 spec,ok:=apk391Spec(source);if !ok{return nil}
 seen:=map[string]bool{};items:=[]Drama{}
 for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
  if apk391InsideNavigation(a){continue}
  link:=apk391SameSite(source,providerHTMLAttr(a,"href"))
  id:=apk391VideoID(link)
  if id==""||seen[id]{continue}
  title:=apk391CardTitle(a,source)
  if title==""{continue}
  posterNode:=a
  if a.Parent!=nil{posterNode=a.Parent}
  cover:=lujj31CardPoster(posterNode,link)
  // Recovered source landing pages mix navigation, video links and categories.
  // Cards without real cover evidence frequently are category navigation.
  if cover==""{continue}
  seen[id]=true
  items=append(items,Drama{ID:providerDramaID(source,id),Source:source,SourceID:id,
   Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:spec.Name})
  if len(items)>=150{break}
 }
 return items
}

func (d *Downloader) fetchApk391CatalogPage(ctx context.Context,source string,page int,category,query string)([]Drama,bool,error){
 spec,ok:=apk391Spec(source)
 if !ok{return nil,false,errors.New("未注册的来源")}
 if page<1||page>100000||len(query)>256||!apk391ValidCategory(source,category){
  return nil,false,errors.New("非法分类或分页参数")
 }
 if spec.Kind!="maccms"{
  return nil,false,fmt.Errorf("%s 原生 %s 目录 API 尚未恢复，不会用其他来源假冒",spec.Name,spec.Kind)
 }
 routes:=[]string{}
 if query!=""{
  if page>1{return []Drama{},false,nil}
  routes=[]string{"/index.php/vod/search/wd/"+url.PathEscape(query)+".html"}
 }else if category==""{
  if page==1{routes=[]string{"/","/index.php/vod/show/id/1.html"}}
  if page>1{routes=[]string{fmt.Sprintf("/index.php/vod/show/id/1/page/%d.html",page),
   fmt.Sprintf("/vodtype/1-%d.html",page)}}
 }else{
  if page==1{
   routes=[]string{fmt.Sprintf("/index.php/vod/type/id/%s.html",category),
    fmt.Sprintf("/vodtype/%s.html",category),
    fmt.Sprintf("/index.php/vod/show/id/%s.html",category)}
  } else {
   routes=[]string{fmt.Sprintf("/index.php/vod/type/id/%s/page/%d.html",category,page),
    fmt.Sprintf("/vodtype/%s-%d.html",category,page)}
  }
 }
 var failure error
 for _,path:=range routes{
  doc,_,err:=d.apk391Page(ctx,source,path)
  if err!=nil{failure=err;continue}
  rows:=apk391Cards(doc,source)
  if len(rows)>0{return rows,len(rows)>0&&query=="",nil}
  failure=fmt.Errorf("%s 网页无影片：%s",spec.Name,path)
 }
 if failure==nil{failure=errors.New("无可用页面地址")}
 return nil,false,failure
}

func (d *Downloader) fetchApk391Detail(ctx context.Context,source,id string)(Drama,[]Chapter,error){
 spec,ok:=apk391Spec(source)
 if !ok||!apk391Numeric.MatchString(id){return Drama{},nil,errors.New("非法影片编号")}
 if spec.Kind!="maccms"{return Drama{},nil,fmt.Errorf("%s 需要专用详情接口",spec.Name)}
 paths:=[]string{fmt.Sprintf("/index.php/vod/detail/id/%s.html",id),
  fmt.Sprintf("/voddetail/%s.html",id),
  fmt.Sprintf("/vodplay/%s-1-1.html",id),
  fmt.Sprintf("/index.php/vod/play/id/%s/sid/1/nid/1.html",id)}
 var failure error
 for _,path:=range paths{
  doc,actual,err:=d.apk391Page(ctx,source,path)
  if err!=nil{failure=err;continue}
  chapters:=[]Chapter{};seen:=map[string]bool{}
  for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
   link:=apk391SameSite(source,providerHTMLAttr(a,"href"))
   if link==""||seen[link]||apk391VideoID(link)!=id{continue}
   pu,_:=url.Parse(link)
   if !strings.Contains(strings.ToLower(pu.Path),"play"){continue}
   seen[link]=true
   n:=len(chapters)+1
   title:=providerHTMLText(a);if title==""{title=fmt.Sprintf("第%d集",n)}
   chapters=append(chapters,Chapter{ID:providerChapterID(source,id,strconv.Itoa(n)),
    Source:source,Title:title,CurrentEpisode:rawEpisode(n),PageURL:link,
    VideoURL:"apk391-play://"+link,Referer:apk391BaseURL(source)+"/"})
   if len(chapters)>=500{break}
  }
  if len(chapters)==0{
   found:=false
   for _,node:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="script"}){
    for child:=node.FirstChild;child!=nil;child=child.NextSibling{
     if child.Type==nethtml.TextNode&&lujj31PlayerData.MatchString(child.Data){found=true;break}
    }
    if found{break}
   }
   if !found{failure=errors.New("未找到当前影片的分集或播放器");continue}
   chapters=[]Chapter{{ID:providerChapterID(source,id,"1"),Source:source,Title:"第1集",
    CurrentEpisode:rawEpisode(1),PageURL:actual,VideoURL:"apk391-play://"+actual,
    Referer:apk391BaseURL(source)+"/"}}
  }
  title:="影片 "+id
  for _,h:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="h1"}){
   if str:=strings.TrimSpace(providerHTMLText(h));str!=""{title=truncate(str,512);break}
  }
  cover:=lujj31DetailPoster(doc,actual)
  return Drama{ID:providerDramaID(source,id),Source:source,SourceID:id,
   Title:title,Name:title,Cover:cover,CoverURL:cover,
   EpisodeCount:len(chapters),TotalEpisode:len(chapters),ChannelName:spec.Name},chapters,nil
 }
 if failure==nil{failure=errors.New("影片详情不可用")}
 return Drama{},nil,failure
}

func (d *Downloader) resolveApk391Media(ctx context.Context,task Task)(providerMedia,error){
 source,id,ok:=splitProviderDramaID(task.DramaID)
 spec,exists:=apk391Spec(source)
 if !ok||!exists||spec.Kind!="maccms"||!apk391Numeric.MatchString(id)||
  !strings.HasPrefix(task.Chapter.ID,providerChapterID(source,id,"")){
  return providerMedia{},errors.New("无效影片来源或分集编号")
 }
 page:=apk391SameSite(source,task.Chapter.PageURL)
 if page==""||apk391VideoID(page)!=id{return providerMedia{},errors.New("分集播放页来源不一致")}
 ua:=context.WithValue(ctx,providerTextUserAgentKey{},lujj31UserAgent)
 body,err:=d.fetchProviderText(ua,page,apk391BaseURL(source)+"/")
 if err!=nil{return providerMedia{},err}
 fields:=lujj31PlayerFields(body,id)
 address,err:=lujj31PlayerURL(body,id)
 if err==nil&&isProviderHTTPMediaURL(address){
  return d.prepareWebProviderMedia(ctx,providerMedia{URL:address,Referer:page},spec.Name)
 }
 if token:=fields["url"];token!=""&&!isProviderHTTPMediaURL(token){
  parsed:=d.maccmsSiteCloudParse(ctx,apk391BaseURL(source),body,token,page)
  if isProviderHTTPMediaURL(parsed){
   return d.prepareWebProviderMedia(ctx,providerMedia{URL:parsed,Referer:page},spec.Name)
  }
 }
 if err!=nil{return providerMedia{},err}
 return providerMedia{},errors.New("站源没有解析到可播放媒体")
}
