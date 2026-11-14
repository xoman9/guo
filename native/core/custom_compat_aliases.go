package core

// Compatibility aliases for site adapters carried forward from custom.
// Original upstream Category and ProviderMedia implementations are unchanged.
type nativeCategory = Category
type providerMedia = ProviderMedia
type nativeDrama = LibraryDrama
const (
 sourceLujj31 = "lujj31"
 sourceLiuguo = "liuguo"
 sourceXiaoguo = "xiaoguo"
)
var webProviderNumericID = WebProviderNumericID
var webProviderInteger = WebProviderInteger
var providerHTMLAttr = ProviderHTMLAttr
var providerHTMLClass = ProviderHTMLClass
var providerHTMLNodes = ProviderHTMLNodes
var providerHTMLFirstClass = ProviderHTMLFirstClass
var providerHTMLText = ProviderHTMLText
var nativeText = NativeText
var nativeNormalize = NativeNormalize
var mergeNativeDrama = MergeNativeDrama
var nativeDramaScore = NativeDramaScore
var sortProviderChapters = SortProviderChapters
var chapterEpisodeNumber = ChapterEpisodeNumber
var rawEpisode = RawEpisode
var episodeIndex = EpisodeIndex
var extractPageTitle = ExtractPageTitle
var extractDescription = ExtractDescription
var extractAttr = ExtractAttr
var extractTags = ExtractTags
var cleanText = CleanText
var firstNonEmpty = FirstNonEmpty
var mapString = MapString
var mapStringSlice = MapStringSlice
var resolveProviderURL = ResolveProviderURL
var normalizeProviderMediaURL = NormalizeProviderMediaURL
var isProviderHTTPMediaURL = IsProviderHTTPMediaURL
var providerReleaseDate = ProviderReleaseDate
var providerChinaTime = ProviderChinaTime
var providerSourceForURL = ProviderSourceForURL
var nativeSearchText = NativeSearchText
var truncate = Truncate
var providerCoverAddress = ProviderCoverAddress
var validNativeCoverURL = ValidNativeCoverURL
var providerDramaID = ProviderDramaID
var providerChapterID = ProviderChapterID
var splitProviderDramaID = SplitProviderDramaID
