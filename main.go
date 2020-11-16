//go:generate statik -src=./assets -include=*.png

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

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		Rotation:  0,
		Exploding: false,
		Moon:      moon,
		Earth:     earth,
		Asteroid:  asteroid,
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
	} else {
		g.Exploding = true
	}

	// Update object positions
	g.Earth.Update(g)
	g.Moon.Update(g)
	g.Asteroid.Update(g)
	g.Crosshair.Update()
	g.Explosion.Update(g)

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.Earth.Image, g.Earth.Op)
	screen.DrawImage(g.Moon.Image, g.Moon.Op)
	screen.DrawImage(g.Asteroid.Image, g.Asteroid.Op)
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

// An Object is something that can be seen and positioned in the game
type Object struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Radius float64
}

// NewObject makes a new game Object with fields calculated from the input image
// after laoding it from the statikFS
func NewObject(filename string) *Object {
	image := loadImage(filename)
	return &Object{
		Image:  image,
		Op:     &ebiten.DrawImageOptions{},
		Radius: float64(image.Bounds().Dx()) / 2,
	}
}

// Moon is moon
type Moon struct {
	*Object
}

// Update recalculates moon position
func (o Moon) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-g.Earth.Radius-o.Radius*2,
		-g.Earth.Radius-o.Radius*2,
	)
	o.Op.GeoM.Rotate(g.Rotation / 3)
	o.Op.GeoM.Translate(g.Earth.Pt())
}

// Earth is earth
type Earth struct {
	*Object
	Center image.Point
}

// Update repositions Earth
func (o Earth) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-o.Radius,
		-o.Radius,
	)
	o.Op.GeoM.Rotate(g.Rotation)
	o.Op.GeoM.Translate(o.Pt())
}

// Pt is a shortcut for the Earth's X and Y coordinates
func (o Earth) Pt() (X, Y float64) {
	return float64(o.Center.X), float64(o.Center.Y)
}

// Asteroid is asteroid
type Asteroid struct {
	*Object
	Angle    float64
	Distance float64
}

// Update recalculates Asteroid position
func (o Asteroid) Update(g *Game) {
	const RotationSpeed float64 = 3
	o.Op.GeoM.Reset()

	// Spin the asteroid
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)
	o.Op.GeoM.Rotate(g.Rotation * RotationSpeed)

	// Move it back to where it was because maths is hard
	o.Op.GeoM.Translate(o.Radius, o.Radius)

	// Positions it at correct distance for angle correction
	o.Op.GeoM.Translate(
		-g.Earth.Radius+o.Radius*2-o.Distance,
		-g.Earth.Radius+o.Radius*2-o.Distance,
	)

	// Turn to correct angle
	o.Op.GeoM.Rotate(o.Angle)

	// Move post-rotation centre to match Earth's centre
	o.Op.GeoM.Translate(g.Earth.Pt())
}

// An Explosion is an animated impact explosion
type Explosion struct {
	*Object
	Frame int
}

// Update sets positioning and animation for Explosions
func (o *Explosion) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(-g.Earth.Radius, -g.Earth.Radius)
	o.Op.GeoM.Rotate(g.Asteroid.Angle)
	o.Op.GeoM.Translate(g.Earth.Pt())

	if g.Exploding {
		if o.Frame < 7 {
			o.Frame++
		} else {
			g.Exploding = false
		}
	}
}

// The Crosshair is a target showing where the the player will shoot
type Crosshair struct {
	*Object
	Radius float64
}

// Update recalculates the crosshair position
func (o Crosshair) Update() {
	o.Op.GeoM.Reset()
	mx, my := ebiten.CursorPosition()
	o.Op.GeoM.Translate(
		float64(mx)-o.Radius,
		float64(my)-o.Radius,
	)
}

func loadImage(name string) *ebiten.Image {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatalf("error initialising statikFS: %v\n", err)
	}

	file, err := statikFs.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatalf("error decoding file %s as PNG: %v\n", name, err)
	}

	return ebiten.NewImageFromImage(raw)
}
