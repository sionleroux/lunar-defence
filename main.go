//go:generate statik -src=. -include=*.png

package main

import (
	"errors"
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	_ "github.com/jatekalkotok/lunar-defence/statik"
	"github.com/rakyll/statik/fs"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")

	gameWidth, gameHeight := 1280, 960
	rand.Seed(time.Now().UnixNano())

	moon := &Moon{
		loadImage("/moon.png"),
	}

	earth := &Earth{
		loadImage("/earth.png"),
		0,
		image.Point{gameWidth / 2, gameHeight / 2},
	}

	asteroid := &Asteroid{
		loadImage("/asteroid.png"),
		rand.Float64() * math.Pi * 2,
		float64(moon.image.Bounds().Dx()) * 2,
	}

	game := &Game{
		gameWidth, gameHeight,
		moon,
		earth,
		asteroid,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	width    int
	height   int
	moon     *Moon
	earth    *Earth
	asteroid *Asteroid
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Asteroid collision TODO: it doesn't stop at the right place
	if g.asteroid.D <= float64(-g.moon.image.Bounds().Dx()*2) {
		return nil
	}

	g.earth.R = g.earth.R - 0.02
	g.asteroid.D = g.asteroid.D - 1

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {

	// Position earth
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		-float64(g.earth.image.Bounds().Dx())/2,
		-float64(g.earth.image.Bounds().Dy())/2,
	)
	op.GeoM.Rotate(g.earth.R)
	op.GeoM.Translate(float64(g.earth.XY.X), float64(g.earth.XY.Y))
	screen.DrawImage(g.earth.image, op)

	// Position moon
	op.GeoM.Reset()
	op.GeoM.Translate(float64(g.earth.XY.X), float64(g.earth.XY.Y))
	opR := &ebiten.DrawImageOptions{}
	opR.GeoM.Translate(
		-float64(g.earth.image.Bounds().Dx())/2-float64(g.moon.image.Bounds().Dx())*2,
		-float64(g.earth.image.Bounds().Dy())/2-float64(g.moon.image.Bounds().Dy())*2,
	)
	opR.GeoM.Rotate(g.earth.R / 3)
	opR.GeoM.Concat(op.GeoM)
	screen.DrawImage(g.moon.image, opR)

	// Position asteroid
	op.GeoM.Reset()
	op.GeoM.Translate(
		-float64(g.earth.image.Bounds().Dx())/2-g.asteroid.D,
		-float64(g.earth.image.Bounds().Dy())/2-g.asteroid.D,
	)
	op.GeoM.Rotate(g.asteroid.R)
	op.GeoM.Translate(float64(g.earth.XY.X), float64(g.earth.XY.Y))
	screen.DrawImage(g.asteroid.image, op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.width, g.height
}

// Moon is moon
type Moon struct {
	image *ebiten.Image
}

// Earth is earth
type Earth struct {
	image *ebiten.Image
	R     float64
	XY    image.Point
}

// Asteroid is asteroid
type Asteroid struct {
	image *ebiten.Image
	R     float64
	D     float64
}

func loadImage(name string) *ebiten.Image {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	file, err := statikFs.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(raw)
}
