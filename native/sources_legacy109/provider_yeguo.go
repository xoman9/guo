package legacy109
import(
 "context"
 "errors"
 "fmt"
 "net/url"
 "strconv"
 "strings"
 "duanjuapp/native/core"
)
func (p *yeguo109) Categories(ctx context.Context,_ bool)([]core.Category,error){
 m,err:=p.call(ctx,"/api/home/contentOptions",nil)
 if err!=nil{return nil,err}
 filters:=hjMap(m["video_filter"])
 if filters==nil{return nil,errors.New("Yeguo missing filters")}
 seen:=map[string]bool{};var out []core.Category
 for _,name:=range yg109Fields{
  rows:=hjRows(hjMap(filters[name])["list"])
  if len(rows)>300{return nil,errors.New("too many Yeguo filters")}
  for _,v:=range rows{
   m:=hjMap(v);id,title:=hjString(m["value"]),hjString(m["name"])
   if id=="0"{continue}
   candidate:=name+":"+id
   if !yg109ValidCat(candidate)||title==""||seen[candidate]{continue}
   seen[candidate]=true;out=append(out,core.Category{ID:candidate,Name:title})
  }
 }
 if len(out)==0{return nil,errors.New("Yeguo returned no filters")}
 return out,nil
}
func ygDrama109(row map[string]any,site string)(core.Drama,error){
 id,title:=hgField(row,"video_id","id"),hjString(row["title"])
 if !dsd109Num(id)||title==""{return core.Drama{},errors.New("invalid Yeguo drama")}
 count:=0
 for _,field:=range []string{"episode_count","episodes","total_serial"}{
  n,ok:=hjInt(row[field],100000);if ok&&n>count{count=n}
 }
 cover:=core.ResolveProviderURL(site+"/",hjString(row["cover"]))
 return core.Drama{
  ID:core.ProviderDramaID("yeguo",id),Source:"yeguo",SourceID:id,
  Title:title,Name:title,Cover:cover,TotalEpisode:count,Desc:hjString(row["description"]),ChannelName:"野果",
 },nil
}
func (p *yeguo109) CatalogPage(ctx context.Context,page int,cat string)([]core.Drama,bool,error){
 return p.listing(ctx,page,cat,"")
}
func (p *yeguo109) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
 query=strings.TrimSpace(query);if query==""{return nil,false,nil}
 return p.listing(ctx,page,"",query)
}
func (p *yeguo109) listing(ctx context.Context,page int,cat,query string)([]core.Drama,bool,error){
 if page<1||page>1000000||len(query)>1024||!yg109ValidCat(cat){
  return nil,false,errors.New("invalid Yeguo listing")
 }
 route:="/api/theater/exploreList"
 args:=url.Values{"page":{strconv.Itoa(page)},"limit":{"20"}}
 if query!=""{route="/api/search/result";args.Set("keyword",query);args.Set("tab","video")
 }else if cat!=""{
  k,v,_:=strings.Cut(cat,":");args.Set(k,v)
 }
 m,err:=p.call(ctx,route,args)
 if err!=nil{return nil,false,err}
 rows:=hjRows(m["list"])
 if rows==nil{return nil,false,errors.New("invalid Yeguo list")}
 hasMore:=len(rows)>=20
 if total,ok:=hjInt(m["total"],100000000);ok{hasMore=page*20<total}
 if v,ok:=m["has_more"].(bool);ok{hasMore=v}
 seen:=map[string]bool{};var out []core.Drama
 for _,value:=range rows{
  item,err:=ygDrama109(hjMap(value),yg109Site)
  if err!=nil{return nil,false,err}
  if seen[item.ID]{return nil,false,errors.New("duplicate Yeguo drama")}
  seen[item.ID]=true;out=append(out,item)
 }
 return out,hasMore,nil
}
func (p *yeguo109) Detail(ctx context.Context,id string)(core.Drama,[]core.Chapter,error){
 if !dsd109Num(id){return core.Drama{},nil,errors.New("invalid Yeguo drama")}
 m,err:=p.call(ctx,"/api/playlet/detail",url.Values{
  "video_id":{id},"id":{id},"episode_id":{"0"},"related_limit":{"0"},
 })
 if err!=nil{return core.Drama{},nil,err}
 drama,err:=ygDrama109(m,yg109Site)
 if err!=nil||drama.SourceID!=id{return core.Drama{},nil,errors.New("Yeguo detail mismatch")}
 episodes:=hjRows(m["episodes"])
 if len(episodes)>10000{return core.Drama{},nil,errors.New("too many episodes")}
 seen:=map[string]bool{};orders:=map[int]bool{}
 var chapters []core.Chapter
 for _,v:=range episodes{
  row:=hjMap(v)
  if row["is_adv"]==true||hjString(row["is_adv"])=="1"{continue}
  eid:=hjString(row["id"])
  n,valid:=hjInt(row["sort"],100000)
  if !dsd109Num(eid)||!valid||n<1||seen[eid]||orders[n]{
   return core.Drama{},nil,errors.New("invalid Yeguo episode")
  }
  seen[eid],orders[n]=true,true
  title:=hjString(row["title"]);if title==""{title=fmt.Sprintf("第%d集",n)}
  chapters=append(chapters,core.Chapter{
   ID:core.ProviderChapterID("yeguo",id,eid),Source:"yeguo",
   Title:title,CurrentEpisode:core.RawEpisode(n),
   PageURL:fmt.Sprintf("%s/drama/video/%s/ep-%d/",yg109Site,id,n),Referer:yg109Site+"/",
  })
 }
 core.SortProviderChapters(chapters)
 drama.TotalEpisode=len(chapters)
 return drama,chapters,nil
}
// verifyChapter requires a chapter to be present in the current drama's
// canonical detail list. It also prevents stale/mismatched episode page URLs.
func (p *yeguo109) verifyChapter(ctx context.Context,dramaID string,chapter core.Chapter)error{
 _,chapters,err:=p.Detail(ctx,dramaID)
 if err!=nil{return fmt.Errorf("Yeguo episode verification failed: %w",err)}
 for _,candidate:=range chapters{
  if candidate.ID!=chapter.ID{continue}
  if chapter.PageURL!=""&&candidate.PageURL!=chapter.PageURL{
   return errors.New("Yeguo episode page does not match selected drama")
  }
  return nil
 }
 return errors.New("Yeguo episode does not belong to selected drama")
}
func (p *yeguo109) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
 source,id,valid:=core.SplitProviderDramaID(task.DramaID)
 if !valid||source!="yeguo"||!dsd109Num(id){return core.ProviderMedia{},errors.New("invalid Yeguo drama")}
 prefix:=core.ProviderChapterID("yeguo",id,"")
 if !strings.HasPrefix(task.Chapter.ID,prefix){return core.ProviderMedia{},errors.New("invalid Yeguo chapter")}
 eid:=strings.TrimPrefix(task.Chapter.ID,prefix)
 if !dsd109Num(eid){return core.ProviderMedia{},errors.New("invalid episode")}
 // The play endpoint may accept a valid numeric episode belonging to a
 // different drama. Verify current ownership before requesting its URL.
 if err:=p.verifyChapter(ctx,id,task.Chapter);err!=nil{return core.ProviderMedia{},err}
 m,err:=p.call(ctx,"/api/playlet/play",url.Values{
  "playlet_id":{id},"video_id":{id},"episode_id":{eid},
 })
 if err!=nil{return core.ProviderMedia{},err}
 if hgField(m,"playlet_id")!=id||hgField(m,"id")!=eid{
  return core.ProviderMedia{},errors.New("Yeguo playback does not match requested episode")
 }
 if m["is_adv"]==true||hjString(m["is_adv"])=="1"{
  return core.ProviderMedia{},errors.New("Yeguo did not provide the full episode")
 }
 n,ok:=hjInt(m["episode_sort"],100000)
 if !ok||n<1{return core.ProviderMedia{},errors.New("invalid Yeguo playback order")}
 referer:=fmt.Sprintf("%s/drama/video/%s/ep-%d/",yg109Site,id,n)
 if task.Chapter.PageURL!=""&&task.Chapter.PageURL!=referer{
  return core.ProviderMedia{},errors.New("Yeguo playback order does not match chapter")
 }
 var addresses []string
 for _,field:=range []string{"video_url","video_url_h265"}{
  raw:=hjString(m[field])
  if raw==""{continue}
  address:=core.ResolveProviderURL(referer,raw)
  if core.IsProviderHTTPMediaURL(address){addresses=append(addresses,address)}
 }
 if len(addresses)==0{return core.ProviderMedia{},errors.New("Yeguo has not authorized a playable URL")}
 media:=core.ProviderMedia{URL:addresses[0],Referer:referer}
 return p.host.PrepareWebMedia(ctx,media,"野果")
}
