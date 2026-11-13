package missavrecovery

import "testing"

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
