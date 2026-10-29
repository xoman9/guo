import urllib.request,urllib.error,json,re,pathlib
sites={"zhenguo":"https://zhenshi27.xyz","caoguo":"https://ccctv01.top"}
paths=["/","/vodtype/1.html","/index.php/vod/show/id/1.html","/index.php/vod/type/id/1.html","/api.php/provide/vod/","/api.php/provide/vod/?ac=list","/api.php/provide/vod/?ac=detail&pg=1","/vod/"]
rows=[]
for source,base in sites.items():
 for path in paths:
  try:
   req=urllib.request.Request(base+path,headers={"User-Agent":"Mozilla/5.0","Accept":"text/html,application/json"})
   with urllib.request.urlopen(req,timeout=8) as resp:
    b=resp.read(150000).decode("utf-8","replace")
    item={"source":source,"path":path,"status":resp.status,"url":resp.url,"size":len(b),"title":re.findall(r"<title[^>]*>([^<]+)",b,re.I)[:1],"video_links":re.findall(r'href=["\\\']([^"\\\']*(?:voddetail|vodplay|/play/)[^"\\\']*)',b,re.I)[:6],"json_prefix":b[:100] if b.lstrip().startswith(("{","[")) else ""}
  except Exception as e:item={"source":source,"path":path,"error":str(e)[:160]}
  rows.append(item);print(json.dumps(item,ensure_ascii=False),flush=True)
pathlib.Path("zhenguo-caoguo-probe").mkdir(exist_ok=True)
pathlib.Path("zhenguo-caoguo-probe/results.json").write_text(json.dumps(rows,ensure_ascii=False,indent=2))
