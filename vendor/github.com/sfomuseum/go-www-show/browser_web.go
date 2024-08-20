package show

import (
	"context"

	"github.com/pkg/browser"
)

// WebBrowser implements the `Browser` interface for loading URLs in a web browser.
type WebBrowser struct {
	Browser
}

func init() {

	ctx := context.Background()

	err := RegisterBrowser(ctx, "web", NewWebBrowser)

	if err != nil {
		panic(err)
	}
}

// NewWebBrowser retuns a new `WebBrowser` instance.
func NewWebBrowser(ctx context.Context, uri string) (Browser, error) {
	br := &WebBrowser{}
	return br, nil
}

// OpenURL opens 'url' in the operating system's default web browser.
func (br *WebBrowser) OpenURL(ctx context.Context, url string) error {
	return browser.OpenURL(url)
}
