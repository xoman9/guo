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
)

//熊果独立网页 API，与 MacCMS 不兼容。
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
 payload:=map[string]interface{}{"command":"WEB_GET_INFO","pageNumber":page,"RecordsPage":20,"typeId":typeID,"typeMid":xiongguoTypeMid(category),"languageType":"CN","content":""}
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
  title:=truncate(strings.TrimSpace(item.Name),512)
  entries=append(entries,Drama{ID:providerDramaID("xiongguo",id),Source:"xiongguo",SourceID:id,Title:title,Name:title,Cover:cover,CoverURL:cover,ChannelName:"熊果"})
 }
 return entries,page*20<parsed.Data.Count,nil
}
