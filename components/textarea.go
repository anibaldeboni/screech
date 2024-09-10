package components

import (
	"strings"

	"github.com/anibaldeboni/screech/output"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type TextArea struct {
	renderer        *sdl.Renderer
	text            string
	lines           []string
	scrollOffset    int
	maxVisibleLines int
	font            *ttf.Font
	maxWidth        int
}

func NewTextArea(renderer *sdl.Renderer, text string, font *ttf.Font, maxVisibleLines int, maxWidth int) *TextArea {
	component := &TextArea{
		renderer:        renderer,
		text:            text,
		maxVisibleLines: maxVisibleLines,
		font:            font,
		maxWidth:        maxWidth,
	}
	component.splitTextToLines()
	return component
}

func (t *TextArea) splitTextToLines() {
	lines := strings.Split(t.text, "\n")

	var wrappedLines []string
	for _, line := range lines {
		wrappedLines = append(wrappedLines, t.wrapLine(line, t.maxWidth)...)
	}

	t.lines = wrappedLines
}

func (t *TextArea) wrapLine(line string, maxWidth int) []string {
	var wrappedLines []string
	words := strings.Split(line, " ")
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		width, _, _ := t.font.SizeUTF8(currentLine + " " + word)
		if width <= maxWidth {
			currentLine += " " + word
		} else {
			wrappedLines = append(wrappedLines, currentLine)
			currentLine = word
		}
	}
	wrappedLines = append(wrappedLines, currentLine)
	return wrappedLines
}

func (t *TextArea) ScrollDown() {
	if t.scrollOffset < len(t.lines)-t.maxVisibleLines {
		t.scrollOffset++
	}
}

func (t *TextArea) ScrollUp() {
	if t.scrollOffset > 0 {
		t.scrollOffset--
	}
}

func (t *TextArea) Draw(primaryColor sdl.Color) {
	startIndex := t.scrollOffset
	endIndex := startIndex + t.maxVisibleLines
	if endIndex > len(t.lines) {
		endIndex = len(t.lines)
	}
	visibleLines := t.lines[startIndex:endIndex]

	for index, line := range visibleLines {
		textSurface, err := uilib.RenderText(line, primaryColor, t.font)
		if err != nil {
			output.Printf("Error rendering text: %v\n", err)
			return
		}
		defer textSurface.Free()

		texture, err := t.renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			output.Printf("Error creating texture: %v\n", err)
			return
		}
		defer func() { _ = texture.Destroy() }()

		_ = t.renderer.Copy(texture, nil, &sdl.Rect{X: 40, Y: 90 + 30*int32(index), W: textSurface.W, H: textSurface.H})
	}
}

func (t *TextArea) GetScrollOffset() int {
	return t.scrollOffset
}

func (t *TextArea) GetLines() []string {
	return t.lines
}
