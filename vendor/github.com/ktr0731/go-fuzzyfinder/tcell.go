package fuzzyfinder

import (
	"github.com/gdamore/tcell/termbox"
)

// terminal is an abstraction for mocking termbox-go.
type terminal interface {
	init() error
	size() (width int, height int)
	clear(termbox.Attribute, termbox.Attribute) error
	setCell(x, y int, ch rune, fg, bg termbox.Attribute)
	setCursor(x, y int)
	pollEvent() termbox.Event
	flush() error
	close()
}

// termImpl is the implementation for termbox-go.
type termImpl struct{}

func (t *termImpl) init() error {
	return termbox.Init()
}

func (t *termImpl) size() (width int, height int) {
	return termbox.Size()
}

func (t *termImpl) clear(fg termbox.Attribute, bg termbox.Attribute) error {
	termbox.Clear(fg, bg)
	return nil
}

func (t *termImpl) setCell(x int, y int, ch rune, fg termbox.Attribute, bg termbox.Attribute) {
	termbox.SetCell(x, y, ch, fg, bg)
}

func (t *termImpl) setCursor(x int, y int) {
	termbox.SetCursor(x, y)
}

func (t *termImpl) pollEvent() termbox.Event {
	return termbox.PollEvent()
}

func (t *termImpl) flush() error {
	return termbox.Flush()
}

func (t *termImpl) close() {
	termbox.Close()
}
