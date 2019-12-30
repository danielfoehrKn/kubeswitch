// Package fuzzyfinder provides terminal user interfaces for fuzzy-finding.
//
// Note that, all functions are not goroutine-safe.
package fuzzyfinder

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/termbox"
	"github.com/ktr0731/go-fuzzyfinder/matching"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
)

var (
	// ErrAbort is returned from Find* functions if there are no selections.
	ErrAbort   = errors.New("abort")
	errEntered = errors.New("entered")
)

var (
	defaultFinder = &finder{}
)

type state struct {
	items      []string           // All item names.
	allMatched []matching.Matched // All items.
	matched    []matching.Matched // Matched items against to the input.

	// x is the current index of the input line.
	x int
	// cursorX is the position of input line.
	// Note that cursorX is the actual width of input runes.
	cursorX int

	// The current index of filtered items (matched).
	// The initial state is 0.
	y int
	// cursorY is the position of item line.
	// Note that the max size of cursorY depends on max height.
	cursorY int

	input []rune

	// selections holds whether a key is selected or not. Each key is
	// an index of an item (Matched.Idx). Each value represents the position
	// which it is selected.
	selection map[int]int
	// selectionIdx hods the next index, which is used to a selection's value.
	selectionIdx int
}

type finder struct {
	term      terminal
	stateMu   sync.RWMutex
	state     state
	drawTimer *time.Timer
	eventCh   chan struct{}
	opt       *opt
}

func (f *finder) initFinder(items []string, matched []matching.Matched, opt opt) error {
	if f.term == nil {
		f.term = &termImpl{}
	}

	if err := f.term.init(); err != nil {
		return errors.Wrap(err, "failed to initialize termbox")
	}

	f.opt = &opt
	f.state = state{}

	if opt.multi {
		f.state.selection = map[int]int{}
	}

	f.state.items = items
	f.state.matched = matched
	f.state.allMatched = matched
	if !isInTesting() {
		f.drawTimer = time.AfterFunc(0, func() {
			f._draw()
			f._drawPreview()
			f.term.flush()
		})
		f.drawTimer.Stop()
	}
	f.eventCh = make(chan struct{}, 30) // A large value
	return nil
}

func (f *finder) updateItems(items []string, matched []matching.Matched) {
	f.stateMu.Lock()
	f.state.items = items
	f.state.matched = matched
	f.state.allMatched = matched
	f.stateMu.Unlock()
	f.eventCh <- struct{}{}
}

// _draw is used from draw with a timer.
func (f *finder) _draw() {
	width, height := f.term.size()
	f.term.clear(termbox.ColorDefault, termbox.ColorDefault)

	maxWidth := width
	if f.opt.previewFunc != nil {
		maxWidth = width/2 - 1
	}

	// input line
	f.term.setCell(0, height-1, '>', termbox.ColorBlue, termbox.ColorDefault)
	var r rune
	var w int
	for _, r = range f.state.input {
		// Add a space between '>' and runes.
		f.term.setCell(2+w, height-1, r, termbox.ColorDefault|termbox.AttrBold, termbox.ColorDefault)
		w += runewidth.RuneWidth(r)
	}
	f.term.setCursor(2+f.state.cursorX, height-1)

	// Number line
	for i, r := range fmt.Sprintf("%d/%d", len(f.state.matched), len(f.state.items)) {
		f.term.setCell(2+i, height-2, r, termbox.ColorYellow, termbox.ColorDefault)
	}

	// Item lines
	itemAreaHeight := height - 2 - 1
	matched := f.state.matched
	offset := f.state.cursorY
	y := f.state.y
	// From the first (the most bottom) item in the item lines to the end.
	matched = matched[y-offset:]

	for i, m := range matched {
		if i > itemAreaHeight {
			break
		}
		if i == f.state.cursorY {
			f.term.setCell(0, height-3-i, '>', termbox.ColorRed, termbox.ColorBlack)
			f.term.setCell(1, height-3-i, ' ', termbox.ColorRed, termbox.ColorBlack)
		}

		if f.opt.multi {
			if _, ok := f.state.selection[m.Idx]; ok {
				f.term.setCell(1, height-3-i, '>', termbox.ColorRed, termbox.ColorBlack)
			}
		}

		var posIdx int
		w := 2
		for j, r := range []rune(f.state.items[m.Idx]) {
			fg := termbox.ColorDefault
			bg := termbox.ColorDefault
			// Highlight selected strings.
			if posIdx < len(f.state.input) {
				from, to := m.Pos[0], m.Pos[1]
				if !(from == -1 && to == -1) && (from <= j && j <= to) {
					if unicode.ToLower(f.state.input[posIdx]) == unicode.ToLower(r) {
						fg |= termbox.ColorGreen
						posIdx++
					}
				}
			}
			if i == f.state.cursorY {
				fg |= termbox.AttrBold | termbox.ColorYellow
				bg = termbox.ColorBlack
			}

			rw := runewidth.RuneWidth(r)
			// Shorten item cells.
			if w+rw+2 > maxWidth {
				f.term.setCell(w, height-3-i, '.', fg, bg)
				f.term.setCell(w+1, height-3-i, '.', fg, bg)
				w += 2
				break
			} else {
				f.term.setCell(w, height-3-i, r, fg, bg)
				w += rw
			}
		}
	}
}

