package fuzzyfinder

type opt struct {
	mode        mode
	previewFunc func(i, width, height int) string
	multi       bool
	hotReload   bool
}

type mode int

const (
	// ModeSmart enables a smart matching. It is the default matching mode.
	// At the beginning, matching mode is ModeCaseInsensitive, but it switches
	// over to ModeCaseSensitive if an upper case character is inputted.
	ModeSmart mode = iota
	// ModeCaseSensitive enables a case-sensitive matching.
	ModeCaseSensitive
	// ModeCaseInsensitive enables a case-insensitive matching.
	ModeCaseInsensitive
)

// Option represents available fuzzy-finding options.
type Option func(*opt)

// WithMode specifies a matching mode. The default mode is ModeSmart.
func WithMode(m mode) Option {
	return func(o *opt) {
		o.mode = m
	}
}

// WithPreviewWindow enables to display a preview for the selected item.
// The argument f receives i, width and height. i is the same as Find's one.
// width and height are the size of the terminal so that you can use these to adjust
// a preview content. Note that width and height are calculated as a rune-based length.
//
// If there is no selected item, previewFunc passes -1 to previewFunc.
//
// If f is nil, the preview feature is disabled.
func WithPreviewWindow(f func(i, width, height int) string) Option {
	return func(o *opt) {
		o.previewFunc = f
	}
}

// WithHotReload reloads the passed slice automatically when some entries are appended.
// The caller must pass a pointer of the slice instead of the slice itself.
func WithHotReload() Option {
	return func(o *opt) {
		o.hotReload = true
	}
}

// withMulti enables to select multiple items by tab key.
func withMulti() Option {
	return func(o *opt) {
		o.multi = true
	}
}
