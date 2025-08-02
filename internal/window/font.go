package window

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var fontSource FontSource

type FontSource struct {
	normal *text.GoTextFaceSource
	bold   *text.GoTextFaceSource
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

func init() {
	normal := loadFontFaceSource("/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc", 2)
	bold := loadFontFaceSource("/System/Library/Fonts/ヒラギノ角ゴシック W6.ttc", 2)

	fontSource = FontSource{
		normal: normal,
		bold:   bold,
	}
}
