package main

import (
	"image"
	"image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/jatekalkotok/lunar-defence/statik"
	"github.com/rakyll/statik/fs"
)

// An Object is something that can be seen and positioned in the game
type Object struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Center image.Point
	Radius float64
}

// Overlaps reports whether o and p have a non-empty intersection
func (o *Object) Overlaps(p *Object) bool {
	diff := o.Center.Sub(p.Center)
	distance := math.Sqrt(math.Pow(float64(diff.X), 2) + math.Pow(float64(diff.Y), 2))
	if distance <= o.Radius+p.Radius {
		return true
	}
	return false
}

// NewObject makes a new game Object with fields calculated from the input image
// after laoding it from the statikFS
func NewObject(filename string) *Object {
	img := loadImage(filename)
	return NewObjectFromImage(img)
}

// NewObjectFromImage makes a new game Object with fields calculated from an
// already loaded image
func NewObjectFromImage(img *ebiten.Image) *Object {
	return &Object{
		Image:  img,
		Op:     &ebiten.DrawImageOptions{},
		Center: image.Pt(0, 0),
		Radius: float64(img.Bounds().Dx()) / 2,
	}
}

// Moon is our moon, orbiting around the earth
type Moon struct {
	*Object
}

// Update recalculates moon position
func (o Moon) Update(g *Game) {
	t := g.Rotation / 3
	d := g.Earth.Radius + o.Radius*5

	// Calculated centre for collision detection
	x := (d) * math.Cos(t)
	y := (d) * math.Sin(t)
	o.Center = image.Pt(
		int(x)+g.Width/2,
		int(y)+g.Height/2,
	)

	// Spin the moon
	// Re-translate GeoM
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)
	o.Op.GeoM.Rotate(t)
	o.Op.GeoM.Translate(o.Radius, o.Radius)
	o.Op.GeoM.Translate(float64(o.Center.X), float64(o.Center.Y))
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)

	for _, v := range g.Asteroids {
		if o.Overlaps(v.Object) && v.Alive && !v.Explosion.Exploding {
			v.Explosion.Exploding = true
			g.Count--
		}
	}
}

// Draw renders a Moon to the screen
func (o *Moon) Draw(screen *ebiten.Image) {
	screen.DrawImage(o.Image, o.Op)
}

// Earth is the earth, our home planet
type Earth struct {
	*Object
	Center   image.Point
	Impacted bool
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

// Draw renders a Earth to the screen
func (o *Earth) Draw(screen *ebiten.Image) {
	if !o.Impacted {
		screen.DrawImage(o.Image, o.Op)
	}
}

// Pt is a shortcut for the Earth's X and Y coordinates
func (o Earth) Pt() (X, Y float64) {
	return float64(o.Center.X), float64(o.Center.Y)
}

// Asteroid is an asteroid on impact course with the Earth
type Asteroid struct {
	*Object
	Angle     float64
	Distance  float64
	Explosion *Explosion
	Alive     bool
	Impacting bool
}

// Update recalculates Asteroid position
func (o *Asteroid) Update(g *Game) {
	const RotationSpeed float64 = 3

	// Asteroid impacts earth
	if o.Distance > 0 {
		o.Distance = o.Distance - 1
	} else if o.Alive {
		o.Impacting = true
		o.Explosion.Exploding = true
	}

	// Calculated centre for collision detection
	t := o.Angle
	d := o.Distance + g.Earth.Radius
	x := (d) * math.Cos(t)
	y := (d) * math.Sin(t)
	o.Center = image.Pt(
		int(x)+g.Width/2,
		int(y)+g.Height/2,
	)

	// Re-translate GeoM
	o.Op.GeoM.Reset()

	// Spin the asteroid
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)
	o.Op.GeoM.Rotate(g.Rotation * RotationSpeed)
	o.Op.GeoM.Translate(o.Radius, o.Radius)

	// Move to newly calculated x, y with image offset to center
	o.Op.GeoM.Translate(float64(o.Center.X), float64(o.Center.Y))
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)

	// Handle Explosion
	o.Explosion.Update(g, o)
	if o.Explosion.Done && o.Alive {
		o.Alive = false
	}
}

// Draw renders a Asteroid to the screen
func (o *Asteroid) Draw(screen *ebiten.Image) {
	if o.Alive {
		screen.DrawImage(o.Image, o.Op)
		o.Explosion.Draw(screen)
	}
}

// Asteroids are multiple of a single Asteroid
type Asteroids []*Asteroid

// Update updates all the Asteroids
func (as Asteroids) Update(g *Game) {
	for _, v := range as {
		v.Update(g)
	}

	// TODO: delete dead asteroids
	// append(s[:index], s[index+1:]...)

}

// Draw updates all the Asteroids
func (as Asteroids) Draw(screen *ebiten.Image) {
	for _, v := range as {
		v.Draw(screen)
	}
}

// Alive returns true if any Asteroids are alive
func (as Asteroids) Alive() bool {
	for _, v := range as {
		if v.Alive {
			return true
		}
	}
	return false
}

// Impacting returns true if any Asteroids are impacting
func (as Asteroids) Impacting() bool {
	for _, v := range as {
		if v.Impacting {
			return true
		}
	}
	return false
}

// An Explosion is an animated impact explosion
type Explosion struct {
	*Object
	Frame     int
	Exploding bool
	Done      bool
}

// Update sets positioning and animation for Explosions
func (o *Explosion) Update(g *Game, a *Asteroid) {
	o.Center.X, o.Center.Y = a.Center.X, a.Center.Y
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(float64(o.Center.X), float64(o.Center.Y))
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)

	if o.Exploding {
		if o.Frame < 7 {
			o.Frame++
		} else {
			o.Exploding = false
			o.Done = true
		}
	}
}

// Draw renders an Explosion to the screen
func (o *Explosion) Draw(screen *ebiten.Image) {
	const frameSize int = 87
	if o.Exploding {
		screen.DrawImage(o.Image.SubImage(image.Rect(
			o.Frame*frameSize, 0, // top-left
			(1+o.Frame)*frameSize, frameSize, // bottom-right
		)).(*ebiten.Image), o.Op)
	}
}

// The Crosshair is a target showing where the the player will shoot
type Crosshair struct {
	*Object
}

// Update recalculates the crosshair position
func (o *Crosshair) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Center = image.Pt(ebiten.CursorPosition())
	o.Op.GeoM.Translate(
		float64(o.Center.X)-o.Radius,
		float64(o.Center.Y)-o.Radius,
	)
	if clicked() {
		for _, v := range g.Asteroids {
			if o.Overlaps(v.Object) && v.Alive && !v.Explosion.Exploding {
				v.Explosion.Exploding = true
				g.Count--
			}
		}
	}
}

// Draw renders a Crosshair to the screen
func (o *Crosshair) Draw(screen *ebiten.Image) {
	screen.DrawImage(o.Image, o.Op)
}

// Shorthand for when the left mouse button has just been clicked
func clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// Load an image from statikFS into an ebiten Image object
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
