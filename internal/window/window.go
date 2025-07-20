package window

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pishiko/tenmusu/internal/html"
)

type FontSource struct {
	normal *text.GoTextFaceSource
	bold   *text.GoTextFaceSource
}

type Window struct {
	fontSource FontSource
	tokens     []html.Token
	scrollY    int
}

func (b *Window) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		b.scrollY += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		b.scrollY -= 1
	}
	return nil
}
func (b *Window) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	ebitenutil.DebugPrint(screen, "FPS: "+fmt.Sprintf("%.2f", ebiten.ActualFPS()))

	layout := NewLayout(b.tokens, float64(b.scrollY), b.fontSource, screen.Bounds())
	drawables := layout.Drawables()
	for _, drawable := range drawables {

		if drawable.y+drawable.h < 0 || drawable.y > float64(screen.Bounds().Dy()) {
			continue
		}

		// Draw the character
		op := &text.DrawOptions{}
		color := ebiten.ColorScale{}
		color.SetR(0)
		color.SetG(0)
		color.SetB(0)
		color.SetA(1)
		op.ColorScale = color

		if drawable.style == "italic" {
			const deg = -15
			ang := deg * math.Pi / 180
			op.GeoM.Skew(ang, 0)
		}
		op.GeoM.Translate(drawable.x, drawable.y)
		text.Draw(screen, drawable.word, drawable.font, op)
	}

}
func (b *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func loadFontFaceSource(path string, index int) *text.GoTextFaceSource {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read font file: %v", err))
	}
	defer f.Close()

	fonts, err := text.NewGoTextFaceSourcesFromCollection(f)
	if err != nil {
		panic(err)
	}
	if index < 0 || index >= len(fonts) {
		panic(fmt.Sprintf("Invalid font index: %d", index))
	}
	// for i, src := range fonts {
	// 	md := src.Metadata()
	// 	fmt.Printf("Index=%d, Family=%q, Style=%v, Weight=%v\n",
	// 		i, md.Family, md.Style, md.Weight)
	// }
	return fonts[index]
}

func NewWindow(tokens []html.Token) *Window {

	normal := loadFontFaceSource("/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc", 2)
	bold := loadFontFaceSource("/System/Library/Fonts/ヒラギノ角ゴシック W6.ttc", 2)

	return &Window{
		fontSource: FontSource{
			normal: normal,
			bold:   bold,
		},
		tokens:  tokens,
		scrollY: 0,
	}
}

func Open(tokens []html.Token) {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("tenmusu")
	if err := ebiten.RunGame(NewWindow(tokens)); err != nil {
		panic(err)
	}
	println("Exiting...")
}
