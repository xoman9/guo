package core

import (
    "context"
    "testing"
)

// An embedded Host interface satisfies the contract; network methods are
// deliberately unavailable. This guarantees module startup is offline.
type officialOfflineHost struct{ Host }

func TestOfficialJSModulesLoadWithoutNetwork(t *testing.T) {
    ids := []string{"heitao","huangju","jinguoduanju","dsd","beabox","yeguo","huangguo-video","huangguoai"}
    if len(officialModules) != len(ids) {
        t.Fatalf("got %d pinned modules, want %d",len(officialModules),len(ids))
    }
    for _, id := range ids {
        t.Run(id,func(t *testing.T) {
            module:=officialModules[id]
            if module==nil || module.manifest.ID!=id || module.manifest.Version!="1.0.1" {
                t.Fatalf("missing or unpinned JS module: %s",id)
            }
            provider:=module.provider(&officialOfflineHost{}).(*jsProvider)
            var accepted bool
            if err:=provider.call(context.Background(),"validCategory",&accepted,"__invalid_category__"); err!=nil {
                t.Fatalf("%s VM start or JS function call failed: %v",id,err)
            }
            if accepted { t.Fatal("invalid category was accepted") }
            if provider.module!=module { t.Fatal("provider lost module identity") }
        })
    }
}
