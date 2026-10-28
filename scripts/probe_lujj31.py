#!/usr/bin/env python3
"""One-shot, read-only diagnostic for the configured lujj31 website.

Never downloads media or prints full HTML, URLs with tokens, or response bodies.
"""
import json
import re
import socket
from urllib.error import HTTPError, URLError
from urllib.parse import urlsplit
from urllib.request import Request, urlopen

HOST = "lujj31.lol"
ORIGIN = "https://" + HOST
AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/130.0.0.0 Safari/537.36"

try:
    ips = sorted({row[4][0] for row in socket.getaddrinfo(HOST, 443, type=socket.SOCK_STREAM)})
    print("DNS resolved:", ips[:8])
except OSError as exc:
    print("DNS failed:", type(exc).__name__, str(exc)[:150])


def probe(path):
    target = ORIGIN + path
    print("\nCHECK path:", path)
    req = Request(target, headers={"User-Agent": AGENT, "Accept": "application/json,text/html,*/*"})
    try:
        response = urlopen(req, timeout=12)
    except HTTPError as exc:
        response = exc
    except (URLError, TimeoutError, OSError) as exc:
        print("NETWORK ERROR:", type(exc).__name__, str(exc)[:180])
        return
    with response:
        body = response.read(96 * 1024)
        content_type = response.headers.get("Content-Type", "")[:100]
        final = urlsplit(response.url)
        print("STATUS:", response.status, "FINAL HOST:", final.hostname, "PATH:", final.path)
        print("CONTENT TYPE:", content_type, "SAMPLED BYTES:", len(body))
        try:
            parsed = json.loads(body)
            if isinstance(parsed, dict):
                print("JSON keys:", sorted(parsed)[:25])
                for field in ("class", "list", "data"):
                    value = parsed.get(field)
                    if isinstance(value, list):
                        print("JSON", field, "count:", len(value))
                    elif isinstance(value, dict):
                        print("JSON", field, "keys:", sorted(value)[:12])
                print("JSON pagecount:", parsed.get("pagecount", "(absent)"))
            else:
                print("JSON root type:", type(parsed).__name__)
            return
        except (ValueError, UnicodeDecodeError):
            pass
        html = body.decode("utf-8", errors="replace")
        print("HTML title:", re.sub(r"\\s+", " ", re.search(r"<title[^>]*>(.*?)</title>", html, re.I | re.S).group(1))[:120] if re.search(r"<title[^>]*>(.*?)</title>", html, re.I | re.S) else "(none)")
        print("MacCMS player data:", bool(re.search(r"player_(?:aaaa|data)\\s*=", html)))
        print("MacCMS API path:", bool(re.search(r"api\\.php/provide/vod", html, re.I)))
        print("MacCMS vod links:", bool(re.search(r"vod/(?:detail|play|type)/|voddetail|vodplay|/index\\.php/vod/", html, re.I)))
        print("JavaScript challenge markers:", bool(re.search(r"cf-chl|challenge-platform|captcha|just a moment", html, re.I)))
        hrefs = re.findall(r"""href=["']([^"']+)["']""", html, re.I)
        path_shapes = []
        for link in hrefs:
            parsed_link = urlsplit(link)
            if parsed_link.path and not any(s == parsed_link.path for s in path_shapes):
                path_shapes.append(parsed_link.path[:90])
            if len(path_shapes) >= 12:
                break
        print("Example link PATHS only:", path_shapes)


probe("/")
probe("/api.php/provide/vod/?ac=list")
probe("/api.php/provide/vod/?ac=detail&pg=1")
