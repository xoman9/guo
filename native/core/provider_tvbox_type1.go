package core

// Standard TVBox type=1 MacCMS JSON providers. A type=3 csp_* spider is
// not a URL and cannot be treated as a MacCMS interface.
import (
 "context"
 "encoding/json"
 "encoding/xml"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"
)

type tvboxType1Spec struct{ id, name, endpoint string; xml bool }
var tvboxType1Specs = []tvboxType1Spec{
 {"tvbox-laoguo", "老果", "https://apilsbzy1.com/api.php/provide/vod/?ac=list", false},
 {"tvbox-155", "一五五", "https://155api.com/api.php/provide/vod/?ac=list", false},
 {"tvbox-jkun", "jkun", "https://jkunzyapi.com/api.php/provide/vod/?ac=list", false},
 {"tvbox-senlin", "森林", "https://slapibf.com/api.php/provide/vod/?ac=list", false},
 {"tvbox-jingpinx", "精品下", "http://www.jingpinx.com/api.php/provide/vod/?ac=list", false},
 {"tvbox-fanguo", "番果", "http://fhapi9.com/api.php/provide/vod/", false},
 {"tvbox-leguo", "樂果", "https://lbapi9.com/api.php/provide/vod/at/xml", true},
}
var tvboxValidID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

type tvboxType1Provider struct{ host Host; spec tvboxType1Spec }
func (p *tvboxType1Provider) Descriptor() SourceDescriptor {
 return SourceDescriptor{ID:p.spec.id,Name:p.spec.name,DefaultOff:true,
  Description:"TVBox 標準資源 API",NoHostRotation:true,
  ValidCategory:func(s string)bool{return s==""||tvboxValidID.MatchString(s)}}
}
func (p *tvboxType1Provider) fetch(ctx context.Context,action string, params url.Values) (map[string]any,error) {
 endpoint,err:=url.Parse(p.spec.endpoint)
 if err!=nil||(endpoint.Scheme!="https"&&endpoint.Scheme!="http") || endpoint.User!=nil {
  return nil,errors.New("無效的 TVBox 資源接口")
 }
 suffix:=strings.ToLower(endpoint.Path)
 if p.spec.xml {
  if !strings.HasSuffix(suffix,"/api.php/provide/vod/at/xml"){return nil,errors.New("無效的 TVBox XML 接口")}
 }else if !strings.HasSuffix(suffix,"/api.php/provide/vod/") {
  return nil,errors.New("無效的 TVBox JSON 接口")
 }
 q:=endpoint.Query();q.Set("ac",action)
 for k,v:=range params {if len(v)>0{q.Set(k,v[0])}}
 endpoint.RawQuery=q.Encode()
 response,err:=p.host.Fetch(ctx,HTTPRequest{Label:p.spec.name+"接口",Method:"GET",URL:endpoint.String(),Limit:4<<20})
 if err!=nil{return nil,err}
 var data map[string]any
 if p.spec.xml {
  data,err=tvboxXMLToMap(response.Body)
  if err!=nil{return nil,fmt.Errorf("%s XML 接口資料無法解析：%w",p.spec.name,err)}
 }else if err=json.Unmarshal(response.Body,&data);err!=nil{
  return nil,fmt.Errorf("%s 未返回標準 JSON TVBox 接口資料",p.spec.name)
 }
 if _,ok:=data["list"];!ok && action!="list" && action!="detail"{return nil,errors.New("TVBox 接口未返回影片列表")}
 return data,nil
}
// MacCMS v10 XML TVBox type=0 response. Convert only the known
// class/video/play fields; never interpret XML content as executable data.
type tvboxXMLDocument struct {
 XMLName xml.Name `xml:"rss"`
 Class struct{ Items []struct{
  ID string `xml:"id,attr"`
  Name string `xml:",chardata"`
 } `xml:"ty"` } `xml:"class"`
 List struct{
  PageCount string `xml:"pagecount,attr"`
  Videos []struct{
   ID string `xml:"id"`
   Name string `xml:"name"`
   TypeName string `xml:"type"`
   Pic string `xml:"pic"`
   Content string `xml:"des"`
   Note string `xml:"note"`
   Play struct{Lines []struct{
    Flag string `xml:"flag,attr"`
    Text string `xml:",chardata"`
   } `xml:"dd"`} `xml:"dl"`
  } `xml:"video"`
 } `xml:"list"`
}
func tvboxXMLToMap(body []byte)(map[string]any,error){
 var doc tvboxXMLDocument
 if err:=xml.Unmarshal(body,&doc);err!=nil{return nil,err}
 if doc.XMLName.Local!="rss"{return nil,errors.New("XML 缺少 RSS 根節點")}
 result:=map[string]any{"pagecount":doc.List.PageCount}
 classes:=[]any{}
 for _,entry:=range doc.Class.Items{
  if entry.ID!=""&&strings.TrimSpace(entry.Name)!=""{
   classes=append(classes,map[string]any{"type_id":entry.ID,"type_name":strings.TrimSpace(entry.Name)})
  }
 }
 result["class"]=classes
 videos:=[]any{}
 for _,v:=range doc.List.Videos{
  from,urls:=[]string{},[]string{}
  for _,line:=range v.Play.Lines{
   if strings.TrimSpace(line.Text)==""{continue}
   from=append(from,line.Flag)
   urls=append(urls,strings.TrimSpace(line.Text))
  }
  videos=append(videos,map[string]any{
   "vod_id":v.ID,"vod_name":v.Name,"type_name":v.TypeName,
   "vod_pic":v.Pic,"vod_content":v.Content,"vod_remarks":v.Note,
   "vod_play_from":strings.Join(from,"$$"),"vod_play_url":strings.Join(urls,"$$"),
  })
 }
 result["list"]=videos
 return result,nil
}

