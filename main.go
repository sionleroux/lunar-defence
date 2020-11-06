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
	game := &Game{
		moon,
		0,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	moon  *ebiten.Image
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
	op.GeoM.Translate(float64(g.moonX), 0)
	screen.DrawImage(g.moon, op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return 640, 480
}
