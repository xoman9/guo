package com.guoapp

import android.app.Activity
import android.os.Handler
import android.os.Looper
import android.webkit.JavascriptInterface
import android.webkit.WebView
import android.webkit.WebViewClient
import android.net.Uri
import android.webkit.WebResourceRequest
import android.webkit.WebResourceError
import android.webkit.WebResourceResponse
import io.flutter.plugin.common.MethodChannel
import org.json.JSONArray
import org.json.JSONObject

// Runs only bundled, inspected TVBox csp_GM user scripts on their first-party
// sites. The site still decides whether the WebView may access its pages.
internal class TVBoxJSBridge(private val activity: Activity) {
    private val handler = Handler(Looper.getMainLooper())
    private var active: WebView? = null
    private var pending: MethodChannel.Result? = null
    private var timedOut: Runnable? = null

    private fun finish(value: String? = null, error: String? = null) {
        val completion = pending ?: return
        pending = null
        timedOut?.let(handler::removeCallbacks)
        timedOut = null
        val view = active
        active = null
        if (view != null) {
            view.stopLoading()
            view.removeJavascriptInterface("GmSpiderInject")
            view.destroy()
        }
        if (error == null) completion.success(value)
        else completion.error("tvbox_js", error, null)
    }

    fun close() { finish(error = "JS 執行已取消") }

    fun execute(source: String, method: String, args: List<Any?>, result: MethodChannel.Result) {
        if (pending != null) {
            result.error("tvbox_js_busy", "另一個 JavaScript 站源請求仍在執行", null)
            return
        }
        val allowedMethods = setOf(
            "homeContent", "categoryContent", "searchContent", "detailContent", "playerContent")
        val host = when (source) {
            "missav" -> "missav.ws"
            "tvbox-netflav" -> "netflav.com"
            else -> {
                result.error("tvbox_js_source", "不支援的 JavaScript 站源", null)
                return
            }
        }
        if (method !in allowedMethods) {
            result.error("tvbox_js_method", "不支援的 JS 操作", null)
            return
        }
        val scriptName = if (source == "missav") "missav" else "netflav"
        val script = try {
            activity.assets.open("flutter_assets/assets/tvbox/js/$scriptName.user.js")
                .bufferedReader().use { it.readText() }
        } catch (error: Exception) {
            result.error("tvbox_js_missing", "缺少內建 $scriptName JavaScript：${error.message}", null)
            return
        }
        val jquery = try {
            activity.assets.open("flutter_assets/assets/tvbox/js/jquery.slim.min.js")
                .bufferedReader().use { it.readText() }
        } catch (error: Exception) {
            result.error("tvbox_js_missing", "缺少內建 jQuery：${error.message}", null)
            return
        }
        val url = try { route(source, method, args) }
        catch (error: Exception) {
            result.error("tvbox_js_route", error.message, null)
            return
        }
        pending = result
        val requestArgs = JSONArray().put(method)
        args.forEach { requestArgs.put(it ?: JSONObject.NULL) }
        val view = WebView(activity)
        active = view
        view.settings.javaScriptEnabled = true
        view.settings.domStorageEnabled = true
        view.settings.javaScriptCanOpenWindowsAutomatically = false
        view.settings.allowFileAccess = false
        view.settings.allowContentAccess = false
        view.settings.setSupportMultipleWindows(false)
        android.webkit.CookieManager.getInstance().setAcceptThirdPartyCookies(view, false)
        val api = object {
            @JavascriptInterface
            fun GetSpiderArgs(): String = requestArgs.toString()
            @JavascriptInterface
            fun SetSpiderResult(payload: String) {
                handler.post {
                    if (active === view && pending != null) {
                        if (payload.isBlank()) {
                            finish(error = "$scriptName JS 未回傳資料")
                        } else if (method != "playerContent" || !payload.contains("\"type\":\"match\"")) {
                            finish(value = payload)
                        }
                        // TVBox {type:match} means observe a REAL media request.
                        // It is not a playable URL and must not be returned as one.
                    }
                }
            }
        }
        view.addJavascriptInterface(api, "GmSpiderInject")
        var injected = false
        view.webViewClient = object : WebViewClient() {
            override fun shouldInterceptRequest(v: WebView?, request: WebResourceRequest?): WebResourceResponse? {
                if (method == "playerContent" && request != null) {
                    val uri: Uri = request.url
                    val path = uri.path.orEmpty().lowercase()
                    val media = path.endsWith("/master.m3u8") ||
                        path.endsWith("/playlist.m3u8") || path.endsWith("/v.mp4")
                    if (media && uri.scheme == "https" && !uri.host.isNullOrEmpty()) {
                        val address = uri.toString()
                        handler.post {
                            if (active === view && pending != null) {
                                finish(value = JSONObject().put("url", address).toString())
                            }
                        }
                    }
                }
                return null
            }
            override fun shouldOverrideUrlLoading(v: WebView?, request: WebResourceRequest?): Boolean {
                val uri = request?.url ?: return true
                val name = uri.host.orEmpty()
                return uri.scheme != "https" || (name != host && !name.endsWith(".$host"))
            }
            override fun onReceivedHttpError(v: WebView?, request: WebResourceRequest?,
                response: WebResourceResponse?) {
                if (request?.isForMainFrame == true &&
                    (response?.statusCode == 403 || response?.statusCode == 429)) {
                    handler.post { finish(error = "$scriptName HTTP ${response?.statusCode}：網站拒絕請求，請稍後再試或自行完成網站驗證") }
                }
            }
            override fun onReceivedError(v: WebView?, request: WebResourceRequest?,
                error: WebResourceError?) {
                if (request?.isForMainFrame == true) {
                    handler.post { finish(error = "$scriptName 網頁載入失敗：${error?.description}") }
                }
            }
            override fun onPageFinished(v: WebView?, url: String?) {
                if (active !== view || injected) return
                injected = true
                val boot = """
                    window.unsafeWindow = window;
                    window.GM_info = {script:{name:${JSONObject.quote(scriptName)}}};
                    window.GM_toastLong = function(){};
                """.trimIndent()
                // Load inspected local jQuery before invoking the bundled
                // user script. This does not depend on an external CDN.
                view.evaluateJavascript(boot) {
                    if (active === view) {
                        view.evaluateJavascript(jquery) {
                            if (active === view) view.evaluateJavascript(script, null)
                        }
                    }
                }
            }
        }
        val timer = Runnable { finish(error = "$scriptName JS 執行逾時；可能需要網站驗證或網頁已改版") }
        timedOut = timer
        handler.postDelayed(timer, 30000)
        view.loadUrl(url)
    }

