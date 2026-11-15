// ==UserScript==
// @name         MissAV
// @namespace    gmspider
// @version      2024.12.03
// @description  MissAV GMSpider
// @author       Luomo
// @match        https://missav.*/*
// @require      https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.slim.min.js
// @grant        unsafeWindow
// ==/UserScript==
console.log(JSON.stringify(GM_info));
(function () {
    const GMSpiderArgs = {};
    if (typeof GmSpiderInject !== 'undefined') {
        let args = JSON.parse(GmSpiderInject.GetSpiderArgs());
        GMSpiderArgs.fName = args.shift();
        GMSpiderArgs.fArgs = args;
    } else {
        GMSpiderArgs.fName = "detailContent";
        GMSpiderArgs.fArgs = [true];
    }
    Object.freeze(GMSpiderArgs);
    const GmSpider = (function () {
        const filter = {
            key: "filter",
            name: "过滤",
            value: [{
                n: "所有",
                v: ""
            }, {
                n: "单人作品",
                v: "&filters=individual"
            }, {
                n: "多人作品",
                v: "&filters=multiple"
            }, {
                n: "中文字幕",
                v: "&filters=chinese-subtitle"
            }]
        };
        const filterWithoutSort = [
            filter
        ];
        const defaultFilter = [
            filter,
            {
                key: "sort",
                name: "排序方式",
                value: [
                    {
                        n: "发行日期",
                        v: "&sort=released_at"
                    },
                    {
                        n: "最近更新",
                        v: "&sort=published_at"
                    },
                    {
                        n: "收藏数",
                        v: "&sort=saved"
                    },
                    {
                        n: "今日浏览数",
                        v: "&sort=today_views"
                    },
                    {
                        n: "本周浏览数",
                        v: "&sort=weekly_views"
                    },
                    {
                        n: "本月浏览数",
                        v: "&sort=monthly_views"
                    },
                    {
                        n: "总浏览数",
                        v: "&sort=views"
                    }
                ]
            }];

        function pageList(result) {
            result.pagecount = parseInt($("#price-currency").text().replace(/[^0-9]/ig, ""));
            result.total = result.pagecount * result.limit;
            $(".gap-5 .thumbnail").each(function (i) {
                result.list.push({
                    vod_id: ($(this).find(".text-secondary").attr("alt") || ($(this).find(".text-secondary").attr("href") || "").split("?")[0].replace(/\/$/, "").split("/").pop()),
                    vod_name: $(this).find(".text-secondary").text().trim(),
                    vod_pic: $(this).find("img").data("src"),
                    vod_year: $(this).find(".right-1").text().trim(),
                    vod_remarks: $(this).find(".left-1").text().trim(),
                })
            });
            return result;
        }

        function categoryList(result) {
            $(".gap-4 div").each(function () {
                result.list.push({
                    vod_id: getCategoryFromUrl($(this).find(".text-nord13").attr("href")),
                    vod_name: $(this).find(".text-nord13").text().trim(),
                    vod_remarks: $(this).find(".text-nord10 a").text().trim(),
                    vod_tag: "folder",
                    style: {
                        "type": "rect",
                        "ratio": 2
                    }
                })
            });
            result.limit = 36;
            result.pagecount = parseInt($("#price-currency").text().replace(/[^0-9]/ig, ""));
            result.total = result.pagecount * result.limit;
            return result;
        }

        function getCategoryFromUrl(url) {
            return url.split('/cn/').at(1);
        }

        function formatDetail(detail, ...keys) {
            let format = "";
            for (let key of keys) {
                format += key in detail ? (Array.isArray(detail[key]) ? detail[key].join(" ") : detail[key]) : "";
            }
            return format;
        }

        return {
            homeContent: function (filter) {
                let result = {
                    class: [
                        {type_id: "new", type_name: "所有影片"},
                        {type_id: "madou", type_name: "麻豆传媒"},
                        {type_id: "chinese-subtitle", type_name: "中文字幕"},
                        {type_id: "uncensored-leak", type_name: "无码流出"},
                        {type_id: "actresses/ranking", type_name: "热门女优"},
                        {type_id: "makers", type_name: "发行商"},
                        {type_id: "genres", type_name: "类型"},
                    ],
                    filters: {
                        "new": defaultFilter,
                        "madou": defaultFilter,
                        "chinese-subtitle": defaultFilter,
                        "uncensored-leak": defaultFilter,
                        "actresses/ranking": defaultFilter,
                        "makers": defaultFilter,
                        "genres": defaultFilter
                    },
                    list: []
                };
                $(".gap-5:eq(5) .thumbnail").each(function () {
                    result.list.push({
                        vod_id: ($(this).find(".text-secondary").attr("alt") || ($(this).find(".text-secondary").attr("href") || "").split("?")[0].replace(/\/$/, "").split("/").pop()),
                        vod_name: $(this).find(".text-secondary").text().trim(),
                        vod_pic: $(this).find("img").data("src"),
                        vod_year: $(this).find(".absolute").text().trim()
                    })
                });
                console.log(result);
                return result;
            },
            categoryContent: function (tid, pg, filter, extend) {
                let result = {
                    list: [],
                    limit: 12,
                    total: 0,
                    page: pg,
                    pagecount: 0
                };
                if (tid === "actresses/ranking") {
                    $(".gap-4 .space-y-4").each(function () {
                        result.list.push({
                            vod_id: getCategoryFromUrl($(this).find(".space-y-2 a").attr("href")),
                            vod_name: $(this).find(".truncate").text().trim(),
                            vod_pic: $(this).find("img").length > 0 ? $(this).find("img").attr("src") : "",
                            vod_remarks: $(this).find(".text-sm").text().trim(),
                            vod_tag: "folder",
                            style: {
                                "type": "rect",
                                "ratio": 1
                            }
                        })
                    });
                    result.limit = 100;
                    result.total = 100;
                    result.pagecount = 1;
                } else if (tid === "makers") {
                    function getNavs(name) {
                        $("nav.hidden .relative a.group span:contains('" + name + "')").parents(".relative:first").find(".py-1 a").each(function () {
                            result.list.push({
                                vod_id: getCategoryFromUrl($(this).attr("href")),
                                vod_name: $(this).text().trim(),
                                vod_remarks: name,
                                vod_tag: "folder",
                                style: {
                                    "type": "rect",
                                    "ratio": 2
                                }
                            })
                        })
                    }

                    if (pg == 1) {
                        getNavs("国产 AV");
                        getNavs("无码影片");
                        getNavs("素人");
                    }
                    result = categoryList(result)
                } else if (tid === "genres") {
                    result = categoryList(result)
                } else {
                    result = pageList(result);
                }
                return result;
            },
            detailContent: function (ids) {
                let detail = {};
                $(".space-y-2:not(.list-disc) .text-secondary").each(function () {
                    const key = $(this).find("span:first").text().replace(":", "");
                    if ($(this).find("a").length === 0) {
                        detail[key] = $(this).find("span:first").remove().end().text().trim();
                    } else {
                        detail[key] = [];
                        $(this).find("a").each(function () {
                            const id = getCategoryFromUrl($(this).attr("href"));
                            const name = $(this).text();
                            detail[key].push(`[a=cr:{"id":"${id}","name":"${name}"}/]${name}[/a]`);
                        })
                    }
                });

                console.log($('a.items-center:contains("显示更多")'));
                const movieId = String(ids[0] || "").trim();
                // The old spider used only the ID as the display title.
                // A broken/short ID could produce a literal "[" in the
                // player header. Prefer page metadata when it is available.
                const pageTitle = (
                    $('head meta[property="og:title"]').attr("content") ||
                    $('head meta[name="description"]').attr("content") ||
                    $("h1").first().text() ||
                    ""
                ).trim();
                const movieTitle = pageTitle.length > 1 ? pageTitle :
                    (movieId.length > 1 ? movieId.toUpperCase() : "MissAV");
                const vod = {
                    vod_id: movieId,
                    vod_name: movieTitle,
                    vod_pic: $("head link[as=image]").attr("href"),
                    vod_year: $("#space-y-2 time").text(),
                    vod_remarks: formatDetail(detail, "类型"),
                    vod_actor: formatDetail(detail, "女优"),
                    vod_content: $('a.items-center:contains("显示更多")').length > 0 ? $("head meta[name=description]").attr("content") : $("head meta[property='og:title']").attr("content"),
                    vod_play_from: "MissAV",
                    vod_play_url: (typeof hls !== "undefined" && hls && hls.url) ? "多视轨$" + hls.url : "",
                };
                console.log({list: [vod]})
                return {list: [vod]};
            },
            searchContent: function (key, quick, pg) {
                let result = {
                    list: [],
                    limit: 12,
                    total: 0,
                    page: pg,
                    pagecount: 0
                };
                result = pageList(result);
                return result;
            }
        };
    })();
    $(document).ready(function () {
        let result = "";
        if ($("#cf-wrapper").length > 0) {
            console.log("源站不可用:" + $('title').text());
            GM_toastLong("源站不可用:" + $('title').text());
        } else {
            result = GmSpider[GMSpiderArgs.fName](...GMSpiderArgs.fArgs);
        }
        console.log(result);
        if (typeof GmSpiderInject !== 'undefined') {
            // The official TVBox MissAV spider returns its HLS URL through
            // detailContent, not playerContent. On Android WebView the page's
            // player may initialize AFTER DOM ready; wait for that actual URL
            // instead of immediately returning an unplayable empty value.
            if (GMSpiderArgs.fName === "detailContent" && result && result.list && result.list[0]) {
                const vod = result.list[0];
                const publish = () => GmSpiderInject.SetSpiderResult(JSON.stringify(result));
                let attempts = 0;
                const checkHls = () => {
                    let candidate = "";
                    try {
                        if (typeof hls !== "undefined" && hls && typeof hls.url === "string") {
                            candidate = hls.url;
                        }
                    } catch (e) {}
                    try {
                        const parsed = new URL(candidate, window.location.href);
                        if (parsed.protocol === "https:" &&
                            /\.(?:m3u8|mp4)$/i.test(parsed.pathname)) {
                            vod.vod_play_url = "多视轨$" + parsed.href;
                            publish();
                            return;
                        }
                    } catch (e) {}
                    if (++attempts < 30) {
                        setTimeout(checkHls, 500);
                    } else {
                        // Keep details accessible: Android uses the actual
                        // first-party playback page as its fallback resolver.
                        publish();
                    }
                };
                if (vod.vod_play_url) publish();
                else checkHls();
            } else {
                GmSpiderInject.SetSpiderResult(JSON.stringify(result));
            }
        }
    });
})();