func tvboxRows(value any) []map[string]any {
 raw,ok:=value.([]any);if !ok{return nil}
 out:=make([]map[string]any,0,len(raw))
 for _,v:=range raw {if row,ok:=v.(map[string]any);ok{out=append(out,row)}}
 return out
}
func tvboxString(v any) string {
 if v==nil{return ""}
 switch x:=v.(type){case string:return strings.TrimSpace(x)
 case json.Number:return x.String()
 case float64:return strconv.FormatFloat(x,'f',-1,64)
 default:return ""}
}
func tvboxFirst(row map[string]any,keys ...string)string{
 for _,key:=range keys {if s:=tvboxString(row[key]);s!=""{return s}}
 return ""
}
func (p *tvboxType1Provider) drama(row map[string]any) (Drama,bool){
 id:=tvboxFirst(row,"vod_id","id")
 if !tvboxValidID.MatchString(id){return Drama{},false}
 title:=tvboxFirst(row,"vod_name","title","name")
 if title=="" {return Drama{},false}
 cover:=tvboxFirst(row,"vod_pic","cover","pic")
 if cover!="" {u,err:=url.Parse(cover);if err!=nil||u.User!=nil||u.Scheme!="https"&&u.Scheme!="http"{cover=""}}
 return Drama{ID:ProviderDramaID(p.spec.id,id),Source:p.spec.id,SourceID:id,
  Title:title,Name:title,Cover:cover,CoverURL:cover,
  Desc:tvboxFirst(row,"vod_content","vod_blurb","vod_remarks"),
  CategoryName:tvboxFirst(row,"type_name"),ChannelName:p.spec.name},true
}
func (p *tvboxType1Provider) Categories(ctx context.Context,force bool)([]Category,error){
 data,err:=p.fetch(ctx,"list",nil);if err!=nil{return nil,err}
 seen:=map[string]bool{};out:=[]Category{}
 for _,r:=range tvboxRows(data["class"]){
  id:=tvboxFirst(r,"type_id","id"); name:=tvboxFirst(r,"type_name","name")
  if id!=""&&name!=""&&tvboxValidID.MatchString(id)&&!seen[id]{
   seen[id]=true;out=append(out,Category{ID:id,Name:name})
  }
 }
 return out,nil // Some APIs have no category metadata.
}
func (p *tvboxType1Provider) CatalogPage(ctx context.Context,page int,category string)([]Drama,bool,error){
 return p.list(ctx,page,category,"")
}
func (p *tvboxType1Provider) Search(ctx context.Context,page int,query string)([]Drama,bool,error){
 return p.list(ctx,page,"",query)
}
func (p *tvboxType1Provider) list(ctx context.Context,page int,category,query string)([]Drama,bool,error){
 if page<1||page>10000||len(query)>200||(category!=""&&!tvboxValidID.MatchString(category)){
  return nil,false,errors.New("TVBox 分類、頁碼或搜尋參數無效")
 }
 params:=url.Values{"pg":{strconv.Itoa(page)}}
 if category!=""{params.Set("t",category)}
 if query!=""{params.Set("wd",query)}
 data,err:=p.fetch(ctx,"list",params);if err!=nil{return nil,false,err}
 seen:=map[string]bool{};out:=[]Drama{}
 for _,row:=range tvboxRows(data["list"]){
  drama,ok:=p.drama(row);if !ok||seen[drama.ID]{continue}
  seen[drama.ID]=true;out=append(out,drama)
 }
 pages,_:=strconv.Atoi(tvboxString(data["pagecount"]))
 return out, pages>page,nil
}
func (p *tvboxType1Provider) Detail(ctx context.Context,id string)(Drama,[]Chapter,error){
 if !tvboxValidID.MatchString(id){return Drama{},nil,errors.New("影片 ID 無效")}
 data,err:=p.fetch(ctx,"detail",url.Values{"ids":{id}})
 if err!=nil{return Drama{},nil,err}
 for _,row:=range tvboxRows(data["list"]){
  drama,ok:=p.drama(row);if !ok||drama.SourceID!=id{continue}
  chapters:=[]Chapter{}
  from:=strings.Split(tvboxFirst(row,"vod_play_from"),"$$$")
  groups:=strings.Split(tvboxFirst(row,"vod_play_url"),"$$$")
  // Keep source/index stable: an unrelated play line must never alter film identity.
  for line,group:=range groups {
   for pos,item:=range strings.Split(group,"#"){
    pair:=strings.SplitN(item,"$",2);name:="第"+strconv.Itoa(pos+1)+"集";address:=strings.TrimSpace(item)
    if len(pair)==2 {name=strings.TrimSpace(pair[0]);address=strings.TrimSpace(pair[1])}
    u,e:=url.Parse(address)
    if e!=nil||u.User!=nil||(u.Scheme!="http"&&u.Scheme!="https")||u.Hostname()==""{continue}
    if name==""{name="第"+strconv.Itoa(pos+1)+"集"}
    key:=strconv.Itoa(line+1)+"-"+strconv.Itoa(pos+1)
    chapters=append(chapters,Chapter{ID:ProviderChapterID(p.spec.id,id,key),Source:p.spec.id,
     Title:name,VideoURL:address,PageURL:address,CurrentEpisode:json.RawMessage(strconv.Itoa(len(chapters)+1))})
   }
  }
  _=from // Names of MacCMS play lines are optional.
  return drama,chapters,nil
 }
 return Drama{},nil,errors.New("找不到指定影片詳情")
}
func (p *tvboxType1Provider) ResolveMedia(ctx context.Context,task Task)(ProviderMedia,error){
 source,id,ok:=SplitProviderDramaID(task.DramaID)
 if !ok||source!=p.spec.id||!tvboxValidID.MatchString(id)||
  !strings.HasPrefix(task.Chapter.ID,ProviderChapterID(source,id,"")) {
  return ProviderMedia{},errors.New("影片與分集 ID 不一致")
 }
 address:=strings.TrimSpace(task.Chapter.VideoURL)
 u,err:=url.Parse(address)
 if err!=nil||u.User!=nil||(u.Scheme!="https"&&u.Scheme!="http")||u.Hostname()==""{
  return ProviderMedia{},errors.New("該 TVBox 播放線路需要額外解析器，未提供直連影片網址")
 }
 return p.host.PrepareWebMedia(ctx,ProviderMedia{URL:address,Referer:p.spec.endpoint},p.spec.name)
}
func init(){
 for _,entry:=range tvboxType1Specs {
  spec:=entry
  descriptor:=SourceDescriptor{ID:spec.id,Name:spec.name,DefaultOff:true,
   Description:"TVBox type:1",NoHostRotation:true,
   ValidCategory:func(s string)bool{return s==""||tvboxValidID.MatchString(s)}}
  Register(descriptor,func(host Host)Provider{return &tvboxType1Provider{host:host,spec:spec}})
 }
}
