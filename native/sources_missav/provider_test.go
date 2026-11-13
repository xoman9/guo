package missavrecovery

import ("testing"; "duanjuapp/native/core")

func TestCategoryValidation(t *testing.T) {
 for _,x:=range []string{"","new","romance","ai-genre"} {
  if !validCategory(x){t.Fatalf("expected valid %q",x)}
 }
 for _,x:=range []string{"../evil","other/a","spaces x","x?foo=1"} {
  if validCategory(x){t.Fatalf("unexpected category %q",x)}
 }
}
func TestCardExtractionAndDeduplication(t *testing.T) {
 raw:=`<a href="/cn/abc-123"><img alt="Sample title" data-src="/img/a.jpg"></a><a href="/cn/abc-123">Repeat</a><a href="/cn/new">New</a>`
 rows:=parseCards(raw)
 if len(rows)!=1||rows[0].SourceID!="abc-123"||rows[0].Title!="Sample title" {
  t.Fatalf("unexpected rows: %+v",rows)
 }
}

func TestMissAVMediaPageMatchesDramaAndChapter(t *testing.T) {
    drama := "abc-123"
    task := core.Task{
        DramaID: core.ProviderDramaID(sourceID, drama),
        Chapter: core.Chapter{
            ID: core.ProviderChapterID(sourceID, drama, "1"),
            PageURL: baseURL + "/cn/" + drama,
        },
    }
    path, err := mediaPageForTask(task)
    if err != nil || path != "/cn/abc-123" { t.Fatalf("valid task: %q %v", path, err) }
    other := task
    other.Chapter.PageURL = baseURL + "/cn/another"
    if _, err := mediaPageForTask(other); err == nil { t.Fatal("mismatched page accepted") }
    other = task
    other.Chapter.ID = core.ProviderChapterID(sourceID, "another", "1")
    if _, err := mediaPageForTask(other); err == nil { t.Fatal("mismatched chapter accepted") }
    other = task
    other.DramaID = core.ProviderDramaID("other", drama)
    if _, err := mediaPageForTask(other); err == nil { t.Fatal("wrong source accepted") }
}

func TestMissAVDirectMediaURL(t *testing.T) {
    page := baseURL + "/cn/abc-123"
    for _, body := range []string{
        `<video src="https://cdn.example.test/clip.mp4"></video>`,
        `{"src":"https:\/\/cdn.example.test\/clip.mp4"}`,
        `<source src="https://cdn.example.test/hls/main.m3u8?token=test">`,
    } {
        got, err := directMediaURL(page, body)
        if err != nil || got == "" { t.Fatalf("URL not parsed in %q: %v", body, err) }
    }
    for _, body := range []string{
        `<script>player.loadDynamic()</script>`,
        `<source src="javascript:alert(1)">`,
    } {
        if got, err := directMediaURL(page, body); err == nil {
            t.Fatalf("non-public direct URL accepted %q: %q", body, got)
        }
    }
}
