package missavrecovery

import (
    "context"
    "errors"
    "fmt"
    "html"
    "net/url"
    "regexp"
    "strconv"
    "strings"

    "duanjuapp/native/core"
)

// Experimental reconstruction from original 1.0.10+110 ARM64 binary.
// The original private Go source is NOT included in the APK and cannot be
// recovered by interpreting the function names alone. This module is built
// independently of the 13 existing custom sources until verified.
const sourceID = "missav"
const baseURL = "https://missav.live"
var categorySlug = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var categoryLinks = regexp.MustCompile(`(?is)<a\b[^>]*href=["'](?:https?://(?:www\.)?missav\.live)?/cn/genres/([a-z0-9_-]{1,64})/?["'][^>]*>(.*?)</a>`)
var cardLinks = regexp.MustCompile(`(?is)<a\b[^>]*href=["'](?:https?://(?:www\.)?missav\.live)?/cn/([a-z0-9][a-z0-9_-]{2,127})/?["'][^>]*>(.*?)</a>`)
var titleAttr = regexp.MustCompile(`(?is)\b(?:alt|title)=["']([^"']+)["']`)
var imageURL = regexp.MustCompile(`(?is)<(?:img|source)\b[^>]*\b(?:data-src|data-original|src|srcset)=["']([^"']+)["']`)
var ogImage = regexp.MustCompile(`(?is)<meta\b[^>]*\bproperty=["']og:image["'][^>]*\bcontent=["']([^"']+)["']`)
var mediaURL = regexp.MustCompile(`(?i)https?:?\\?/\\?/[^"'< >\s\\]+?\.(?:m3u8|mp4)(?:\?[^"'<>\s\\]*)?`)
var nextPage = regexp.MustCompile(`(?is)(?:rel=["']next["']|aria-label=["']next["']|\bpage=\d+)`)

var descriptor = core.SourceDescriptor{
    ID: sourceID, Name: "MissAV +",
    Description: "由 110 APK 路徑與函式資訊重建的實驗站源；未驗證動態媒體解析",
    DefaultOff: true, BaseURL: baseURL,
    NoHostRotation: true,
    ValidCategory: validCategory,
}

type provider struct { host core.Host }
var _ core.Provider = (*provider)(nil)
var _ core.Searcher = (*provider)(nil)

func init() {
    core.Register(descriptor, func(h core.Host) core.Provider { return &provider{host:h} })
}

func validCategory(category string) bool {
    return category == "" || category == "new" || categorySlug.MatchString(category)
}
func (p *provider) Descriptor() core.SourceDescriptor { return descriptor }
func (p *provider) BaseURL() string { return baseURL }

