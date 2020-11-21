package main

import (
	"image"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	_ "github.com/jatekalkotok/lunar-defence/statik"
	"github.com/rakyll/statik/fs"
)

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
	log.Printf("loading %s\n", name)

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
