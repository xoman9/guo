#!/usr/bin/env python3
import urllib.request,re,json,html,urllib.parse,os,pathlib
base="https://chengguodj.com"
out=pathlib.Path("chengguo-js-proof");out.mkdir(exist_ok=True)
def get(url):
 req=urllib.request.Request(url,headers={"User-Agent":"Mozilla/5.0","Referer":base+"/"})
 with urllib.request.urlopen(req,timeout=20) as r:return r.read(3_000_000).decode("utf-8","replace")
page=get(base+"/manju")
scripts=re.findall(r'<script[^>]+src=["\']([^"\']+)["\']',page,re.I)
print("SCRIPT_COUNT",len(scripts),flush=True)
terms=re.compile(r"decrypt|encrypted|fallback_url|CryptoJS|createDecipher|AES|cbc|ecb|nihmcl|imgDecrypt|decryptImage|decryptCover|cover\.url",re.I)
results=[]
for name in scripts[:75]:
 u=urllib.parse.urljoin(base,name)
 if not u.startswith((base+"/","https://chengguodj.com/")):continue
 try:
  content=get(u)
  matches=[m for m in terms.finditer(content)]
  snippets=[content[max(0,m.start()-250):min(len(content),m.end()+450)] for m in matches[:20]]
  if snippets:
   results.append({"url":u,"length":len(content),"hits":len(matches),"snippets":snippets})
  print("SCANNED",u,len(content),"MATCHES",len(matches),flush=True)
 except Exception as e:print("ERROR",u,str(e)[:180],flush=True)
(out/"report.json").write_text(json.dumps(results,ensure_ascii=False,indent=2))
print("CANDIDATE_FILES",len(results))
