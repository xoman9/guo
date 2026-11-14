package core

import "testing"

// Regression: an APK donor may contain working parsers, but without BaseURL
// in the 110 Registry none can fetch real categories or a film catalog.
func TestRecoveredTaoLiuXiaoHaveRegisteredBaseURLs(t *testing.T) {
 cases:=[]struct{source, expected string}{
  {sourceLujj31,lujj31BaseURL},
  {sourceLiuguo,liuguoBaseURL},
  {sourceXiaoguo,xiaoguoBaseURL},
 }
 for _,tc:=range cases {
  desc,ok:=registeredDescriptor(tc.source)
  if !ok {t.Errorf("%s missing from registry",tc.source);continue}
  if desc.BaseURL==""||desc.BaseURL!=tc.expected {
   t.Errorf("%s: blank/wrong BaseURL %q, expected %q",tc.source,desc.BaseURL,tc.expected)
  }
  if !desc.DefaultOff{t.Errorf("%s must be a + source",tc.source)}
  if !desc.NoHostRotation{t.Errorf("%s must remain pinned to its recovered site",tc.source)}
 }
}