func (f *finder) _drawPreview() {
	if f.opt.previewFunc == nil {
		return
	}

	width, height := f.term.size()
	var idx int
	if len(f.state.matched) == 0 {
		idx = -1
	} else {
		idx = f.state.matched[f.state.y].Idx
	}

	sp := strings.Split(f.opt.previewFunc(idx, width, height), "\n")
	prevLines := make([][]rune, 0, len(sp))
	for _, s := range sp {
		prevLines = append(prevLines, []rune(s))
	}

	// top line
	for i := width / 2; i < width; i++ {
		var r rune
		if i == width/2 {
			r = '┌'
		} else if i == width-1 {
			r = '┐'
		} else {
			r = '─'
		}
		f.term.setCell(i, 0, r, termbox.ColorBlack, termbox.ColorDefault)
	}
	// bottom line
	for i := width / 2; i < width; i++ {
		var r rune
		if i == width/2 {
			r = '└'
		} else if i == width-1 {
			r = '┘'
		} else {
			r = '─'
		}
		f.term.setCell(i, height-1, r, termbox.ColorBlack, termbox.ColorDefault)
	}
	// Start with h=1 to exclude each corner rune.
	const vline = '│'
	var wvline = runewidth.RuneWidth(vline)
	for h := 1; h < height-1; h++ {
		w := width / 2
		for i := width / 2; i < width; i++ {
			switch {
			// Left vertical line.
			case i == width/2:
				f.term.setCell(i, h, vline, termbox.ColorBlack, termbox.ColorDefault)
				w += wvline
			// Right vertical line.
			case i == width-1:
				f.term.setCell(i, h, vline, termbox.ColorBlack, termbox.ColorDefault)
				w += wvline
			// Spaces between left and right vertical lines.
			case w == width/2+wvline, w == width-1-wvline:
				f.term.setCell(w, h, ' ', termbox.ColorDefault, termbox.ColorDefault)
				w++
			default: // Preview text
				if h-1 >= len(prevLines) {
					w++
					continue
				}
				j := i - width/2 - 2 // Two spaces.
				l := prevLines[h-1]
				if j >= len(l) {
					w++
					continue
				}
				rw := runewidth.RuneWidth(l[j])
				if w+rw > width-1-2 {
					f.term.setCell(w, h, '.', termbox.ColorDefault, termbox.ColorDefault)
					f.term.setCell(w+1, h, '.', termbox.ColorDefault, termbox.ColorDefault)
					w += 2
					continue
				}

				f.term.setCell(w, h, l[j], termbox.ColorDefault, termbox.ColorDefault)
				w += rw
			}
		}
	}
}

func (f *finder) draw(d time.Duration) {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if isInTesting() {
		// Don't use goroutine scheduling.
		f._draw()
		f._drawPreview()
		f.term.flush()
	} else {
		f.drawTimer.Reset(d)
	}
}

