#!/usr/bin/env python3
"""Read-only public route diagnostics for jiaoguo. Never fabricate catalog availability."""
import json, pathlib, urllib.request, urllib.error, re
base="https://haijiao126.cc"
paths=["/","/index.php/vod/show/id/1.html","/index.php/vod/type/id/1.html","/vod/","/vodtype/1.html"]
out=pathlib.Path("jiaoguo-probe");out.mkdir(exist_ok=True)
result=[]
for path in paths:
 u=base+path
 req=urllib.request.Request(u,headers={"User-Agent":"Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/124.0 Mobile Safari/537.36","Accept":"text/html,application/xhtml+xml"})
 try:
  with urllib.request.urlopen(req,timeout=9) as r:
   body=r.read(600000).decode("utf-8","replace");status=r.status;final=r.url;ctype=r.headers.get("Content-Type","")
  examples=re.findall(r'href=["\']([^"\']*(?:voddetail|vodplay|/video/|/post/|/archives/)[^"\']*)["\']',body,re.I)[:10]
  entry={"path":path,"status":status,"url":final,"content_type":ctype,"size":len(body),"title":re.findall(r"<title[^>]*>([^<]+)",body,re.I)[:1],"video_links":examples}
 except Exception as e:entry={"path":path,"error":str(e)[:350]}
 print(json.dumps(entry,ensure_ascii=False),flush=True);result.append(entry)
(out/"report.json").write_text(json.dumps(result,ensure_ascii=False,indent=2))
