//go:generate statik -src=./assets -include=*.png

package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	gameWidth, gameHeight := 1280, 960
	rand.Seed(time.Now().UnixNano())

	earth := &Earth{
		Object:   NewObject(("/earth.png")),
		Center:   image.Point{gameWidth / 2, gameHeight / 2},
		Impacted: false,
	}

	moon := &Moon{Object: NewObject("/moon.png")}
	crosshair := &Crosshair{Object: NewObject(("/crosshair.png"))}
	asteroids := NewAsteroids(earth.Radius)

	gotext := NewObject("/gameover.png")
	gotext.Op.GeoM.Translate(
		float64(gameWidth/2-gotext.Image.Bounds().Dx()/2),
		float64(gameHeight/2-gotext.Image.Bounds().Dy()/2),
	)

	fontdata, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	fontface, err := opentype.NewFace(fontdata, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		FontFace:  fontface,
		GameOver:  false,
		Rotation:  0,
		Count:     0,
		Moon:      moon,
		Earth:     earth,
		Asteroids: asteroids,
		Crosshair: crosshair,
		GOText:    gotext,
		Entities: []Entity{
			asteroids,
			moon,
			earth,
			crosshair,
		},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// NewAsteroids makes a fresh set of asteroids
func NewAsteroids(earthRadius float64) Asteroids {
	const howMany int = 20
	asteroids := make(Asteroids, 0, howMany)
	asteroidImage := loadImage("/asteroid.png")
	explosionImage := loadImage("/explosion.png")
	for i := 0; i < howMany; i++ {
		explosion := &Explosion{
			Object:    NewObjectFromImage(explosionImage),
			Frame:     1,
			Exploding: false,
			Done:      false,
		}
		explosion.Radius = float64(explosion.Image.Bounds().Dy() / 2)

		asteroids = append(asteroids, &Asteroid{
			Object:    NewObjectFromImage(asteroidImage),
			Angle:     rand.Float64() * math.Pi * 2,
			Distance:  earthRadius*2 + rand.Float64()*earthRadius*2,
			Explosion: explosion,
			Alive:     true,
			Impacting: false,
		})
	}

	return asteroids
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
	FontFace  font.Face
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

	// Game restart
	if g.GameOver && clicked() {
		g.Count = 0
		g.Asteroids = NewAsteroids(g.Earth.Radius)
		g.Entities[0] = g.Asteroids
		g.Earth.Impacted = false
		g.GameOver = false
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

	f, _ := font.BoundString(g.FontFace, "I")
	h := (f.Max.Y - f.Min.Y).Ceil() * 2
	text.Draw(screen, strconv.Itoa(g.Count), g.FontFace, 20, h, color.White)
	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
