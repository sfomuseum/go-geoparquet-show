package show

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// Browser is an interface for rendering URLs.
type Browser interface {
	// OpenURL opens a given in a URL specific to the browser's implementation context.
	OpenURL(context.Context, string) error
}

var browser_roster roster.Roster

// BrowserInitializationFunc is a function defined by individual browser package and used to create
// an instance of that browser
type BrowserInitializationFunc func(ctx context.Context, uri string) (Browser, error)

// RegisterBrowser registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Browser` instances by the `NewBrowser` method.
func RegisterBrowser(ctx context.Context, scheme string, init_func BrowserInitializationFunc) error {

	err := ensureBrowserRoster()

	if err != nil {
		return err
	}

	return browser_roster.Register(ctx, scheme, init_func)
}

func ensureBrowserRoster() error {

	if browser_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		browser_roster = r
	}

	return nil
}

// NewBrowser returns a new `Browser` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `BrowserInitializationFunc`
// function used to instantiate the new `Browser`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterBrowser` method.
func NewBrowser(ctx context.Context, uri string) (Browser, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := browser_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(BrowserInitializationFunc)
	return init_func(ctx, uri)
}

// BrowserSchemes returns the list of schemes that have been registered.
func BrowserSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureBrowserRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range browser_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
