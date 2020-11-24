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
		Object:   NewObject(("/earth.png")),
		Center:   image.Point{gameWidth / 2, gameHeight / 2},
		Impacted: false,
	}

	explosion := &Explosion{
		Object:    NewObject("/explosion.png"),
		Frame:     1,
		Exploding: false,
	}
	explosion.Radius = float64(explosion.Image.Bounds().Dy() / 2)

	asteroid := &Asteroid{
		Object:    NewObject(("/asteroid.png")),
		Angle:     rand.Float64() * math.Pi * 2,
		Distance:  earth.Radius * 2,
		Explosion: explosion,
		Alive:     true,
	}

	crosshair := &Crosshair{Object: NewObject(("/crosshair.png"))}

	gotext := NewObject("/gameover.png")
	gotext.Op.GeoM.Translate(
		float64(gameWidth/2-gotext.Image.Bounds().Dx()/2),
		float64(gameHeight/2-gotext.Image.Bounds().Dy()/2),
	)

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		GameOver:  false,
		Rotation:  0,
		Moon:      moon,
		Earth:     earth,
		Asteroid:  asteroid,
		Crosshair: crosshair,
		GOText:    gotext,
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
	Moon      *Moon
	Earth     *Earth
	Asteroid  *Asteroid
	GameOver  bool
	Crosshair *Crosshair
	GOText    *Object
}

// Update calculates game logic
func (g *Game) Update() error {

	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Game over
	if g.Earth.Impacted && !g.Asteroid.Alive {
		g.GameOver = true
	}

	// Asteroid impacts earth
	if g.Asteroid.Distance > 0 {
		g.Asteroid.Distance = g.Asteroid.Distance - 1
		g.Rotation = g.Rotation - 0.02
	} else if g.Asteroid.Alive {
		g.Earth.Impacted = true
		g.Asteroid.Explosion.Exploding = true
	}

	// Update object positions
	g.Earth.Update(g)
	g.Moon.Update(g)
	g.Asteroid.Update(g)
	g.Asteroid.Explosion.Update(g)
	g.Crosshair.Update(g)

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {

	// Draw game objects
	g.Earth.Draw(screen)
	g.Moon.Draw(screen)
	g.Asteroid.Draw(screen)
	g.Asteroid.Explosion.Draw(screen)
	g.Crosshair.Draw(screen)

	if g.GameOver {
		screen.DrawImage(g.GOText.Image, g.GOText.Op)
	}
	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
