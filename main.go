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
		&ebiten.DrawImageOptions{},
	}

	earth := &Earth{
		loadImage("/earth.png"),
		&ebiten.DrawImageOptions{},
		0,
		image.Point{gameWidth / 2, gameHeight / 2},
	}

	asteroid := &Asteroid{
		loadImage("/asteroid.png"),
		&ebiten.DrawImageOptions{},
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
	g.earth.Update()
	screen.DrawImage(g.earth.image, g.earth.op)

	g.moon.Update(g.earth)
	screen.DrawImage(g.moon.image, g.moon.op)

	g.asteroid.Update(g.earth)
	screen.DrawImage(g.asteroid.image, g.asteroid.op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.width, g.height
}

// Moon is moon
type Moon struct {
	image *ebiten.Image
	op    *ebiten.DrawImageOptions
}

// Update recalculates moon position
func (o Moon) Update(e *Earth) {
	o.op.GeoM.Reset()
	o.op.GeoM.Translate(
		-float64(e.image.Bounds().Dx())/2-float64(o.image.Bounds().Dx())*2,
		-float64(e.image.Bounds().Dy())/2-float64(o.image.Bounds().Dy())*2,
	)
	o.op.GeoM.Rotate(e.R / 3)
	o.op.GeoM.Translate(float64(e.XY.X), float64(e.XY.Y))
}

// Earth is earth
type Earth struct {
	image *ebiten.Image
	op    *ebiten.DrawImageOptions
	R     float64
	XY    image.Point
}

// Update repositions Earth
func (o Earth) Update() {
	o.op.GeoM.Reset()
	o.op.GeoM.Translate(
		-float64(o.image.Bounds().Dx())/2,
		-float64(o.image.Bounds().Dy())/2,
	)
	o.op.GeoM.Rotate(o.R)
	o.op.GeoM.Translate(float64(o.XY.X), float64(o.XY.Y))
}

// Asteroid is asteroid
type Asteroid struct {
	image *ebiten.Image
	op    *ebiten.DrawImageOptions
	R     float64
	D     float64
}

// Update recalculates Asteroid position
func (o Asteroid) Update(e *Earth) {
	o.op.GeoM.Reset()
	o.op.GeoM.Translate(
		-float64(e.image.Bounds().Dx())/2-o.D,
		-float64(e.image.Bounds().Dy())/2-o.D,
	)
	o.op.GeoM.Rotate(o.R)
	o.op.GeoM.Translate(float64(e.XY.X), float64(e.XY.Y))
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
