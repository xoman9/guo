package core

import (
 "context"
 "encoding/json"
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
 records:=chengguoNuxtRecords(doc)
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
  if record,ok:=records[slug];ok&&isProviderHTTPMediaURL(record.Cover){cover=record.Cover}
  seen[slug]=true
  title=truncate(title,512)
  episodes:=records[slug].Episodes
  results=append(results,Drama{ID:providerDramaID("chengguodj",slug),Source:"chengguodj",SourceID:slug,Title:title,Name:title,Cover:cover,CoverURL:cover,EpisodeCount:episodes,TotalEpisode:episodes,ChannelName:"橙果"})
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

var chengguoNuxtPattern=regexp.MustCompile(`(?s)<script[^>]+id="__NUXT_DATA__"[^>]*>(.*?)</script>`)
func chengguoNuxtRecords(doc *nethtml.Node) map[string]struct{Cover string; Episodes int} {
 result:=map[string]struct{Cover string; Episodes int}{}
 for _,script:=range providerHTMLNodes(doc,func(n *nethtml.Node)bool{return n.Data=="script"&&providerHTMLAttr(n,"id")=="__NUXT_DATA__"}) {
  var source string
  for child:=script.FirstChild;child!=nil;child=child.NextSibling{if child.Type==nethtml.TextNode{source+=child.Data}}
  var nodes []json.RawMessage
  if json.Unmarshal([]byte(source),&nodes)!=nil{continue}
  refString:=func(v json.RawMessage)string{
   var index int
   if json.Unmarshal(v,&index)!=nil||index<0||index>=len(nodes){return ""}
   var value string
   _=json.Unmarshal(nodes[index],&value)
   return value
  }
  refInt:=func(v json.RawMessage)int{
   var index int
   if json.Unmarshal(v,&index)!=nil||index<0||index>=len(nodes){return 0}
   var value int
   _=json.Unmarshal(nodes[index],&value)
   return value
  }
  for _,entry:=range nodes{
   var obj map[string]json.RawMessage
   if json.Unmarshal(entry,&obj)!=nil||obj["slug"]==nil||obj["cover"]==nil{continue}
   slug:=refString(obj["slug"])
   if !strings.HasPrefix(slug,"dj-")&&!strings.HasPrefix(slug,"app-"){continue}
   var coverID int
   if json.Unmarshal(obj["cover"],&coverID)!=nil||coverID<0||coverID>=len(nodes){continue}
   var coverFields map[string]json.RawMessage
   if json.Unmarshal(nodes[coverID],&coverFields)!=nil{continue}
   cover:=refString(coverFields["url"])
   if !isProviderHTTPMediaURL(cover){
    cover=refString(coverFields["fallback_url"])
    if strings.HasPrefix(cover,"/"){cover=apk391BaseURL("chengguodj")+cover}
   }
   result[slug]=struct{Cover string; Episodes int}{Cover:cover,Episodes:refInt(obj["published_episode_count"])}
  }
 }
 return result
}
