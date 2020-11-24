//go:generate statik -src=./assets -include=*.png

package main

import (
	"errors"
	"image"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	gameWidth, gameHeight := 1280, 960
	rand.Seed(time.Now().UnixNano())

	moon := &Moon{Object: NewObject("/moon.png")}

	earth := &Earth{
		Object: NewObject(("/earth.png")),
		Center: image.Point{gameWidth / 2, gameHeight / 2},
	}

	asteroid := &Asteroid{
		Object:   NewObject(("/asteroid.png")),
		Angle:    rand.Float64() * math.Pi * 2,
		Distance: earth.Radius * 2,
	}

	crosshair := &Crosshair{Object: NewObject(("/crosshair.png"))}

	explosion := &Explosion{
		Object: NewObject("/explosion.png"),
		Frame:  1,
	}
	explosion.Radius = float64(explosion.Image.Bounds().Dy() / 2)

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		Rotation:  0,
		Exploding: false,
		Moon:      moon,
		Earth:     earth,
		Asteroid:  asteroid,
		AAlive:    true,
		Crosshair: crosshair,
		Explosion: explosion,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width     int
	Height    int
	Rotation  float64
	Exploding bool
	Moon      *Moon
	Earth     *Earth
	Asteroid  *Asteroid
	AAlive    bool
	Crosshair *Crosshair
	Explosion *Explosion
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	if g.Asteroid.Distance > 0 {
		g.Asteroid.Distance = g.Asteroid.Distance - 1
		g.Rotation = g.Rotation - 0.02
	} else if g.AAlive {
		g.Exploding = true
	}

	// Update object positions
	g.Earth.Update(g)
	g.Moon.Update(g)
	g.Asteroid.Update(g)
	g.Crosshair.Update(g)
	g.Explosion.Update(g)

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.Earth.Image, g.Earth.Op)
	screen.DrawImage(g.Moon.Image, g.Moon.Op)
	if g.AAlive {
		screen.DrawImage(g.Asteroid.Image, g.Asteroid.Op)
	}
	screen.DrawImage(g.Crosshair.Image, g.Crosshair.Op)
	if g.Exploding {
		frameWidth := 87
		screen.DrawImage(g.Explosion.Image.SubImage(image.Rect(
			g.Explosion.Frame*frameWidth,
			0,
			(1+g.Explosion.Frame)*frameWidth,
			frameWidth,
		)).(*ebiten.Image), g.Explosion.Op)
	}
	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
