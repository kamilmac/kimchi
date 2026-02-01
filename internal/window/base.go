package window

import "github.com/kmacinski/blocks/internal/config"

// Base provides common functionality for windows
type Base struct {
	name    string
	focused bool
	styles  config.Styles
}

// NewBase creates a new base window
func NewBase(name string, styles config.Styles) Base {
	return Base{
		name:   name,
		styles: styles,
	}
}

// Name returns the window name
func (b *Base) Name() string {
	return b.name
}

// Focused returns whether the window is focused
func (b *Base) Focused() bool {
	return b.focused
}

// SetFocus sets the focus state
func (b *Base) SetFocus(focused bool) {
	b.focused = focused
}

// Styles returns the window styles
func (b *Base) Styles() config.Styles {
	return b.styles
}
