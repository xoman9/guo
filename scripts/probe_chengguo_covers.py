#!/usr/bin/env python3
"""Verify real Chengguo poster bytes, not just HTML or URL strings."""
import json, re, urllib.request, urllib.parse, pathlib, hashlib, struct
from html.parser import HTMLParser

ROOT="https://chengguodj.com"
OUT=pathlib.Path("chengguo-cover-proof")
OUT.mkdir(exist_ok=True)
UA="Mozilla/5.0 (Android 14; Mobile) AppleWebKit/537.36 Chrome/124.0 Mobile Safari/537.36"
def get(url):
    req=urllib.request.Request(url,headers={"User-Agent":UA,"Referer":ROOT+"/","Accept":"image/avif,image/webp,image/apng,image/*,*/*;q=0.8"})
    with urllib.request.urlopen(req,timeout=18) as f:
        return f.read(5_000_000),dict(f.headers),f.url
def kind(b):
    if b.startswith(b"\x89PNG\r\n\x1a\n"):return "png"
    if b.startswith(b"\xff\xd8\xff"):return "jpeg"
    if b.startswith(b"GIF87a") or b.startswith(b"GIF89a"):return "gif"
    if b.startswith(b"RIFF") and b[8:12]==b"WEBP":return "webp"
    if len(b)>12 and b[4:8]==b"ftyp" and b[8:12] in (b"avif",b"avis"):return "avif"
    return None
class Parser(HTMLParser):
    def __init__(self):super().__init__();self.articles=0;self.img=[];self.links=[];self.inarticle=False
    def handle_starttag(self,tag,attrs):
        d=dict(attrs)
        if tag=="article" and d.get("data-xpch")=="card-drama":self.inarticle=True;self.articles+=1
        if self.inarticle and tag=="img":self.img.append(d)
        if self.inarticle and tag=="a" and "/drama/" in d.get("href",""):self.links.append(d["href"])
    def handle_endtag(self,tag):
        if tag=="article":self.inarticle=False
rows=[]; html=b""
for page in ("/manju","/browse"):
    try:
        b,headers,actual=get(ROOT+page);p=Parser();p.feed(b.decode("utf-8","replace"))
        rows.append({"page":page,"size":len(b),"articles":p.articles,"images":len(p.img),"dramas":len(p.links)})
        if len(p.img)>0 and not html:html=b;parsed=p
    except Exception as exc:rows.append({"page":page,"error":repr(exc)})
targets=[]
if html:
    for a in parsed.img[:25]:
        for attr in ("data-src","data-original","data-lazy","data-srcset","srcset","src"):
            raw=a.get(attr,"").strip()
            if raw.startswith("data:") or not raw:continue
            raw=raw.split(",")[0].strip().split(" ")[0]
            targets.append((attr,urllib.parse.urljoin(ROOT,raw)))
    for m in re.finditer(r'<script[^>]*id=["\']__NUXT_DATA__["\'][^>]*>(.*?)</script>',html.decode("utf-8","replace"),re.S|re.I):
        try:
            nodes=json.loads(m.group(1))
            for item in nodes:
                if not isinstance(item,dict) or "cover" not in item:continue
                j=item["cover"]
                if not isinstance(j,int) or j<0 or j>=len(nodes) or not isinstance(nodes[j],dict):continue
                obj=nodes[j]
                for key in ("url","fallback_url","thumbnail","thumbnail_url","poster"):
                    ix=obj.get(key)
                    raw=nodes[ix] if isinstance(ix,int) and 0<=ix<len(nodes) else ix
                    if isinstance(raw,str) and raw.startswith(("http://","https://","/")):
                        targets.append(("nuxt."+key,urllib.parse.urljoin(ROOT,raw)))
        except Exception as exc:rows.append({"nuxt_error":repr(exc)})
seen=set();proofs=[]
for attr,u in targets:
    if u in seen or len(seen)>=30:continue
    seen.add(u)
    try:
        b,h,final=get(u); k=kind(b)
        result={"source":attr,"url":u,"actual":final,"bytes":len(b),"content_type":h.get("Content-Type"),"magic":k,"sha256":hashlib.sha256(b).hexdigest()}
        if k and len(b)>2048:
            name=f"poster-{len(proofs):02d}.{k}"
            (OUT/name).write_bytes(b)
            result["file"]=name
        proofs.append(result)
    except Exception as exc:proofs.append({"source":attr,"url":u,"error":repr(exc)})
report={"pages":rows,"checked":len(proofs),"verified_images":[p for p in proofs if p.get("file")],"requests":proofs}
(OUT/"report.json").write_text(json.dumps(report,ensure_ascii=False,indent=2))
print(json.dumps({"pages":rows,"checked":len(proofs),"verified":len(report["verified_images"]),"samples":proofs[:12]},ensure_ascii=False))
if not report["verified_images"]:raise SystemExit("NO_VERIFIED_COVER: real image bytes unavailable; do not build APK")
