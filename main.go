//go:generate statik -src=./assets -include=*.png

package main

import (
	"errors"
	"fmt"
	"image"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

	const howMany int = 200
	asteroids := make(Asteroids, 0, howMany)
	for i := 0; i < howMany; i++ {
		explosion := &Explosion{
			Object:    NewObject("/explosion.png"),
			Frame:     1,
			Exploding: false,
			Done:      false,
		}
		explosion.Radius = float64(explosion.Image.Bounds().Dy() / 2)

		asteroids = append(asteroids, &Asteroid{
			Object:    NewObject(("/asteroid.png")),
			Angle:     rand.Float64() * math.Pi * 2,
			Distance:  earth.Radius*2 + rand.Float64()*earth.Radius*20,
			Explosion: explosion,
			Alive:     true,
			Impacting: false,
		})
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
		Count:     0,
		Moon:      moon,
		Earth:     earth,
		Asteroids: asteroids,
		Crosshair: crosshair,
		GOText:    gotext,
		Entities: []Entity{
			moon,
			earth,
			asteroids,
			crosshair,
		},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// An Entity represents anything that can update itself in the game and draw
// itself to the main screen
type Entity interface {
	Update(*Game)
	Draw(*ebiten.Image)
}

// Game represents the main game state
type Game struct {
	Width     int
	Height    int
	Rotation  float64
	Count     int
	Moon      *Moon
	Earth     *Earth
	Asteroids Asteroids
	GameOver  bool
	Crosshair *Crosshair
	GOText    *Object
	Entities  []Entity
}

// Update calculates game logic
func (g *Game) Update() error {

	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Impact logic
	if g.Asteroids.Alive() && g.Asteroids.Impacting() {
		g.Earth.Impacted = true
	}

	// Game over
	if g.Earth.Impacted {
		if g.Asteroids.Alive() {
			for _, v := range g.Asteroids {
				v.Explosion.Exploding = true
			}
		} else {
			g.GameOver = true
		}
	}

	// Global rotation for orbiting bodies
	g.Rotation = g.Rotation - 0.02

	// Update object positions
	for _, v := range g.Entities {
		v.Update(g)
	}

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {

	// Draw game objects
	for _, v := range g.Entities {
		v.Draw(screen)
	}

	if g.GameOver {
		screen.DrawImage(g.GOText.Image, g.GOText.Op)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%d", g.Count))
	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
