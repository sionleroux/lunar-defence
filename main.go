package main

import (
	"errors"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")

	moon := loadImage("moon.png")
	earth := loadImage("earth.png")

	gameWidth, gameHeight := 1280, 960
	game := &Game{
		gameWidth, gameHeight,
		moon,
		earth,
		0,
		0,
		image.Point{gameWidth / 2, gameHeight / 2},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	width   int
	height  int
	moon    *ebiten.Image
	earth   *ebiten.Image
	moonX   int
	earthR  int
	earthXY image.Point
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	g.moonX++
	g.earthR++

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		-float64(g.earth.Bounds().Dx()/2),
		-float64(g.earth.Bounds().Dy()/2),
	)
	op.GeoM.Rotate(float64(g.earthR))
	op.GeoM.Translate(float64(g.earthXY.X), float64(g.earthXY.Y))
	screen.DrawImage(g.earth, op)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(g.moonX), 0)
	screen.DrawImage(g.moon, op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.width, g.height
}

func loadImage(name string) *ebiten.Image {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(raw)
}
