package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// The URL is a site entry, not a verified API endpoint. MacCMS JSON is used
// only when the site actually exposes the standard public vod interface.
const lujj31BaseURL = "https://lujj31.lol"

func (d *Downloader) lujj31Request(ctx context.Context, params url.Values) (map[string]any, error) {
	site := d.providerBaseURL(sourceLujj31)
	address := site + "/api.php/provide/vod/?" + params.Encode()
	body, err := d.fetchProviderText(ctx, address, site+"/")
	if err != nil {
		return nil, fmt.Errorf("lujj31 公开目录接口不可用：%w", err)
	}
	if len(body) > 4<<20 {
		return nil, errors.New("lujj31 返回的目录资料过大")
	}
	var response map[string]any
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return nil, errors.New("lujj31 未返回标准 MacCMS JSON；需要确认网站实际接口")
	}
	if response == nil {
		return nil, errors.New("lujj31 返回了空的目录资料")
	}
	return response, nil
}

func lujj31Rows(value any) []map[string]any {
	raw, _ := value.([]any)
	if len(raw) > 500 {
		return nil
	}
	rows := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func lujj31Drama(row map[string]any, site string) Drama {
	id := mapString(row, "vod_id", "id")
	name := strings.TrimSpace(mapString(row, "vod_name", "name"))
	if !webProviderNumericID.MatchString(id) || name == "" {
		return Drama{}
	}
	cover := providerCoverAddress(row["vod_pic"], site+"/")
	description := mapString(row, "vod_content", "vod_blurb")
	category := mapString(row, "type_name")
	return Drama{
		ID: providerDramaID(sourceLujj31, id), Source: sourceLujj31, SourceID: id,
		Title: truncate(name, 512), Name: truncate(name, 512),
		Cover: cover, CoverURL: cover,
		Desc: truncate(description, 12000), Intro: truncate(description, 12000),
		CategoryName: truncate(category, 64), ChannelName: "lujj31",
		Remark: truncate(mapString(row, "vod_remarks"), 128),
	}
}

func (d *Downloader) fetchLujj31Categories(ctx context.Context) ([]nativeCategory, error) {
	response, err := d.lujj31Request(ctx, url.Values{"ac": {"list"}})
	if err != nil {
		return nil, err
	}
	entries := lujj31Rows(response["class"])
	categories := make([]nativeCategory, 0, len(entries))
	seen := map[string]bool{}
	for _, row := range entries {
		id := mapString(row, "type_id")
		name := strings.TrimSpace(mapString(row, "type_name"))
		if !webProviderNumericID.MatchString(id) || name == "" || len([]rune(name)) > 48 || seen[id] {
			continue
		}
		seen[id] = true
		categories = append(categories, nativeCategory{ID: id, Name: name})
	}
	if len(categories) == 0 {
		return nil, errors.New("lujj31 未提供可识别的内容分类，尚不能确认此接口适用")
	}
	return categories, nil
}

func (d *Downloader) fetchLujj31CatalogPage(ctx context.Context, page int, category, query string) ([]Drama, bool, error) {
	if page < 1 || page > 100000 || len(query) > 256 || (category != "" && !webProviderNumericID.MatchString(category)) {
		return nil, false, errors.New("lujj31 目录参数无效")
	}
	params := url.Values{"ac": {"list"}, "pg": {strconv.Itoa(page)}}
	if category != "" {
		params.Set("t", category)
	}
	if query != "" {
		params.Set("ac", "detail")
		params.Set("wd", query)
	}
	response, err := d.lujj31Request(ctx, params)
	if err != nil {
		return nil, false, err
	}
	records := lujj31Rows(response["list"])
	items := make([]Drama, 0, len(records))
	seen := map[string]bool{}
	for _, row := range records {
		drama := lujj31Drama(row, d.providerBaseURL(sourceLujj31))
		if drama.ID != "" && !seen[drama.ID] {
			seen[drama.ID] = true
			items = append(items, drama)
		}
	}
	if _, exists := response["list"]; !exists {
		return nil, false, errors.New("lujj31 列表结构不符合 MacCMS 格式")
	}
	pageCount, err := strconv.Atoi(mapString(response, "pagecount"))
	more := err == nil && pageCount > page
	return items, more, nil
}

func lujj31Chapters(sourceID, entries, referer string) []Chapter {
	for _, line := range strings.Split(entries, "$$$") {
		var chapters []Chapter
		for _, entry := range strings.Split(line, "#") {
			title, mediaURL, found := strings.Cut(strings.TrimSpace(entry), "$")
			mediaURL = strings.TrimSpace(mediaURL)
			if !found || !isProviderHTTPMediaURL(mediaURL) {
				continue
			}
			number := len(chapters) + 1
			if number > 10000 {
				break
			}
			if strings.TrimSpace(title) == "" {
				title = fmt.Sprintf("第%d集", number)
			}
			chapters = append(chapters, Chapter{
				ID:     providerChapterID(sourceLujj31, sourceID, strconv.Itoa(number)),
				Source: sourceLujj31, Title: truncate(strings.TrimSpace(title), 128),
				CurrentEpisode: rawEpisode(number), VideoURL: mediaURL, Referer: referer,
			})
		}
		if len(chapters) > 0 {
			return chapters
		}
	}
	return nil
}

func (d *Downloader) fetchLujj31Detail(ctx context.Context, sourceID string) (Drama, []Chapter, error) {
	if !webProviderNumericID.MatchString(sourceID) {
		return Drama{}, nil, errors.New("lujj31 剧集 ID 无效")
	}
	response, err := d.lujj31Request(ctx, url.Values{"ac": {"detail"}, "ids": {sourceID}})
	if err != nil {
		return Drama{}, nil, err
	}
	for _, row := range lujj31Rows(response["list"]) {
		if mapString(row, "vod_id") != sourceID {
			continue
		}
		site := d.providerBaseURL(sourceLujj31)
		drama := lujj31Drama(row, site)
		if drama.ID == "" {
			break
		}
		chapters := lujj31Chapters(sourceID, mapString(row, "vod_play_url"), site+"/")
		if len(chapters) == 0 {
			return Drama{}, nil, errors.New("lujj31 详情没有公开的 MP4/M3U8 播放地址")
		}
		drama.TotalEpisode, drama.EpisodeCount = len(chapters), len(chapters)
		return drama, chapters, nil
	}
	return Drama{}, nil, errors.New("lujj31 未找到对应的剧集详情")
}

func (d *Downloader) resolveLujj31Media(ctx context.Context, task Task) (providerMedia, error) {
	source, id, ok := splitProviderDramaID(task.DramaID)
	prefix := providerChapterID(sourceLujj31, id, "")
	if !ok || source != sourceLujj31 || !webProviderNumericID.MatchString(id) ||
		!strings.HasPrefix(task.Chapter.ID, prefix) || !isProviderHTTPMediaURL(task.Chapter.VideoURL) {
		return providerMedia{}, errors.New("lujj31 分集播放资料无效")
	}
	media := providerMedia{URL: task.Chapter.VideoURL, Referer: d.providerBaseURL(sourceLujj31) + "/"}
	return d.prepareWebProviderMedia(ctx, media, "lujj31")
}
