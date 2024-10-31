package components

import (
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/output"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

type Item[T any] struct {
	Label string
	Value T
}

type fmtFunc[T any] func(index int, item Item[T]) string

type List[T any] struct {
	renderer        *sdl.Renderer
	itemFormatter   fmtFunc[T]
	items           []Item[T]
	selectedIndex   int
	scrollOffset    int
	maxVisibleItems int
	position        sdl.Point
}

func NewList[T any](renderer *sdl.Renderer, maxVisibleItems int, position sdl.Point, itemFormatter fmtFunc[T]) *List[T] {
	return &List[T]{
		renderer:        renderer,
		itemFormatter:   itemFormatter,
		maxVisibleItems: maxVisibleItems,
		items:           []Item[T]{},
		position:        position,
	}
}

func (l *List[T]) SetItems(items []Item[T]) {
	l.items = items
	l.selectedIndex = 0
	l.scrollOffset = 0
}

func (l *List[T]) ScrollDown() {
	if l.selectedIndex < len(l.items)-1 {
		l.selectedIndex++
		if l.selectedIndex >= l.scrollOffset+l.maxVisibleItems {
			l.scrollOffset++
		}
	}
}

func (l *List[T]) ScrollUp() {
	if l.selectedIndex > 0 {
		l.selectedIndex--
		if l.selectedIndex < l.scrollOffset {
			l.scrollOffset--
		}
	}
}

func (l *List[T]) Draw(primaryColor sdl.Color, selectedColor sdl.Color) {
	// Draw the items
	startIndex := l.scrollOffset
	endIndex := startIndex + l.maxVisibleItems
	if endIndex > len(l.items) {
		endIndex = len(l.items)
	}
	visibleItems := l.items[startIndex:endIndex]

	for index, item := range visibleItems {
		color := primaryColor
		if index+startIndex == l.selectedIndex {
			color = selectedColor
		}
		itemText := l.itemFormatter(index+startIndex, item)
		textSurface, err := uilib.RenderText(itemText, color, config.ListFont)
		if err != nil {
			output.Printf("Error rendering text: %v\n", err)
			return
		}
		defer textSurface.Free()

		texture, err := l.renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			output.Printf("Error creating texture: %v\n", err)
			return
		}
		defer func() { _ = texture.Destroy() }()

		_ = l.renderer.Copy(texture, nil, &sdl.Rect{X: l.position.X, Y: l.position.Y + 30*int32(index), W: textSurface.W, H: textSurface.H})
	}
}

func (l *List[T]) GetSelectedIndex() int {
	return l.selectedIndex
}

func (l *List[T]) GetScrollOffset() int {
	return l.scrollOffset
}

func (l *List[T]) SelectedValue() T {
	return l.items[l.selectedIndex].Value
}

func (l *List[T]) GetValues() []T {
	values := make([]T, 0, len(l.items))
	for _, item := range l.items {
		values = append(values, item.Value)
	}
	return values
}
