package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	BallSprite   *ebiten.Image
	BatSprite    *ebiten.Image
	StumpsSprite *ebiten.Image
	ScoreFont    *text.GoTextFace
)

//go:embed ball.png
var ballPNG []byte

//go:embed bat1.png
var batPNG []byte

//go:embed stumps.png
var stumpsPNG []byte

func init() {
	BallSprite = scaleImage(loadPNG(ballPNG), 0.7) // Make ball smaller (70% of original)
	BatSprite = scaleImage(loadPNG(batPNG), 1.3)   // Make bat bigger (130% of original)
	StumpsSprite = loadPNG(stumpsPNG)

	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	ScoreFont = &text.GoTextFace{
		Source: fontSource,
		Size:   24,
	}
}

func loadPNG(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(img)
}

func scaleImage(img *ebiten.Image, scale float64) *ebiten.Image {
	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scale)
	newHeight := int(float64(bounds.Dy()) * scale)
	
	scaledImg := ebiten.NewImage(newWidth, newHeight)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	scaledImg.DrawImage(img, op)
	
	return scaledImg
}
