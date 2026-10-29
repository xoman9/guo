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
func chengguoCards(doc *nethtml.Node) []Drama {
 seen:=map[string]bool{}
 results:=[]Drama{}
 for _,a:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="a"}) {
  link:=apk391SameSite("chengguodj",providerHTMLAttr(a,"href"))
  slug:=chengguoSlug(link);if slug==""||seen[slug]{continue}
  title:=strings.TrimSpace(providerHTMLAttr(a,"title"))
  cover:=""
  for p,depth:=a,0;p!=nil&&depth<3;p,depth=p.Parent,depth+1 {
   if title=="" {
    for _,h:=range providerHTMLNodes(p,func(n *nethtml.Node)bool{return n.Data=="h3"||n.Data=="h2"}) {
     if t:=strings.TrimSpace(providerHTMLText(h));t!="" {title=t;break}
    }
   }
   if cover=="" {
    for _,img:=range providerHTMLNodes(p,func(n *nethtml.Node)bool{return n.Data=="img"}) {
     raw:=firstNonEmpty(providerHTMLAttr(img,"data-src"),providerHTMLAttr(img,"src"))
     if v:=lujj31PosterAddress(raw,link);v!=""{cover=v;break}
    }
   }
   if title!=""&&cover!=""{break}
  }
  if title==""||cover==""||apk391BadTitle("chengguodj",title){continue}
  seen[slug]=true
  title=truncate(title,512)
  results=append(results,Drama{ID:providerDramaID("chengguodj",slug),Source:"chengguodj",SourceID:slug,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"橙果"})
  if len(results)>=100{break}
 }
 return results
}
func (d *Downloader) fetchChengguoCatalog(ctx context.Context,page int,category,query string)([]Drama,bool,error){
 if page<1||page>100000||len(query)>256||!apk391ValidCategory("chengguodj",category){return nil,false,errors.New("橙果分页或分类参数无效")}
 path:="/"
 if category!=""{path="/"+category}
 if query!=""{return nil,false,errors.New("橙果搜索尚未接入")}
 if page>1{return []Drama{},false,nil} // 网站分页路径尚未验证，不猜测。
 doc,_,err:=d.apk391Page(ctx,"chengguodj",path)
 if err!=nil{return nil,false,err}
 cards:=chengguoCards(doc)
 if len(cards)==0{return nil,false,fmt.Errorf("橙果页面未取得有效短剧卡片: %s",path)}
 return cards,false,nil
}
