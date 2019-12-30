package fuzzyfinder

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/termbox"
	runewidth "github.com/mattn/go-runewidth"
)

type cell struct {
	ch     rune
	bg, fg termbox.Attribute
}

// TerminalMock is a mocked terminal for testing.
// Most users should use it by calling UseMockedTerminal.
type TerminalMock struct {
	sizeMu        sync.Mutex
	width, height int

	eventsMu sync.Mutex
	events   []termbox.Event

	cellsMu sync.Mutex
	cells   []*cell

	resultMu sync.Mutex
	result   string

	sleepDuration time.Duration
}

// SetSize changes the pseudo-size of the window.
// Note that SetSize resets added cells.
func (m *TerminalMock) SetSize(w, h int) {
	m.sizeMu.Lock()
	defer m.sizeMu.Unlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()
	m.width = w
	m.height = h
	m.cells = make([]*cell, w*h)
}

// SetEvents sets all events, which are fetched by pollEvent.
// A user of this must set the EscKey event at the end.
func (m *TerminalMock) SetEvents(e ...termbox.Event) {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	m.events = e
}

// GetResult returns a flushed string that is displayed to the actual terminal.
// It contains all escape sequences such that ANSI escape code.
func (m *TerminalMock) GetResult() string {
	m.resultMu.Lock()
	defer m.resultMu.Unlock()
	return m.result
}

func (m *TerminalMock) init() error {
	return nil
}

func (m *TerminalMock) size() (width int, height int) {
	m.sizeMu.Lock()
	defer m.sizeMu.Unlock()
	return m.width, m.height
}

func (m *TerminalMock) clear(fg termbox.Attribute, bg termbox.Attribute) error {
	// TODO
	return nil
}

func (m *TerminalMock) setCell(x int, y int, ch rune, fg termbox.Attribute, bg termbox.Attribute) {
	m.sizeMu.Lock()
	defer m.sizeMu.Unlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()

	if x < 0 || x >= m.width {
		return
	}
	if y < 0 || y >= m.height {
		return
	}
	m.cells[y*m.width+x] = &cell{ch: ch, fg: fg, bg: bg}
}

func (m *TerminalMock) setCursor(x int, y int) {
	m.sizeMu.Lock()
	defer m.sizeMu.Unlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()
	if x < 0 || x >= m.width {
		return
	}
	if y < 0 || y >= m.height {
		return
	}
	i := y*m.width + x
	if m.cells[i] == nil {
		m.cells[y*m.width+x] = &cell{ch: '\u2588', fg: termbox.ColorWhite, bg: termbox.ColorDefault}
	} else {
		// Cursor on a rune.
		m.cells[y*m.width+x].bg = termbox.ColorWhite
	}
	return
}

func (m *TerminalMock) pollEvent() termbox.Event {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	if len(m.events) == 0 {
		panic("pollEvent called with empty events. have you set expected events by SetEvents?")
	}
	e := m.events[0]
	m.events = m.events[1:]
	// Wait a moment for goroutine scheduling.
	time.Sleep(m.sleepDuration)
	return e
}

// flush displays all items with formatted layout.
func (m *TerminalMock) flush() error {
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()

	var s string
	for j := 0; j < m.height; j++ {
		prevFg, prevBg := termbox.ColorDefault, termbox.ColorDefault
		for i := 0; i < m.width; i++ {
			c := m.cells[j*m.width+i]
			if c == nil {
				s += " "
				prevFg, prevBg = termbox.ColorDefault, termbox.ColorDefault
				continue
			} else {
				var fgReset bool
				if c.fg != prevFg {
					s += "\x1b\x5b\x6d" // Reset previous color.
					s += parseAttr(c.fg, true)
					prevFg = c.fg
					prevBg = termbox.ColorDefault
					fgReset = true
				}
				if c.bg != prevBg {
					if !fgReset {
						s += "\x1b\x5b\x6d" // Reset previous color.
						prevFg = termbox.ColorDefault
					}
					s += parseAttr(c.bg, false)
					prevBg = c.bg
				}
				s += string(c.ch)
				rw := runewidth.RuneWidth(c.ch)
				if rw != 1 {
					i += rw - 1
				}
			}
		}
		s += "\n"
	}
	s += "\x1b\x5b\x6d" // Reset previous color.
	m.cells = make([]*cell, m.width*m.height)

	m.resultMu.Lock()
	defer m.resultMu.Unlock()
	m.result = s

	return nil
}

func (m *TerminalMock) close() {}

// UseMockedTerminal switches the terminal, which is used from
// this package to a mocked one.
func UseMockedTerminal() *TerminalMock {
	return defaultFinder.UseMockedTerminal()
}

func (f *finder) UseMockedTerminal() *TerminalMock {
	m := &TerminalMock{}
	f.term = m
	return m
}

// parseAttr parses an attribute of termbox
// as an escape sequence.
// parseAttr doesn't support output modes othar than color256 in termbox-go.
func parseAttr(attr termbox.Attribute, isFg bool) string {
	var buf bytes.Buffer
	buf.WriteString("\x1b[")
	if attr >= termbox.AttrReverse {
		buf.WriteString("7;")
		attr -= termbox.AttrReverse
	}
	if attr >= termbox.AttrUnderline {
		buf.WriteString("4;")
		attr -= termbox.AttrUnderline
	}
	if attr >= termbox.AttrBold {
		buf.WriteString("1;")
		attr -= termbox.AttrBold
	}

	if attr > termbox.ColorWhite {
		panic(fmt.Sprintf("invalid color code: %d", attr))
	}

	if attr == termbox.ColorDefault {
		if isFg {
			buf.WriteString("39")
		} else {
			buf.WriteString("49")
		}
	} else {
		color := int(attr) - 1
		if isFg {
			fmt.Fprintf(&buf, "38;5;%d", color)
		} else {
			fmt.Fprintf(&buf, "48;5;%d", color)
		}
	}
	buf.WriteString("m")

	return buf.String()
}
