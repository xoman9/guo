//go:build zhenguojian

package main

import (
    "encoding/json"
    "path/filepath"
    "testing"

    "duanjuapp/native/core"
)

// This test executes the real NativeRequest initialize entry point after the
// source modules are blank-imported through sources_gen.go. A successful Go
// compile alone does not prove the original providers reach the Flutter picker.
func TestSingle110CoreIncludesOriginal15AndCustom13(t *testing.T) {
    dir := t.TempDir()
    request, err := json.Marshal(map[string]string{
        "action": "initialize",
        "directory": filepath.Join(dir, "app"),
        "downloads": filepath.Join(dir, "downloads"),
    })
    if err != nil { t.Fatal(err) }
    var response struct {
        OK bool `json:"ok"`
        Error string `json:"error"`
        Data struct {
            Flavor string `json:"flavor"`
            Sources []struct {
                ID string `json:"id"`
                Name string `json:"name"`
                DefaultOff bool `json:"defaultOff"`
                Available bool `json:"available"`
            } `json:"sources"`
        } `json:"data"`
    }
    if err := json.Unmarshal([]byte(core.NativeRequest(string(request))), &response); err != nil { t.Fatal(err) }
    if !response.OK { t.Fatalf("native initialize failed: %s", response.Error) }
    if response.Data.Flavor != "zhenguojian" { t.Fatalf("unexpected flavor: %q", response.Data.Flavor) }
    found := make(map[string]struct{ optional, available bool })
    for _, site := range response.Data.Sources {
        if _, duplicate := found[site.ID]; duplicate { t.Errorf("duplicate source ID %q", site.ID) }
        found[site.ID] = struct{optional, available bool}{site.DefaultOff, site.Available}
    }
    original := []string{
        "hongguo", "hanxiaoquan", "guipian", "sorani", "huangdou",
        "huangju", "yeguo", "dsd", "jinguoduanju", "jinguo",
        "heitao", "beabox", "huangguo-video", "huangguoai", "cloudfront",
    }
    optional := []string{
        "lujj31", "liuguo", "xiaoguo", "youguo", "meiguo", "yingguo",
        "luguo", "liguo", "juguo", "ningguo", "mangguo", "chengguodj", "jiaoguo",
    }
    for _, id := range original {
        site, ok := found[id]
        if !ok { t.Errorf("original native 110 provider missing: %s", id); continue }
        if !site.available { t.Errorf("original native 110 provider unexpectedly unavailable: %s", id) }
    }
    for _, id := range optional {
        site, ok := found[id]
        if !ok { t.Errorf("custom provider missing from Flutter source list: %s", id); continue }
        if !site.optional || !site.available {
            t.Errorf("custom provider must be available and default-disabled (+): %s (%+v)", id, site)
        }
    }
    if len(found) < len(original)+len(optional) {
        t.Errorf("expected 28 original/custom IDs, only %d are registered", len(found))
    }
}
