package core

import "context"

// Registered custom sources use the official upstream Provider Registry.
// Their existing catalog/detail/media parsers remain the source of truth.
type customSiteProvider struct {
 d *Downloader
 descriptor SourceDescriptor
 categories func(context.Context) ([]Category, error)
 catalog func(context.Context, int, string, string) ([]Drama, bool, error)
 detail func(context.Context, string) (Drama, []Chapter, error)
 media func(context.Context, Task) (ProviderMedia, error)
}

func (p *customSiteProvider) Descriptor() SourceDescriptor { return p.descriptor }
func (p *customSiteProvider) Categories(ctx context.Context, force bool) ([]Category, error) {
 return p.categories(ctx)
}
func (p *customSiteProvider) CatalogPage(ctx context.Context, page int, category string) ([]Drama, bool, error) {
 return p.catalog(ctx, page, category, "")
}
func (p *customSiteProvider) Search(ctx context.Context, page int, query string) ([]Drama, bool, error) {
 return p.catalog(ctx, page, "", query)
}
func (p *customSiteProvider) Detail(ctx context.Context, id string) (Drama, []Chapter, error) {
 raw, chapters, err := p.detail(ctx, id)
 if err != nil { return raw, chapters, err }
 // Preserve the catalog title if a Tao detail page exposes a generic
 // playback-line label instead of the actual series title.
 if p.descriptor.ID == sourceLujj31 &&
   (!nativeLujj31CatalogTitleValid(raw.DisplayTitle(), id) ||
     raw.DisplayTitle() == "播放線路" || raw.DisplayTitle() == "播放线路") {
  raw.Title = ""
  raw.Name = ""
 }
 return raw, chapters, nil
}
func (p *customSiteProvider) ResolveMedia(ctx context.Context, task Task) (ProviderMedia, error) {
 return p.media(ctx, task)
}

func registerCustomSource(id, name string, host Host,
 categories func(context.Context) ([]Category, error),
 catalog func(context.Context, int, string, string) ([]Drama, bool, error),
 detail func(context.Context, string) (Drama, []Chapter, error),
 media func(context.Context, Task) (ProviderMedia, error)) Provider {
 d, ok := host.(*Downloader)
 if !ok { return nil }
 return &customSiteProvider{
  d: d, descriptor: SourceDescriptor{ID: id, Name: name, DefaultOff: true,
   BaseURL: map[string]string{sourceLujj31: lujj31BaseURL, sourceLiuguo: liuguoBaseURL, sourceXiaoguo: xiaoguoBaseURL}[id],
   NoHostRotation: true,
   ValidCategory: func(category string) bool { return category == "" || WebProviderNumericID.MatchString(category) }},
  categories: categories, catalog: catalog, detail: detail, media: media,
 }
}

func init() {
 Register(SourceDescriptor{ID: sourceLujj31, Name: "桃果", DefaultOff: true, BaseURL: lujj31BaseURL, NoHostRotation: true,
  ValidCategory: func(category string) bool { return category == "" || WebProviderNumericID.MatchString(category) }},
  func(host Host) Provider {
   d, ok := host.(*Downloader); if !ok { return nil }
   return registerCustomSource(sourceLujj31, "桃果", host, d.fetchLujj31Categories,
    d.fetchLujj31CatalogPage, d.fetchLujj31Detail, d.resolveLujj31Media)
  })
 Register(SourceDescriptor{ID: sourceLiuguo, Name: "榴果", DefaultOff: true, BaseURL: liuguoBaseURL, NoHostRotation: true,
  ValidCategory: func(category string) bool { return category == "" || WebProviderNumericID.MatchString(category) }},
  func(host Host) Provider {
   d, ok := host.(*Downloader); if !ok { return nil }
   return registerCustomSource(sourceLiuguo, "榴果", host, d.fetchLiuguoCategories,
    d.fetchLiuguoCatalogPage, d.fetchLiuguoDetail, d.resolveLiuguoMedia)
  })
 Register(SourceDescriptor{ID: sourceXiaoguo, Name: "宵果", DefaultOff: true, BaseURL: xiaoguoBaseURL, NoHostRotation: true,
  ValidCategory: validXiaoguoCategory},
  func(host Host) Provider {
   d, ok := host.(*Downloader); if !ok { return nil }
   provider := registerCustomSource(sourceXiaoguo, "宵果", host, d.fetchXiaoguoCategories,
    d.fetchXiaoguoCatalogPage, d.fetchXiaoguoDetail, d.resolveXiaoguoMedia)
   if p, ok := provider.(*customSiteProvider); ok {
    p.descriptor.ValidCategory = validXiaoguoCategory
   }
   return provider
  })
 // The carried-forward APK391 adapters are registered by their own IDs so
 // upstream's runtime source list can discover them dynamically.
 for _, entry := range apk391Sources {
  spec := entry
  descriptor := SourceDescriptor{
   ID: spec.ID, Name: spec.Name,
   DefaultOff: true, ValidCategory: func(category string) bool {
    return apk391ValidCategory(spec.ID, category)
   },
  }
  Register(descriptor, func(host Host) Provider {
   d, ok := host.(*Downloader)
   if !ok { return nil }
   return &customSiteProvider{
    d: d, descriptor: descriptor,
    categories: func(ctx context.Context) ([]Category, error) {
     return d.fetchApk391Categories(ctx, spec.ID)
    },
    catalog: func(ctx context.Context, page int, category, query string) ([]Drama, bool, error) {
     return d.fetchApk391CatalogPage(ctx, spec.ID, page, category, query)
    },
    detail: func(ctx context.Context, id string) (Drama, []Chapter, error) {
     return d.fetchApk391Detail(ctx, spec.ID, id)
    },
    media: d.resolveApk391Media,
   }
  })
 }
}
