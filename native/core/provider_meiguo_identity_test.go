package core

import (
    "testing"
)

// The catalog card identifies the selected Meiguo video. A MacCMS player
// page can have a different h1/og:image even while its chapters are correct.
func TestMeiguoDetailPreservesSelectedCatalogTitleAndCover(t *testing.T) {
    catalog := nativeDrama{
        ID: "meiguo:72015", Source: "meiguo", SourceID: "72015",
        Title: "原片單片名", Cover: "https://shiresm.lol/covers/72015.jpg",
        Episodes: 1,
    }
    unrelated := nativeDrama{
        ID: catalog.ID, Source: catalog.Source, SourceID: catalog.SourceID,
        Title: "其他影片片名", Cover: "https://shiresm.lol/covers/88888.jpg",
        Episodes: 5,
    }
    got := mergeNativeDrama(catalog, preserveMeiguoCatalogIdentity(catalog, unrelated))
    if got.Title != catalog.Title || got.Cover != catalog.Cover {
        t.Fatalf("detail changed movie identity: %+v", got)
    }
    if got.Episodes != 5 {
        t.Fatalf("detail episode metadata was lost: %+v", got)
    }
}

func TestMeiguoCacheNotRewrittenWhenDetailContainsOtherMovie(t *testing.T) {
    catalog := nativeDrama{
        ID: "meiguo:72015", Source: "meiguo", SourceID: "72015",
        Title: "片單原始片名", Cover: "https://shiresm.lol/poster/72015.jpg",
    }
    engine := &nativeEngine{
        directory: t.TempDir(),
        catalogs: map[string][]nativeDrama{"meiguo": {catalog}},
    }
    wrongDetail := nativeDrama{
        ID: catalog.ID, Source: catalog.Source, SourceID: catalog.SourceID,
        Title: "其他片", Cover: "https://shiresm.lol/poster/another.jpg",
        Episodes: 9,
    }
    if err := engine.saveDetailMetadata(wrongDetail); err != nil {
        t.Fatal(err)
    }
    after := engine.catalogs["meiguo"][0]
    if after.Title != catalog.Title || after.Cover != catalog.Cover || after.Episodes != 9 {
        t.Fatalf("catalog mutated by detail: %+v", after)
    }
}

func TestMeiguoIdentityGuardDoesNotLeakAcrossSourcesOrFilmIDs(t *testing.T) {
    source := nativeDrama{ID: "meiguo:72", Source: "meiguo", SourceID: "72",
        Title: "片名", Cover: "https://example.org/72.jpg"}
    for _, different := range []nativeDrama{
        {ID: "meiguo:73", Source: "meiguo", SourceID: "73", Title: "片名73"},
        {ID: "ningguo:72", Source: "ningguo", SourceID: "72", Title: "其他頻道"},
    } {
        got := preserveMeiguoCatalogIdentity(source, different)
        if got.Title != different.Title || got.ID != different.ID {
            t.Fatalf("identity guard crossed movie/source boundary: %+v", got)
        }
    }
}