// readKey reads a key input.
// It returns ErrAbort if esc, CTRL-C or CTRL-D keys are inputted.
// Also, it returns errEntered if enter key is inputted.
func (f *finder) readKey() error {
	f.stateMu.RLock()
	prevInputLen := len(f.state.input)
	f.stateMu.RUnlock()
	defer func() {
		f.stateMu.RLock()
		currentInputLen := len(f.state.input)
		f.stateMu.RUnlock()
		if prevInputLen != currentInputLen {
			f.eventCh <- struct{}{}
		}
	}()

	e := f.term.pollEvent()
	f.stateMu.Lock()
	defer f.stateMu.Unlock()

	switch e.Type {
	case termbox.EventKey:
		switch e.Key {
		case termbox.KeyEsc, termbox.KeyCtrlC, termbox.KeyCtrlD:
			return ErrAbort
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(f.state.input) == 0 {
				return nil
			}
			if f.state.x == 0 {
				return nil
			}
			// Remove the latest input rune.
			f.state.cursorX -= runewidth.RuneWidth(f.state.input[len(f.state.input)-1])
			f.state.x--
			f.state.input = f.state.input[0 : len(f.state.input)-1]
		case termbox.KeyDelete:
			if f.state.x == len(f.state.input) {
				return nil
			}
			x := f.state.x

			f.state.input = append(f.state.input[:x], f.state.input[x+1:]...)
		case termbox.KeyEnter:
			return errEntered
		case termbox.KeyArrowLeft, termbox.KeyCtrlB:
			if f.state.x > 0 {
				f.state.cursorX -= runewidth.RuneWidth(f.state.input[f.state.x-1])
				f.state.x--
			}
		case termbox.KeyArrowRight, termbox.KeyCtrlF:
			if f.state.x < len(f.state.input) {
				f.state.cursorX += runewidth.RuneWidth(f.state.input[f.state.x])
				f.state.x++
			}
		case termbox.KeyCtrlA:
			f.state.cursorX = 0
			f.state.x = 0
		case termbox.KeyCtrlE:
			f.state.cursorX = runewidth.StringWidth(string(f.state.input))
			f.state.x = len(f.state.input)
		case termbox.KeyCtrlW:
			in := f.state.input[:f.state.x]
			inStr := string(in)
			pos := strings.LastIndex(strings.TrimRightFunc(inStr, unicode.IsSpace), " ")
			if pos == -1 {
				f.state.input = []rune{}
				f.state.cursorX = 0
				f.state.x = 0
				return nil
			}
			pos = utf8.RuneCountInString(inStr[:pos])
			newIn := f.state.input[:pos+1]
			f.state.input = newIn
			f.state.cursorX = runewidth.StringWidth(string(newIn))
			f.state.x = len(newIn)
		case termbox.KeyCtrlU:
			f.state.input = f.state.input[f.state.x:]
			f.state.cursorX = 0
			f.state.x = 0
		case termbox.KeyArrowUp, termbox.KeyCtrlK, termbox.KeyCtrlP:
			if f.state.y+1 < len(f.state.matched) {
				f.state.y++
			}
			_, height := f.term.size()
			if f.state.cursorY+1 < height-2 && f.state.cursorY+1 < len(f.state.matched) {
				f.state.cursorY++
			}
		case termbox.KeyArrowDown, termbox.KeyCtrlJ, termbox.KeyCtrlN:
			if f.state.y > 0 {
				f.state.y--
			}
			if f.state.cursorY-1 >= 0 {
				f.state.cursorY--
			}
		case termbox.KeyTab:
			if !f.opt.multi {
				return nil
			}
			idx := f.state.matched[f.state.y].Idx
			if _, ok := f.state.selection[idx]; ok {
				delete(f.state.selection, idx)
			} else {
				f.state.selection[idx] = f.state.selectionIdx
				f.state.selectionIdx++
			}
			if f.state.y > 0 {
				f.state.y--
			}
			if f.state.cursorY > 0 {
				f.state.cursorY--
			}
		default:
			if e.Key == termbox.KeySpace {
				e.Ch = ' '
			}
			if e.Ch != 0 {
				width, _ := f.term.size()
				maxLineWidth := width - 2 - 1
				if len(f.state.input)+1 > maxLineWidth {
					// Discard inputted rune.
					return nil
				}

				x := f.state.x
				f.state.input = append(f.state.input[:x], append([]rune{e.Ch}, f.state.input[x:]...)...)
				f.state.cursorX += runewidth.RuneWidth(e.Ch)
				f.state.x++
			}
		}
	case termbox.EventResize:
		// To get actual window size, clear all buffers.
		// See termbox.Clear's documentation for more details.
		f.term.clear(termbox.ColorDefault, termbox.ColorDefault)

		width, height := f.term.size()
		itemAreaHeight := height - 2 - 1
		if itemAreaHeight >= 0 && f.state.cursorY > itemAreaHeight {
			f.state.cursorY = itemAreaHeight
		}

		maxLineWidth := width - 2 - 1
		if maxLineWidth < 0 {
			f.state.input = nil
			f.state.cursorX = 0
			f.state.x = 0
		} else if len(f.state.input)+1 > maxLineWidth {
			// Discard inputted rune.
			f.state.input = f.state.input[:maxLineWidth]
			f.state.cursorX = runewidth.StringWidth(string(f.state.input))
			f.state.x = maxLineWidth
		}
	}
	return nil
}

func (f *finder) filter() {
	f.stateMu.RLock()
	if len(f.state.input) == 0 {
		f.stateMu.RUnlock()
		f.stateMu.Lock()
		defer f.stateMu.Unlock()
		f.state.matched = f.state.allMatched
		return
	}

	// TODO: If input is not delete operation, it is able to
	// reduce total iteration.
	// FindAll may take a lot of time, so it is desired to use RLock to avoid goroutine blocking.
	matchedItems := matching.FindAll(string(f.state.input), f.state.items, matching.WithMode(matching.Mode(f.opt.mode)))
	f.stateMu.RUnlock()

	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.state.matched = matchedItems
	if len(f.state.matched) == 0 {
		f.state.cursorY = 0
		f.state.y = 0
		return
	}

	switch {
	case f.state.cursorY >= len(f.state.matched):
		f.state.cursorY = len(f.state.matched) - 1
		f.state.y = len(f.state.matched) - 1
	case f.state.y >= len(f.state.matched):
		f.state.y = len(f.state.matched) - 1
	}
}

