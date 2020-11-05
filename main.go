package main

import (
	"errors"
	"log"

	"github.com/hajimehoshi/ebiten"
)

func main() {
	if err := ebiten.Run(update, 320, 240, 2, "Lunar Defence"); err != nil {
		log.Fatal(err)
	}
}

func update(screen *ebiten.Image) error {

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	// TODO: draw stuff here

	return nil
}
