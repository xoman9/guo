package core

import (
 "embed"
 "fmt"
 "io/fs"
 "path"
)

//go:embed all:bundled
var officialModuleFiles embed.FS

var officialModules = func() map[string]*jsModule {
  result := map[string]*jsModule{}
  for _, id := range []string{"heitao","huangju","jinguoduanju","dsd","beabox","yeguo","huangguo-video","huangguoai","hongguo"} {
    root := path.Join("bundled", id)
    manifest, err := fs.ReadFile(officialModuleFiles,path.Join(root,"manifest.json"))
    if err != nil { panic(err) }
    code, err := fs.ReadFile(officialModuleFiles,path.Join(root,"main.js"))
    if err != nil { panic(err) }
    module, err := parseModule(manifest,code)
    if err != nil {panic(fmt.Errorf("JS %s: %w",id,err))}
    if module.manifest.ID != id {panic("JS id mismatch: "+id)}
    result[id]=module
  }
  return result
}()
