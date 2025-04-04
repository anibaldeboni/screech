package components

import (
	"log"
	"strings"

	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/uilib"
	"github.com/veandco/go-sdl2/sdl"
)

type TextView struct {
	renderer *sdl.Renderer
	lines    []string
	YOffset  int
	size     TextViewSize
	position sdl.Point
}

type TextViewSize struct {
	Width  int
	Height int
}

func NewTextView(renderer *sdl.Renderer, size TextViewSize, position sdl.Point) *TextView {
	return &TextView{
		renderer: renderer,
		size:     size,
		lines:    []string{},
		position: position,
	}
}

func (t *TextView) parseLines(text []string) []string {
	out := make([]string, 0)
	for _, line := range text {
		for len(line) > t.size.Width {
			cut := t.size.Width
			if space := strings.LastIndex(line[:cut], " "); space != -1 {
				cut = space
			}
			out = append(out, line[:cut])
			line = strings.TrimSpace(line[cut:])
		}
		out = append(out, line)
	}
	return out
}

func (t *TextView) SetContent(text []string) {
	t.lines = text
	if t.YOffset > len(t.lines)-1 {
		t.GoToBottom()
	}
}

func (t *TextView) AddText(text string) {
	t.lines = append(t.lines, t.parseLines([]string{text})...)
	t.SetYOffset(t.maxYOffset())
	t.GoToBottom()
}

func (t *TextView) maxYOffset() int {
	return max(0, len(t.lines)-t.size.Height)
}

func (t *TextView) SetYOffset(n int) {
	t.YOffset = clamp(n, 0, t.maxYOffset())
}

func (t *TextView) AtTop() bool {
	return t.YOffset <= 0
}

func (t *TextView) AtBottom() bool {
	return t.YOffset >= t.maxYOffset()
}

func (t *TextView) ScrollDown(n int) {
	if t.AtBottom() || n == 0 || len(t.lines) == 0 {
		return
	}

	t.SetYOffset(t.YOffset + n)
}

func (t *TextView) ScrollUp(n int) {
	if t.AtTop() || n == 0 || len(t.lines) == 0 {
		return
	}

	t.SetYOffset(t.YOffset - n)
}

func (t *TextView) GoToBottom() {
	t.SetYOffset(t.maxYOffset())
}

func (t *TextView) visibleLines() (lines []string) {
	if len(t.lines) > 0 {
		top := max(0, t.YOffset)
		bottom := clamp(t.YOffset+t.size.Height, top, len(t.lines))
		lines = t.lines[top:bottom]
	}
	return lines
}

func (t *TextView) Draw(textColor sdl.Color) {
	for index, item := range t.visibleLines() {
		textSurface, err := uilib.RenderText(item, textColor, config.BodyFont)
		if err != nil {
			log.Printf("Error rendering text: %v\n", err)
			return
		}

		texture, err := t.renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			log.Printf("Error creating texture: %v\n", err)
			return
		}

		_ = t.renderer.Copy(texture, nil, &sdl.Rect{X: t.position.X, Y: t.position.Y + 30*int32(index), W: textSurface.W, H: textSurface.H})
		textSurface.Free()
		_ = texture.Destroy()
	}
}

func (t *TextView) GetScrollOffset() int {
	return t.YOffset
}

func (t *TextView) GetText() []string {
	return t.lines
}
