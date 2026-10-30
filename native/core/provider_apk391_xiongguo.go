package core

import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "strconv"
 "strings"
 "time"
 "sync"
)

//熊果独立网页 API，与 MacCMS 不兼容。
var xiongguoFilmMid sync.Map // ID -> type_Mid from verified catalog
var xiongguoFilmTitle sync.Map // ID -> verified catalog title
var xiongguoFilmCover sync.Map // ID -> verified catalog image

type xiongguoHomeResponse struct {Data struct{ResultList []struct{Type string `json:"t_type"`;List []struct{ID json.Number `json:"id"`;Mid json.Number `json:"type_Mid"`;Name string `json:"vod_name"`;Picture string `json:"vod_pic"`} `json:"t_list"`} `json:"resultList"`} `json:"data"`}

type xiongguoListResponse struct {
 Data struct {
  Count int `json:"count"`
  ResultList []struct {
   ID json.Number `json:"id"`
   TypeMid json.Number `json:"type_Mid"`
   Name string `json:"vod_name"`
   Picture string `json:"vod_pic"`
  } `json:"resultList"`
 } `json:"data"`
}
func xiongguoTypeMid(category string) int {return 1}
func (d *Downloader) fetchXiongguoCatalog(ctx context.Context,page int,category,query string)([]Drama,bool,error){
 if page<1||page>100000||!apk391ValidCategory("xiongguo",category){return nil,false,errors.New("熊果分类或页码无效")}
 if query!="" {return nil,false,errors.New("熊果搜索暂不支持")}
 typeID:=0
 if category!=""{typeID,_=strconv.Atoi(category)}
 command:="WEB_GET_INFO"
 if category==""{command="WEB_GET_ALL"}
 payload:=map[string]interface{}{"command":command,"pageNumber":page,"RecordsPage":20,"typeId":typeID,"typeMid":xiongguoTypeMid(category),"languageType":"CN","content":""}
 body,_:=json.Marshal(payload)
 ctx,cancel:=context.WithTimeout(ctx,8*time.Second);defer cancel()
 req,err:=http.NewRequestWithContext(ctx,http.MethodPost,"https://spiderscloudcn2.51111666.com/forward",bytes.NewReader(body))
 if err!=nil{return nil,false,err}
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("User-Agent",lujj31UserAgent)
 req.Header.Set("Referer",apk391BaseURL("xiongguo")+"/home.html")
 client:=&http.Client{Timeout:8*time.Second}
 res,err:=client.Do(req)
 if err!=nil{return nil,false,fmt.Errorf("熊果分类 API 连接失败: %w",err)}
 defer res.Body.Close()
 if res.StatusCode!=http.StatusOK{return nil,false,fmt.Errorf("熊果分类 API HTTP %d",res.StatusCode)}
 raw,err:=io.ReadAll(io.LimitReader(res.Body,4<<20))
 if err!=nil{return nil,false,err}
 if category==""{
  var home xiongguoHomeResponse
  decoder:=json.NewDecoder(bytes.NewReader(raw));decoder.UseNumber()
  if err:=decoder.Decode(&home);err!=nil{return nil,false,fmt.Errorf("熊果首页 API JSON 解析失败: %w",err)}
  seen:=map[string]bool{};cards:=[]Drama{}
  for _,section:=range home.Data.ResultList{
   if section.Type!="M_VOIDE"{continue}
   for _,entry:=range section.List {
    id:=entry.ID.String();title:=strings.TrimSpace(entry.Name);cover:=strings.TrimSpace(entry.Picture)
    if !apk391Numeric.MatchString(id)||title==""||!isProviderHTTPMediaURL(cover)||seen[id]{continue}
    seen[id]=true
    xiongguoFilmTitle.Store(id,title);xiongguoFilmCover.Store(id,cover)
    if mid:=entry.Mid.String();mid!=""{xiongguoFilmMid.Store(id,mid)}
    cards=append(cards,Drama{ID:providerDramaID("xiongguo",id),Source:"xiongguo",SourceID:id,Title:truncate(title,512),Name:truncate(title,512),Cover:cover,CoverURL:cover,ChannelName:"熊果"})
   }
  }
  return cards,false,nil
 }
 var parsed xiongguoListResponse
 decoder:=json.NewDecoder(bytes.NewReader(raw));decoder.UseNumber()
 if err:=decoder.Decode(&parsed);err!=nil{return nil,false,fmt.Errorf("熊果 API 不是有效 JSON: %w",err)}
 entries:=make([]Drama,0,len(parsed.Data.ResultList))
 seen:=map[string]bool{}
 for _,item:=range parsed.Data.ResultList {
  id:=item.ID.String()
  if !apk391Numeric.MatchString(id)||seen[id]||strings.TrimSpace(item.Name)==""{continue}
  cover:=strings.TrimSpace(item.Picture)
  if !isProviderHTTPMediaURL(cover){continue}
  seen[id]=true
  xiongguoFilmTitle.Store(id,strings.TrimSpace(item.Name));xiongguoFilmCover.Store(id,cover)
  if mid:=item.TypeMid.String();mid!=""{xiongguoFilmMid.Store(id,mid)}
  title:=truncate(strings.TrimSpace(item.Name),512)
  entries=append(entries,Drama{ID:providerDramaID("xiongguo",id),Source:"xiongguo",SourceID:id,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"熊果"})
 }
 return entries,page*20<parsed.Data.Count,nil
}

