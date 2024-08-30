package components

import (
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/output"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

type Item struct {
	Text  string
	ID    string
	Value string
}

type List struct {
	renderer        *sdl.Renderer
	items           []Item
	selectedIndex   int
	scrollOffset    int
	itemFormatter   func(index int, item Item) string
	maxVisibleItems int
	position        sdl.Point
}

func NewList(renderer *sdl.Renderer, maxVisibleItems int, position sdl.Point, itemFormatter func(index int, item Item) string) *List {
	return &List{
		renderer:        renderer,
		itemFormatter:   itemFormatter,
		maxVisibleItems: maxVisibleItems,
		items:           []Item{},
		position:        position,
	}
}

func (l *List) SetItems(items []Item) {
	l.items = items
	l.selectedIndex = 0
	l.scrollOffset = 0
}

func (l *List) ScrollDown() {
	if l.selectedIndex < len(l.items)-1 {
		l.selectedIndex++
		if l.selectedIndex >= l.scrollOffset+l.maxVisibleItems {
			l.scrollOffset++
		}
	}
}

func (l *List) ScrollUp() {
	if l.selectedIndex > 0 {
		l.selectedIndex--
		if l.selectedIndex < l.scrollOffset {
			l.scrollOffset--
		}
	}
}

func (l *List) Draw(primaryColor sdl.Color, selectedColor sdl.Color) {
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
		defer texture.Destroy()

		l.renderer.Copy(texture, nil, &sdl.Rect{X: l.position.X, Y: l.position.Y + 30*int32(index), W: textSurface.W, H: textSurface.H})
	}
}

func (l *List) GetSelectedIndex() int {
	return l.selectedIndex
}

func (l *List) GetScrollOffset() int {
	return l.scrollOffset
}

func (l *List) GetItems() []Item {
	return l.items
}
