package core

// Standard TVBox type=1 MacCMS JSON providers. A type=3 csp_* spider is
// not a URL and cannot be treated as a MacCMS interface.
import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/url"
 "regexp"
 "strconv"
 "strings"
)

type tvboxType1Spec struct{ id, name, endpoint string }
var tvboxType1Specs = []tvboxType1Spec{
 {"tvbox-laoguo", "老果", "https://apilsbzy1.com/api.php/provide/vod/?ac=list"},
 {"tvbox-155", "一五五", "https://155api.com/api.php/provide/vod/?ac=list"},
 {"tvbox-jkun", "jkun", "https://jkunzyapi.com/api.php/provide/vod/?ac=list"},
 {"tvbox-senlin", "森林", "https://slapibf.com/api.php/provide/vod/?ac=list"},
 {"tvbox-jingpinx", "精品下", "http://www.jingpinx.com/api.php/provide/vod/?ac=list"},
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
 if err!=nil||!strings.HasSuffix(strings.ToLower(endpoint.Path),"/api.php/provide/vod/") ||
  (endpoint.Scheme!="https"&&endpoint.Scheme!="http") || endpoint.User!=nil {return nil,errors.New("無效的 TVBox 資源接口")}
 q:=endpoint.Query();q.Set("ac",action)
 for k,v:=range params {if len(v)>0{q.Set(k,v[0])}}
 endpoint.RawQuery=q.Encode()
 response,err:=p.host.Fetch(ctx,HTTPRequest{Label:p.spec.name+"接口",Method:"GET",URL:endpoint.String(),Limit:4<<20})
 if err!=nil{return nil,err}
 var data map[string]any
 if json.Unmarshal(response.Body,&data)!=nil{return nil,fmt.Errorf("%s 未返回標準 JSON TVBox 接口資料",p.spec.name)}
 if _,ok:=data["list"];!ok && action!="list" && action!="detail"{return nil,errors.New("TVBox 接口未返回影片列表")}
 return data,nil
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
