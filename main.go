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

	moonImage := loadImage("/moon.png")
	moon := &Moon{
		Image:  moonImage,
		Op:     &ebiten.DrawImageOptions{},
		Radius: float64(moonImage.Bounds().Dx()),
	}

	earthImage := loadImage("/earth.png")
	earth := &Earth{
		Image:    earthImage,
		Op:       &ebiten.DrawImageOptions{},
		Radius:   float64(earthImage.Bounds().Dx()),
		Rotation: 0,
		Center:   image.Point{gameWidth / 2, gameHeight / 2},
	}

	asteroid := &Asteroid{
		Image:    loadImage("/asteroid.png"),
		Op:       &ebiten.DrawImageOptions{},
		Rotation: rand.Float64() * math.Pi * 2,
		Distance: moon.Radius * 2,
	}

	game := &Game{
		Width:    gameWidth,
		Height:   gameHeight,
		Moon:     moon,
		Earth:    earth,
		Asteroid: asteroid,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width    int
	Height   int
	Moon     *Moon
	Earth    *Earth
	Asteroid *Asteroid
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Asteroid collision TODO: it doesn't stop at the right place
	if g.Asteroid.Distance <= float64(-g.Moon.Radius*2) {
		return nil
	}

	g.Earth.Rotation = g.Earth.Rotation - 0.02
	g.Asteroid.Distance = g.Asteroid.Distance - 1

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	g.Earth.Update()
	screen.DrawImage(g.Earth.Image, g.Earth.Op)

	g.Moon.Update(g.Earth)
	screen.DrawImage(g.Moon.Image, g.Moon.Op)

	g.Asteroid.Update(g.Earth)
	screen.DrawImage(g.Asteroid.Image, g.Asteroid.Op)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Moon is moon
type Moon struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Radius float64
}

// Update recalculates moon position
func (o Moon) Update(earth *Earth) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-earth.Radius/2-o.Radius*2,
		-earth.Radius/2-o.Radius*2,
	)
	o.Op.GeoM.Rotate(earth.Rotation / 3)
	o.Op.GeoM.Translate(earth.Pt())
}

// Earth is earth
type Earth struct {
	Image    *ebiten.Image
	Op       *ebiten.DrawImageOptions
	Radius   float64
	Rotation float64
	Center   image.Point
}

// Update repositions Earth
func (o Earth) Update() {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-o.Radius/2,
		-o.Radius/2,
	)
	o.Op.GeoM.Rotate(o.Rotation)
	o.Op.GeoM.Translate(o.Pt())
}

// Pt is a shortcut for the Earth's X and Y coordinates
func (o Earth) Pt() (X, Y float64) {
	return float64(o.Center.X), float64(o.Center.Y)
}

// Asteroid is asteroid
type Asteroid struct {
	Image    *ebiten.Image
	Op       *ebiten.DrawImageOptions
	Rotation float64
	Distance float64
}

// Update recalculates Asteroid position
func (o Asteroid) Update(earth *Earth) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-earth.Radius/2-o.Distance,
		-earth.Radius/2-o.Distance,
	)
	o.Op.GeoM.Rotate(o.Rotation)
	o.Op.GeoM.Translate(earth.Pt())
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