func (f *finder) find(slice interface{}, itemFunc func(i int) string, opts []Option) ([]int, error) {
	if itemFunc == nil {
		return nil, errors.New("itemFunc must not be nil")
	}

	var opt opt
	for _, o := range opts {
		o(&opt)
	}

	rv := reflect.ValueOf(slice)
	if opt.hotReload && (rv.Kind() != reflect.Ptr || reflect.Indirect(rv).Kind() != reflect.Slice) {
		return nil, errors.Errorf("the first argument must be a pointer to a slice, but got %T", slice)
	} else if !opt.hotReload && rv.Kind() != reflect.Slice {
		return nil, errors.Errorf("the first argument must be a slice, but got %T", slice)
	}

	makeItems := func(sliceLen int) ([]string, []matching.Matched) {
		items := make([]string, sliceLen)
		matched := make([]matching.Matched, sliceLen)
		for i := 0; i < sliceLen; i++ {
			items[i] = itemFunc(i)
			matched[i] = matching.Matched{Idx: i}
		}
		return items, matched
	}

	var (
		items   []string
		matched []matching.Matched
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inited := make(chan struct{})
	if opt.hotReload && rv.Kind() == reflect.Ptr {
		rvv := reflect.Indirect(rv)
		items, matched = makeItems(rvv.Len())

		go func() {
			<-inited

			var prev int
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Millisecond):
					curr := rvv.Len()
					if prev != curr {
						items, matched = makeItems(curr)
						f.updateItems(items, matched)
					}
					prev = curr
				}
			}
		}()
	} else {
		items, matched = makeItems(rv.Len())
	}

	if err := f.initFinder(items, matched, opt); err != nil {
		return nil, errors.Wrap(err, "failed to initialize the fuzzy finder")
	}
	defer f.term.close()

	close(inited)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-f.eventCh:
				f.filter()
				f.draw(0)
			}
		}
	}()

	for {
		f.draw(10 * time.Millisecond)

		err := f.readKey()
		switch {
		case err == ErrAbort:
			return nil, ErrAbort
		case err == errEntered:
			f.stateMu.RLock()
			defer f.stateMu.RUnlock()

			if len(f.state.matched) == 0 {
				return nil, ErrAbort
			}
			if f.opt.multi {
				if len(f.state.selection) == 0 {
					return []int{f.state.matched[f.state.y].Idx}, nil
				}
				poss, idxs := make([]int, 0, len(f.state.selection)), make([]int, 0, len(f.state.selection))
				for idx, pos := range f.state.selection {
					idxs = append(idxs, idx)
					poss = append(poss, pos)
				}
				sort.Slice(idxs, func(i, j int) bool {
					return poss[i] < poss[j]
				})
				return idxs, nil
			}
			return []int{f.state.matched[f.state.y].Idx}, nil
		case err != nil:
			return nil, errors.Wrap(err, "failed to read a key")
		}
	}
}

// Find displays a UI that provides fuzzy finding against to the passed slice.
// The argument slice must be a slice type. If it is not a slice, Find returns
// an error. itemFunc is called by the length of slice. previewFunc is called
// when the cursor which points the current selected item is changed.
// If itemFunc is nil, Find returns an error.
//
// itemFunc receives an argument i. It is the index of the item currently
// selected.
//
// Find returns ErrAbort if a call of Find is finished with no selection.
func Find(slice interface{}, itemFunc func(i int) string, opts ...Option) (int, error) {
	return defaultFinder.Find(slice, itemFunc, opts...)
}

func (f *finder) Find(slice interface{}, itemFunc func(i int) string, opts ...Option) (int, error) {
	res, err := f.find(slice, itemFunc, opts)
	if err != nil {
		return 0, err
	}
	return res[0], err
}

// FindMulti is nearly same as the Find. The only one difference point from
// Find is the user can select multiple items at once by tab key.
func FindMulti(slice interface{}, itemFunc func(i int) string, opts ...Option) ([]int, error) {
	return defaultFinder.FindMulti(slice, itemFunc, opts...)
}

func (f *finder) FindMulti(slice interface{}, itemFunc func(i int) string, opts ...Option) ([]int, error) {
	opts = append(opts, withMulti())
	return f.find(slice, itemFunc, opts)
}

func isInTesting() bool {
	return flag.Lookup("test.v") != nil
}
