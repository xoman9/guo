package core

import (
 "context"
 "errors"
 "fmt"
 "net/url"
 "strings"
 nethtml "golang.org/x/net/html"
)

// 橙果是短剧网站，使用 /drama/dj-xxxx 与 /play/dj-xxxx/n，
// 不采用 MacCMS 的数字 vod ID。
func chengguoSlug(link string) string {
 u,err:=url.Parse(link);if err!=nil{return ""}
 parts:=strings.Split(strings.Trim(u.Path,"/"),"/")
 if len(parts)!=2||parts[0]!="drama"{return ""}
 slug:=parts[1]
 if !strings.HasPrefix(slug,"dj-")||len(slug)>48{return ""}
 for _,r:=range slug[3:]{
  if !(r>='a'&&r<='z'||r>='0'&&r<='9'||r=='-'){return ""}
 }
 return slug
}
// The /browse SSR output uses <article data-xpch="card-drama">. Its image
// often stays a transparent lazy-loading GIF, while title and /drama link
// are separate siblings. Never discard a verified film for missing cover.
func chengguoCards(doc *nethtml.Node) []Drama {
 seen:=map[string]bool{}
 results:=[]Drama{}
 for _,article:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{
  return n.Data=="article"&&providerHTMLAttr(n,"data-xpch")=="card-drama"
 }){
  slug:=""
  for _,a:=range providerHTMLNodes(article,func(n *nethtml.Node)bool{return n.Data=="a"}){
   link:=apk391SameSite("chengguodj",providerHTMLAttr(a,"href"))
   if id:=chengguoSlug(link);id!=""{slug=id;break}
  }
  if slug==""||seen[slug]{continue}
  title:=""
  for _,h:=range providerHTMLNodes(article,func(n *nethtml.Node)bool{return n.Data=="h3"}){
   if text:=strings.TrimSpace(providerHTMLText(h));text!=""{title=text;break}
  }
  if title==""||apk391BadTitle("chengguodj",title){continue}
  cover:=""
  for _,img:=range providerHTMLNodes(article,func(n *nethtml.Node)bool{return n.Data=="img"}){
   raw:=firstNonEmpty(providerHTMLAttr(img,"data-src"),providerHTMLAttr(img,"data-original"),providerHTMLAttr(img,"src"))
   if strings.HasPrefix(raw,"data:")||raw==""{continue}
   if resolved:=lujj31PosterAddress(raw,apk391BaseURL("chengguodj")+"/");resolved!=""{cover=resolved;break}
  }
  seen[slug]=true
  title=truncate(title,512)
  results=append(results,Drama{ID:providerDramaID("chengguodj",slug),Source:"chengguodj",SourceID:slug,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"橙果"})
  if len(results)>=100{break}
 }
 return results
}
func chengguoHasNext(doc *nethtml.Node,page int)bool {
 want:=fmt.Sprintf("/browse/page-%d",page+1)
 for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}){
  if u:=apk391SameSite("chengguodj",providerHTMLAttr(a,"href"));u!=""{
   parsed,err:=url.Parse(u);if err==nil&&strings.TrimRight(parsed.Path,"/")==want{return true}
  }
 }
 return false
}
func (d *Downloader) fetchChengguoCatalog(ctx context.Context,page int,category,query string)([]Drama,bool,error){
 if page<1||page>100000||len(query)>256||!apk391ValidCategory("chengguodj",category){return nil,false,errors.New("橙果分页或分类参数无效")}
 path:="/browse"
 if category!=""{path="/"+category}
 if query!=""{return nil,false,errors.New("橙果搜索尚未接入")}
 if page>1 && category==""{path=fmt.Sprintf("/browse/page-%d",page)}
 if page>1 && category!=""{return []Drama{},false,nil} // Category paging not yet verified.
 doc,_,err:=d.apk391Page(ctx,"chengguodj",path)
 if err!=nil{return nil,false,err}
 cards:=chengguoCards(doc)
 if len(cards)==0{return nil,false,fmt.Errorf("橙果页面未取得有效短剧卡片: %s",path)}
 return cards,category==""&&chengguoHasNext(doc,page),nil
}
