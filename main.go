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

	moonFile, err := os.Open("moon.png")
	if err != nil {
		log.Fatal(err)
	}

	moonRaw, err := png.Decode(moonFile)
	if err != nil {
		log.Fatal(err)
	}

	moon := ebiten.NewImageFromImage(moonRaw)

	earthFile, err := os.Open("earth.png")
	if err != nil {
		log.Fatal(err)
	}

	earthRaw, err := png.Decode(earthFile)
	if err != nil {
		log.Fatal(err)
	}

	earth := ebiten.NewImageFromImage(earthRaw)

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