    private fun route(source: String, method: String, args: List<Any?>): String {
        fun param(i: Int) = (args.getOrNull(i)?.toString() ?: "")
        fun mediaId(): String {
            val first = args.getOrNull(0)
            return if (first is List<*>) first.firstOrNull()?.toString().orEmpty()
                   else param(0)
        }
        fun number(i: Int) = (param(i).toIntOrNull() ?: 1).coerceIn(1, 10000)
        fun encode(input: String) = java.net.URLEncoder.encode(input, "UTF-8")
        fun slug(input: String): String {
            require(input.matches(Regex("[a-zA-Z0-9_-]{1,160}"))) { "不合法的影片 ID" }
            return input
        }
        if (source == "missav") {
            return when(method) {
                "homeContent" -> "https://missav.ws/cn/"
                "categoryContent" -> {
                    val tid = param(0)
                    require(tid.matches(Regex("[a-zA-Z0-9_/-]{1,120}")) && !tid.contains("..")) { "不合法的分類" }
                    "https://missav.ws/cn/$tid?page=${number(1)}"
                }
                "detailContent", "playerContent" -> "https://missav.ws/cn/${slug(mediaId())}"
                "searchContent" -> "https://missav.ws/cn/search/${encode(param(0))}?page=${number(2)}"
                else -> error("Unsupported")
            }
        }
        return when(method) {
            "homeContent" -> "https://netflav.com/trending"
            "categoryContent" -> {
                val category = param(0)
                val path = category.substringBefore("?")
                require(path.matches(Regex("[a-zA-Z0-9_-]{1,80}"))) { "不合法的分類路徑" }
                val builder = Uri.Builder().scheme("https").authority("netflav.com").appendPath(path)
                val query = category.substringAfter("?", "")
                if (query.isNotEmpty()) {
                    for (part in query.split("&")) {
                        val key = part.substringBefore("=")
                        val value = part.substringAfter("=", "")
                        require(key.matches(Regex("[a-zA-Z0-9_-]{1,32}"))) { "不合法的分類篩選參數" }
                        builder.appendQueryParameter(key, value)
                    }
                }
                builder.appendQueryParameter("page", number(1).toString()).build().toString()
            }
            "detailContent" -> "https://netflav.com/video?id=${encode(slug(mediaId()))}/"
            "playerContent" -> "https://netflav.com/video?id=${encode(slug(mediaId()))}#${number(1)}"
            "searchContent" -> "https://netflav.com/search?keyword=${encode(param(0))}&page=${number(2)}&type=title"
            else -> error("Unsupported")
        }
    }
}