func (p *provider) fetch(ctx context.Context, path string) (string,error) {
    if !strings.HasPrefix(path, "/cn/") { return "", errors.New("invalid source route") }
    response,err:=p.host.FetchText(ctx,core.TextRequest{URL:baseURL+path,Referer:baseURL+"/cn/"})
    if err!=nil{return "",err}
    return response.Body,nil
}
func (p *provider) Categories(ctx context.Context, force bool) ([]core.Category,error) {
    categories:=[]core.Category{{ID:"new",Name:"最新"}}
    text,err:=p.fetch(ctx,"/cn/genres/")
    if err!=nil { return categories,err }
    seen:=map[string]bool{"new":true}
    for _,m:=range categoryLinks.FindAllStringSubmatch(text,-1) {
        if len(m)<3 || !validCategory(m[1]) || seen[m[1]] {continue}
        name:=core.CleanText(m[2]);if name==""{continue}
        seen[m[1]]=true
        categories=append(categories,core.Category{ID:m[1],Name:name})
    }
    return categories,nil
}
func (p *provider) CatalogPage(ctx context.Context,page int,category string)([]core.Drama,bool,error){
    if page<1 {return nil,false,errors.New("invalid page")}
    if !validCategory(category) {return nil,false,errors.New("invalid category")}
    path:="/cn/new"
    if category!="" && category!="new" {path="/cn/genres/"+url.PathEscape(category)}
    if page>1 {path+="?page="+strconv.Itoa(page)}
    body,err:=p.fetch(ctx,path)
    if err!=nil{return nil,false,err}
    return parseCards(body),nextPage.MatchString(body),nil
}
func (p *provider) Search(ctx context.Context,page int,query string)([]core.Drama,bool,error){
    query=strings.TrimSpace(query)
    if query==""{return nil,false,nil}
    if page<1{return nil,false,errors.New("invalid page")}
    path:="/cn/search/"+url.PathEscape(query)
    if page>1 {path+="?page="+strconv.Itoa(page)}
    body,err:=p.fetch(ctx,path)
    if err!=nil{return nil,false,err}
    return parseCards(body),nextPage.MatchString(body),nil
}
func parseCards(raw string) []core.Drama {
    var rows []core.Drama
    seen:=map[string]bool{}
    for _,m:=range cardLinks.FindAllStringSubmatch(raw,-1){
        if len(m)<3 {continue}
        slug:=m[1]
        if slug=="genres" || slug=="search" || slug=="new" || seen[slug] {continue}
        content:=m[2]
        title:=""
        if v:=titleAttr.FindStringSubmatch(content);len(v)>1{title=strings.TrimSpace(html.UnescapeString(v[1]))}
        if title=="" {title=core.CleanText(content)}
        if title=="" {continue}
        cover:=""
        if v:=imageURL.FindStringSubmatch(content);len(v)>1{
            cover=core.ResolveProviderURL(baseURL+"/cn/",v[1])
        }
        seen[slug]=true
        rows=append(rows,core.Drama{
            ID:core.ProviderDramaID(sourceID,slug),
            Source:sourceID,SourceID:slug,Title:title,Name:title,
            Cover:cover,TotalEpisode:1,
        })
    }
    return rows
}
func (p *provider) Detail(ctx context.Context,slug string)(core.Drama,[]core.Chapter,error){
    if !categorySlug.MatchString(slug) && !regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`).MatchString(slug){
        return core.Drama{},nil,errors.New("invalid media ID")
    }
    path:="/cn/"+url.PathEscape(slug)
    body,err:=p.fetch(ctx,path)
    if err!=nil{return core.Drama{},nil,err}
    title:=core.ExtractPageTitle(body)
    if title=="" {title=slug}
    cover:=""
    if m:=ogImage.FindStringSubmatch(body);len(m)>1 {cover=core.ResolveProviderURL(baseURL+path,m[1])}
    drama:=core.Drama{ID:core.ProviderDramaID(sourceID,slug),Source:sourceID,
        SourceID:slug,Title:title,Name:title,Cover:cover,Desc:core.ExtractDescription(body),
        TotalEpisode:1}
    chapter:=core.Chapter{ID:core.ProviderChapterID(sourceID,slug,"1"),
        Source:sourceID,Title:"播放",CurrentEpisode:core.RawEpisode(1),
        PageURL:baseURL+path,Referer:baseURL+path}
    return drama,[]core.Chapter{chapter},nil
}
func (p *provider) ResolveMedia(ctx context.Context,task core.Task)(core.ProviderMedia,error){
    page:=task.Chapter.PageURL
    if !strings.HasPrefix(page,baseURL+"/cn/") {
        return core.ProviderMedia{},errors.New("invalid or missing page URL")
    }
    body,err:=p.fetch(ctx,strings.TrimPrefix(page,baseURL))
    if err!=nil{return core.ProviderMedia{},err}
    match:=mediaURL.FindString(body)
    if match=="" {
        return core.ProviderMedia{},fmt.Errorf("MissAV media URL not exposed directly; original 110 packed player routine has not been reconstructed")
    }
    match=strings.ReplaceAll(match,`\/`,`/`)
    match=strings.ReplaceAll(match,`\\`,``)
    parsed:=core.NormalizeProviderMediaURL(page,match)
    if parsed=="" {return core.ProviderMedia{},errors.New("invalid media URL")}
    return core.ProviderMedia{URL:parsed,Referer:page},nil
}