func xiongguoStored(m *sync.Map,id string)string{
 if v,ok:=m.Load(id);ok{if s,ok:=v.(string);ok{return s}}
 return ""
}
func (d *Downloader) fetchXiongguoDetail(ctx context.Context,id string)(Drama,[]Chapter,error){
 if !apk391Numeric.MatchString(id){return Drama{},nil,errors.New("熊果影片编号无效")}
 mid:=xiongguoStored(&xiongguoFilmMid,id)
 if mid==""{mid="1"}
 reqBody,_:=json.Marshal(map[string]interface{}{"command":"WEB_GET_INFO_DETAIL","id":id,"type_Mid":mid,"languageType":"CN"})
 ctx,cancel:=context.WithTimeout(ctx,8*time.Second);defer cancel()
 req,err:=http.NewRequestWithContext(ctx,http.MethodPost,"https://spiderscloudcn2.51111666.com/forward",bytes.NewReader(reqBody))
 if err!=nil{return Drama{},nil,err}
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("User-Agent",lujj31UserAgent)
 req.Header.Set("Referer",apk391BaseURL("xiongguo")+"/home.html")
 res,err:=(&http.Client{Timeout:8*time.Second}).Do(req)
 if err!=nil{return Drama{},nil,fmt.Errorf("熊果详情 API 连接失败: %w",err)}
 defer res.Body.Close()
 if res.StatusCode!=http.StatusOK{return Drama{},nil,fmt.Errorf("熊果详情 API HTTP %d",res.StatusCode)}
 raw,err:=io.ReadAll(io.LimitReader(res.Body,4<<20));if err!=nil{return Drama{},nil,err}
 var envelope struct{Data struct{Result struct{
  ID json.Number `json:"id"`
  Mid json.Number `json:"type_Mid"`
  Name string `json:"vod_name"`
  Picture string `json:"vod_pic"`
  URL string `json:"vod_url"`
 } `json:"result"`} `json:"data"`}
 decoder:=json.NewDecoder(bytes.NewReader(raw));decoder.UseNumber()
 if err:=decoder.Decode(&envelope);err!=nil{return Drama{},nil,fmt.Errorf("熊果详情 JSON 错误: %w",err)}
 movie:=envelope.Data.Result
 if movie.ID.String()!=""&&movie.ID.String()!=id{return Drama{},nil,errors.New("熊果详情影片编号与片单不符")}
 title:=xiongguoStored(&xiongguoFilmTitle,id)
 if title==""{title=strings.TrimSpace(movie.Name)}
 if title==""{title="影片 "+id}
 cover:=xiongguoStored(&xiongguoFilmCover,id)
 if cover==""&&isProviderHTTPMediaURL(movie.Picture){cover=movie.Picture}
 // The site serves a single vod_url per item. Resolving its CDN prefix
 // requires /getDataInit macVodLinkMap; preserve the original token here.
 if strings.TrimSpace(movie.URL)==""{return Drama{},nil,errors.New("熊果详情缺少播放地址")}
 chapter:=Chapter{ID:providerChapterID("xiongguo",id,"1"),Source:"xiongguo",Title:"第1集",CurrentEpisode:rawEpisode(1),
  PageURL:apk391BaseURL("xiongguo")+"/home.html",VideoURL:"xiongguo-play://"+id,Referer:apk391BaseURL("xiongguo")+"/home.html"}
 item:=Drama{ID:providerDramaID("xiongguo",id),Source:"xiongguo",SourceID:id,Title:title,Name:title,Cover:cover,CoverURL:cover,EpisodeCount:1,TotalEpisode:1,ChannelName:"熊果"}
 return item,[]Chapter{chapter},nil
}

