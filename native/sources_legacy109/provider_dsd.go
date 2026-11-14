package legacy109

import (
 "context"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "strconv"
 "strings"
 "golang.org/x/net/html"
 "duanjuapp/native/core"
)

const dsd109Base="https://www.dsd.com.se"
var dsd109Desc=core.SourceDescriptor{ID:"dsd",Name:"帝果",BaseURL:dsd109Base,NoHostRotation:true,Description:"109 MacCMS 合規直連影音路線；vplayer 簽名流程待驗證",ValidCategory:func(v string)bool{return v==""||dsd109Num(v)}}
type dsd109 struct{host core.Host}
var _ core.Provider=(*dsd109)(nil)
var _ core.Searcher=(*dsd109)(nil)
func init(){core.Register(dsd109Desc,func(h core.Host)core.Provider{return &dsd109{h}})}
func (p *dsd109) Descriptor()core.SourceDescriptor{return dsd109Desc}
func (p *dsd109) BaseURL()string{return dsd109Base}
func dsd109Num(s string)bool{
 if len(s)==0||len(s)>12||s=="0"{return false}
 for _,r:=range s{if r<'0'||r>'9'{return false}}
 return true
}
func dsd109Nodes(root *html.Node,pred func(*html.Node)bool)[]*html.Node{
 var out []*html.Node
 var visit func(*html.Node)
 visit=func(n *html.Node){
  if n==nil{return}
  if pred(n){out=append(out,n)}
  for ch:=n.FirstChild;ch!=nil;ch=ch.NextSibling{visit(ch)}
 }
 visit(root);return out
}
func dsd109Attr(n *html.Node,k string)string{
 if n==nil{return ""}
 for _,a:=range n.Attr{if a.Key==k{return strings.TrimSpace(a.Val)}}
 return ""
}
func dsd109Class(n *html.Node,c string)bool{
 for _,item:=range strings.Fields(dsd109Attr(n,"class")){if item==c{return true}}
 return false
}
func dsd109Text(n *html.Node)string{
 if n==nil{return ""}
 var out []string
 for _,child:=range dsd109Nodes(n,func(n *html.Node)bool{return n.Type==html.TextNode}){
  if value:=strings.TrimSpace(child.Data);value!=""{out=append(out,value)}
 }
 return strings.Join(out," ")
}
func dsd109Route(origin,reference string)(string,map[string]string,bool){
 base,err:=url.Parse(origin);if err!=nil{return "",nil,false}
 ref,err:=url.Parse(reference);if err!=nil||ref.User!=nil{return "",nil,false}
 ref=base.ResolveReference(ref)
 if ref.Scheme!="https"||ref.Hostname()!=base.Hostname()||ref.Port()!=""||ref.User!=nil||ref.Fragment!=""{return "",nil,false}
 parts:=strings.Split(strings.Trim(strings.TrimSuffix(strings.TrimPrefix(ref.Path,"/index.php"),".html"),"/"),"/")
 if len(parts)<2||parts[0]!="vod"||len(parts)%2!=0{return "",nil,false}
 action:=parts[1]
 switch action{case "type","show","search","play","detail":default:return "",nil,false}
 params:=map[string]string{}
 for i:=2;i+1<len(parts);i+=2{params[parts[i]]=parts[i+1]}
 for k,vs:=range ref.Query(){if len(vs)==1&&params[k]==""{params[k]=vs[0]}}
 return action,params,true
}
func (p *dsd109) page(ctx context.Context,path string)(*html.Node,string,error){
 if path!="/"&&!strings.HasPrefix(path,"/index.php/vod/"){return nil,"",errors.New("invalid DSD page")}
 response,err:=p.host.FetchText(ctx,core.TextRequest{URL:dsd109Base+path,Referer:dsd109Base+"/",UserAgent:ua})
 if err!=nil{return nil,"",err}
 doc,err:=html.Parse(strings.NewReader(response.Body))
 if err!=nil{return nil,"",err}
 return doc,dsd109Base+path,nil
}
func (p *dsd109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 doc,origin,err:=p.page(ctx,"/");if err!=nil{return nil,err}
 out:=[]core.Category{};seen:=map[string]bool{}
 for _,a:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="a"}){
  action,fields,valid:=dsd109Route(origin,dsd109Attr(a,"href"))
  id:=fields["id"];title:=dsd109Text(a)
  if !valid||action!="type"||!dsd109Num(id)||title==""||seen[id]{continue}
  seen[id]=true;out=append(out,core.Category{ID:id,Name:title})
  if len(out)>32{return nil,errors.New("too many DSD categories")}
 }
 if len(out)==0{return nil,errors.New("no DSD categories")}
 return out,nil
}
func dsd109Cards(doc *html.Node,origin string)[]core.Drama{
 seen:=map[string]bool{}
 var out []core.Drama
 for _,a:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="a"&&(dsd109Class(n,"video-item")||dsd109Class(n,"video-surface"))}){
  action,fields,valid:=dsd109Route(origin,dsd109Attr(a,"href"))
  id:=fields["id"]
  if !valid||(action!="play"&&action!="detail")||!dsd109Num(id)||seen[id]{continue}
  title:=dsd109Attr(a,"title")
  cover:=""
  for _,img:=range dsd109Nodes(a,func(n *html.Node)bool{return n.Data=="img"}){
   if title==""{title=dsd109Attr(img,"alt")}
   for _,key:=range []string{"data-src","data-original","src"}{
    if cover==""{cover=dsd109Attr(img,key)}
   }
  }
  if title==""{title=dsd109Text(a)}
  if title==""{continue}
  seen[id]=true
  out=append(out,core.Drama{
   ID:core.ProviderDramaID("dsd",id),Source:"dsd",SourceID:id,
   Title:title,Name:title,Cover:core.ResolveProviderURL(origin,cover),
  })
  if len(out)>=500{break}
 }
 return out
}
func (p *dsd109) catalog(ctx context.Context,page int,category,query string)([]core.Drama,bool,error){
 if page<1||page>1000000||len(query)>1024||(category!=""&&!dsd109Num(category)){return nil,false,errors.New("invalid DSD catalog")}
 path:=""
 if query!=""{path="/index.php/vod/search.html?"+url.Values{"wd":{query},"page":{strconv.Itoa(page)}}.Encode()
 }else{
  if category==""{return nil,false,errors.New("DSD category required")}
  path=fmt.Sprintf("/index.php/vod/type/id/%s/page/%d.html",category,page)
 }
 doc,origin,err:=p.page(ctx,path)
 if err!=nil{return nil,false,err}
 items:=dsd109Cards(doc,origin)
 more:=false
 for _,a:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="a"}){
  action,fields,ok:=dsd109Route(origin,dsd109Attr(a,"href"))
  if !ok {continue}
  if query!=""&&(action!="search"||fields["wd"]!=query){continue}
  if query==""&&(action!="type"&&action!="show"||fields["id"]!=category){continue}
  next,_:=strconv.Atoi(fields["page"])
  if next>page&&next<=1000000{more=true}
 }
 return items,more,nil
}
func (p *dsd109) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
 if category!=""{return p.catalog(ctx,page,category,"")}
 categories,err:=p.Categories(ctx,false)
 if err!=nil{return nil,false,err}
 seen:=map[string]bool{};var out []core.Drama;more:=false
 for _,cat:=range categories{
  items,hasNext,err:=p.catalog(ctx,page,cat.ID,"")
  if err!=nil{return nil,false,err}
  more=more||hasNext
  for _,item:=range items{if !seen[item.ID]{seen[item.ID]=true;out=append(out,item)}}
 }
 return out,more,nil
}
func (p *dsd109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query)
 if query==""{return nil,false,nil}
 return p.catalog(ctx,page,"",query)
}
func dsd109Player(doc *html.Node)(map[string]any,error){
 for _,script:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="script"}){
  text:=dsd109Text(script)
  at:=strings.Index(text,"player_")
  if at<0{continue}
  body:=text[at:]
  eq:=strings.Index(body,"=")
  if eq<0||eq>64{continue}
  decoder:=json.NewDecoder(strings.NewReader(body[eq+1:]))
  decoder.UseNumber()
  var result map[string]any
  if decoder.Decode(&result)==nil&&result!=nil{return result,nil}
 }
 return nil,errors.New("DSD player information missing")
}
func dsd109Page(id string,line,nid int)string{
 return fmt.Sprintf("/index.php/vod/play/id/%s/sid/%d/nid/%d.html",id,line,nid)
}
func (p *dsd109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !dsd109Num(id){return core.Drama{},nil,errors.New("invalid DSD id")}
 doc,origin,err:=p.page(ctx,dsd109Page(id,1,1))
 if err!=nil{return core.Drama{},nil,err}
 player,_:=dsd109Player(doc)
 if v:=hjString(player["id"]);v!=""&&v!=id{return core.Drama{},nil,errors.New("DSD returned wrong drama")}
 title:=hjString(hjMap(player["vod_data"])["vod_name"])
 if title==""{
  for _,h:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="h1"||n.Data=="h2"}){
   title=dsd109Text(h);if title!=""{break}
  }
 }
 if title==""{return core.Drama{},nil,errors.New("DSD detail title unavailable")}
 line:=1
 if n,ok:=hjInt(player["sid"],1000000);ok&&n>0{line=n}
 seen:=map[int]bool{};var chapters []core.Chapter
 add:=func(n int,label string){
  if n<1||seen[n]{return}
  seen[n]=true
  if label==""{label=fmt.Sprintf("第%d集",n)}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("dsd",id,fmt.Sprintf("%d-%d",line,n)),
   Source:"dsd",Title:label,CurrentEpisode:core.RawEpisode(n),
   PageURL:dsd109Base+dsd109Page(id,line,n),Referer:origin,
  })
 }
 current:=1
 if n,ok:=hjInt(player["nid"],1000000);ok&&n>0{current=n}
 add(current,"")
 for _,a:=range dsd109Nodes(doc,func(n *html.Node)bool{return n.Data=="a"}){
  action,params,ok:=dsd109Route(origin,dsd109Attr(a,"href"))
  if !ok||action!="play"||params["id"]!=id{continue}
  s,_:=strconv.Atoi(params["sid"]);n,_:=strconv.Atoi(params["nid"])
  if s==line{add(n,dsd109Text(a))}
 }
 core.SortProviderChapters(chapters)
 drama:=core.Drama{ID:core.ProviderDramaID("dsd",id),Source:"dsd",SourceID:id,Title:title,Name:title,TotalEpisode:len(chapters)}
 return drama,chapters,nil
}
func dsd109DecodePlayer(player map[string]any,origin string)(string,error){
 raw:=hjString(player["url"])
 if raw==""{return "",errors.New("DSD media not issued")}
 var err error
 switch hjString(player["encrypt"]){
 case "","0":
 case "1":raw,err=url.QueryUnescape(raw)
 case "2":
  var data []byte
  data,err=base64.StdEncoding.DecodeString(raw)
  if err==nil{raw,err=url.QueryUnescape(string(data))}
 default:return "",errors.New("unsupported player codec")
 }
 if err!=nil{return "",err}
 address:=core.ResolveProviderURL(origin,raw)
 if !core.IsProviderHTTPMediaURL(address){return "",errors.New("DSD media is not a direct supported HTTP URL")}
 return address,nil
}
func (p *dsd109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="dsd"||!dsd109Num(id){return core.ProviderMedia{},errors.New("invalid DSD drama")}
 prefix:=core.ProviderChapterID("dsd",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid DSD chapter")}
 parts:=strings.Split(strings.TrimPrefix(task.Chapter.ID,prefix),"-")
 if len(parts)!=2{return core.ProviderMedia{},errors.New("invalid DSD episode key")}
 line,_:=strconv.Atoi(parts[0]);nid,_:=strconv.Atoi(parts[1])
 if line<1||nid<1||line>1000000||nid>1000000{return core.ProviderMedia{},errors.New("invalid DSD episode")}
 expected:=dsd109Base+dsd109Page(id,line,nid)
 if task.Chapter.PageURL!=expected{return core.ProviderMedia{},errors.New("DSD chapter page does not match episode")}
 doc,origin,err:=p.page(ctx,dsd109Page(id,line,nid))
 if err!=nil{return core.ProviderMedia{},err}
 player,err:=dsd109Player(doc)
 if err!=nil{return core.ProviderMedia{},err}
 lineActual,_:=hjInt(player["sid"],1000000);nidActual,_:=hjInt(player["nid"],1000000)
 if hjString(player["id"])!=id||lineActual!=line||nidActual!=nid{return core.ProviderMedia{},errors.New("DSD returned a different episode")}
 if preview,_:=hjInt(player["trysee"],86400);preview>0{return core.ProviderMedia{},errors.New("only preview available")}
 address,err:=dsd109DecodePlayer(player,origin)
 if err!=nil{return core.ProviderMedia{},err}
 media:=core.ProviderMedia{URL:address,Referer:origin}
 if strings.Contains(strings.ToLower(address),".m3u8"){return p.host.PrepareWebMedia(ctx,media,"帝果")}
 return media,nil
}
