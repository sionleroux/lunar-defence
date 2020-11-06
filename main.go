package main

import (
	"errors"
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

	game := &Game{
		moon,
		earth,
		0,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	moon  *ebiten.Image
	earth *ebiten.Image
	moonX int
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	g.moonX++

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.earth, op)
	op.GeoM.Translate(float64(g.moonX), 0)
	screen.DrawImage(g.moon, op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return 640, 480
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