func (d *Downloader) resolveXiongguoMedia(ctx context.Context,task Task)(providerMedia,error){
 source,id,ok:=splitProviderDramaID(task.DramaID)
 if !ok||source!="xiongguo"||!apk391Numeric.MatchString(id)||
  !strings.HasPrefix(task.Chapter.ID,providerChapterID(source,id,"")){
  return providerMedia{},errors.New("熊果播放分集编号无效")
 }
 mid:=xiongguoStored(&xiongguoFilmMid,id)
 if mid==""{mid="1"}
 ctx,cancel:=context.WithTimeout(ctx,10*time.Second);defer cancel()
 post:=func(path string,payload interface{})([]byte,error){
  body,_:=json.Marshal(payload)
  req,err:=http.NewRequestWithContext(ctx,http.MethodPost,"https://spiderscloudcn2.51111666.com/"+path,bytes.NewReader(body))
  if err!=nil{return nil,err}
  req.Header.Set("Content-Type","application/json")
  req.Header.Set("User-Agent",lujj31UserAgent)
  req.Header.Set("Referer",apk391BaseURL(source)+"/home.html")
  res,err:=(&http.Client{Timeout:9*time.Second}).Do(req)
  if err!=nil{return nil,err}
  defer res.Body.Close()
  if res.StatusCode!=http.StatusOK{return nil,fmt.Errorf("熊果播放 API HTTP %d",res.StatusCode)}
  return io.ReadAll(io.LimitReader(res.Body,4<<20))
 }
 detailRaw,err:=post("forward",map[string]interface{}{"command":"WEB_GET_INFO_DETAIL","id":id,"type_Mid":mid,"languageType":"CN"})
 if err!=nil{return providerMedia{},err}
 var detail struct{Data struct{Result struct{
  ID json.Number `json:"id"`
  URL string `json:"vod_url"`
  Server json.Number `json:"vod_server_id"`
 } `json:"result"`} `json:"data"`}
 decoder:=json.NewDecoder(bytes.NewReader(detailRaw));decoder.UseNumber()
 if err:=decoder.Decode(&detail);err!=nil{return providerMedia{},err}
 info:=detail.Data.Result
 if info.ID.String()!=""&&info.ID.String()!=id{return providerMedia{},errors.New("熊果播放影片编号不符")}
 token:=strings.TrimSpace(info.URL)
 if token==""{return providerMedia{},errors.New("熊果播放缺少 vod_url")}
 mediaURL:=token
 if !isProviderHTTPMediaURL(mediaURL){
  initRaw,err:=post("getDataInit",map[string]interface{}{"name":"John","age":31,"city":"New York"})
  if err!=nil{return providerMedia{},err}
  var init struct{Data struct{Links map[string]struct{
   Link1 string `json:"LINK_1"`
   Link2 string `json:"LINK_2"`
  } `json:"macVodLinkMap"`} `json:"data"`}
  if err:=json.Unmarshal(initRaw,&init);err!=nil{return providerMedia{},err}
  server:=info.Server.String()
  links,ok:=init.Data.Links[server]
  if !ok{return providerMedia{},errors.New("熊果播放线路服务器不存在")}
  mediaURL=links.Link1+token
  if !isProviderHTTPMediaURL(mediaURL){mediaURL=links.Link2+token}
 }
 if !isProviderHTTPMediaURL(mediaURL){return providerMedia{},errors.New("熊果播放地址无法转换为有效媒体 URL")}
 return d.prepareWebProviderMedia(ctx,providerMedia{URL:mediaURL,Referer:apk391BaseURL(source)+"/home.html"},"熊果")
}
