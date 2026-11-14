package core

import (
 "context"
 "golang.org/x/net/html"
)

func (d *Downloader) fetchProviderText(ctx context.Context, address, referer string) (string, error) {
 response, err := d.FetchText(ctx, TextRequest{URL: address, Referer: referer})
 if err != nil { return "", err }
 return response.Body, nil
}

func (d *Downloader) fetchProviderPage(ctx context.Context, address, referer, agent string) (*html.Node, string, error) {
 return FetchHostPage(ctx, d, address, referer, agent)
}
