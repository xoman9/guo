package missavrecovery

import (
    "context"
    "errors"
    "html"
    "net/url"
    "regexp"
    "strconv"
    "strings"

    "duanjuapp/native/core"
)

// Experimental reconstruction from the original 1.0.10+110 ARM64 binary.
// This does not reproduce the original dynamic playback implementation.
const sourceID = "missav"
const baseURL = "https://missav.live"
var categorySlug = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var mediaSlug = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$`)
var categoryLinks = regexp.MustCompile(`(?is)<a\b[^>]*href=["'](?:https?://(?:www\.)?missav\.live)?/cn/genres/([a-z0-9_-]{1,64})/?["'][^>]*>(.*?)</a>`)
var cardLinks = regexp.MustCompile(`(?is)<a\b[^>]*href=["'](?:https?://(?:www\.)?missav\.live)?/cn/([a-zA-Z0-9][a-zA-Z0-9_-]{0,127})/?["'][^>]*>(.*?)</a>`)
var titleAttr = regexp.MustCompile(`(?is)\b(?:alt|title)=["']([^"']+)["']`)
var imageURL = regexp.MustCompile(`(?is)<(?:img|source)\b[^>]*\b(?:data-src|data-original|src|srcset)=["']([^"']+)["']`)
var ogImage = regexp.MustCompile(`(?is)<meta\b[^>]*\bproperty=["']og:image["'][^>]*\bcontent=["']([^"']+)["']`)
var directVideoAttribute = regexp.MustCompile(`(?is)<(?:video|source)\b[^>]*\b(?:src|data-src|data-hls)\s*=\s*["']([^"']+)["']`)
var directPlayerJSON = regexp.MustCompile(`(?is)"(?:video_url|play_url|src)"\s*:\s*"([^"]+)"`)
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
    if !mediaSlug.MatchString(slug){
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
// mediaPageForTask ensures a playlist entry cannot request another drama's page.
func mediaPageForTask(task core.Task) (string, error) {
    source, slug, ok := core.SplitProviderDramaID(task.DramaID)
    if !ok || source != sourceID || !mediaSlug.MatchString(slug) {
        return "", errors.New("invalid MissAV drama ID")
    }
    if task.Chapter.ID != core.ProviderChapterID(sourceID, slug, "1") {
        return "", errors.New("MissAV chapter does not match drama")
    }
    expected := baseURL + "/cn/" + url.PathEscape(slug)
    if task.Chapter.PageURL != expected {
        return "", errors.New("MissAV chapter page does not match drama")
    }
    return "/cn/" + url.PathEscape(slug), nil
}

// directMediaURL supports only publicly visible HTTP(S) media URLs.
// The original 110 dynamic player has not been reconstructed.
func directMediaURL(page, body string) (string, error) {
    normalized := strings.ReplaceAll(body, `\/`, "/")
    unique := make(map[string]bool)
    add := func(raw string) {
        parsed := core.NormalizeProviderMediaURL(page, raw)
        if core.IsProviderHTTPMediaURL(parsed) {
            unique[parsed] = true
        }
    }
    for _, match := range directVideoAttribute.FindAllStringSubmatch(normalized, -1) {
        if len(match) > 1 { add(match[1]) }
    }
    for _, match := range directPlayerJSON.FindAllStringSubmatch(normalized, -1) {
        if len(match) > 1 { add(match[1]) }
    }
    if len(unique) == 0 {
        return "", errors.New("MissAV media URL not directly exposed by player; dynamic playback unavailable")
    }
    if len(unique) != 1 {
        return "", errors.New("MissAV player exposes ambiguous media URLs")
    }
    for candidate := range unique { return candidate, nil }
    return "", errors.New("MissAV media URL unavailable")
}

func (p *provider) ResolveMedia(ctx context.Context, task core.Task) (core.ProviderMedia, error) {
    path, err := mediaPageForTask(task)
    if err != nil { return core.ProviderMedia{}, err }
    page := baseURL + path
    body, err := p.fetch(ctx, path)
    if err != nil { return core.ProviderMedia{}, err }
    media, err := directMediaURL(page, body)
    if err != nil { return core.ProviderMedia{}, err }
    return core.ProviderMedia{URL: media, Referer: page}, nil
}
