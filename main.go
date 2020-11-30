// Copyright 2020 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

//go:generate statik -src=./assets -include=*.png

package main

import (
	"errors"
	"fmt"
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
	"gopkg.in/ini.v1"
)

var (
	HowManyStart       int     = 5
	EdgeOfScreenOffset float64 = 3
	DistanceVariance   float64 = 7
	TimeBetweenWaves   int     = 2
	WaveMultiplier     int     = 2
	RotationSpeed      float64 = 0.02
	MoonOrbitRatio     float64 = 2
	MoonOrbitDistance  float64 = 5
	AsteroidSpinRatio  float64 = 3
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	applyConfigs()

	gameWidth, gameHeight := 1280, 960
	rand.Seed(time.Now().UnixNano())
	howMany := HowManyStart // starting number of asteroids
	fontFace := loadFont()

	game := &Game{
		Width:      gameWidth,
		Height:     gameHeight,
		FontFace:   fontFace,
		Loading:    true,
		GameOver:   false,
		Breathless: false,
		Rotation:   0,
		Count:      0,
		Wave:       0,
		HowMany:    howMany,
		Moon:       nil,
		Earth:      nil,
		Asteroids:  nil,
		Crosshair:  nil,
		GOText:     nil,
		Entities:   nil,
	}

	go NewGame(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// NewGame sets up a new game object with default states and game objects
func NewGame(game *Game) {
	earth := &Earth{
		Object:   NewObject(("/earth.png")),
		Center:   image.Point{game.Width / 2, game.Height / 2},
		Impacted: false,
	}
	game.Earth = earth

	explosion := &Explosion{
		Object:    NewObjectFromImage(loadImage("/explosion.png")),
		Frame:     1,
		Exploding: false,
		Done:      false,
	}
	explosion.Radius = float64(explosion.Image.Bounds().Dy() / 2)
	game.Crosshair = &Crosshair{
		Object:    NewObject(("/crosshair.png")),
		Explosion: explosion,
	}

	game.Moon = &Moon{
		Object: NewObject("/moon.png"),
		Turret: &Turret{
			NewObject("/turret.png"),
		},
	}

	gotext := NewObject("/gameover.png")
	gotext.Op.GeoM.Translate(
		float64(game.Width/2-gotext.Image.Bounds().Dx()/2),
		float64(game.Height/2-gotext.Image.Bounds().Dy()/2),
	)
	game.GOText = gotext

	entities := []Entity{
		Asteroids{},
		game.Moon,
		game.Earth,
		game.Crosshair,
	}
	game.Entities = entities

	game.Loading = false
}

// NewAsteroids makes a fresh set of asteroids
func NewAsteroids(earthRadius float64, howMany int) Asteroids {
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

		edgeOfScreenOffset := earthRadius * EdgeOfScreenOffset
		distance := rand.Float64() * earthRadius * float64(howMany) / DistanceVariance
		asteroids = append(asteroids, &Asteroid{
			Object:    NewObjectFromImage(asteroidImage),
			Angle:     rand.Float64() * math.Pi * 2,
			Distance:  edgeOfScreenOffset + distance,
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
	Width      int
	Height     int
	Loading    bool
	FontFace   font.Face
	Rotation   float64
	Count      int
	Wave       int
	HowMany    int
	Moon       *Moon
	Earth      *Earth
	Asteroids  Asteroids
	GameOver   bool
	Breathless bool // when you need a break between waves
	Crosshair  *Crosshair
	GOText     *Object
	Entities   []Entity
}

// Update calculates game logic
func (g *Game) Update() error {

	// Pressing Esc any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	// Skip updating while the game is loading
	if g.Loading {
		return nil
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
		} else if !g.GameOver {
			g.GameOver = true
			log.Println("game over")
			g.Breathless = true
			takeABreath := time.NewTimer(time.Second)
			go func() {
				log.Println("waiting")
				<-takeABreath.C
				g.Breathless = false
			}()
		}
	}

	// Next wave
	if !g.GameOver && !g.Asteroids.Alive() && !g.Breathless && g.Wave > 0 {
		log.Println("wave passed")
		g.Wave++
		g.Breathless = true
		takeABreath := time.NewTimer(time.Second * time.Duration(TimeBetweenWaves))
		go func() {
			log.Println("waiting")
			<-takeABreath.C
			g.HowMany *= WaveMultiplier
			g.Restart()
			g.Breathless = false // needs to come after restart
		}()
	}

	// Global rotation for orbiting bodies
	g.Rotation = g.Rotation - RotationSpeed

	// Update object positions
	for _, v := range g.Entities {
		v.Update(g)
	}

	// On wave zero, click to start the game
	if g.Wave == 0 && clicked() {
		g.Wave++
		g.Restart()
	}

	// Game restart
	if g.GameOver && clicked() && !g.Breathless {
		g.Restart()
	}

	return nil
}

// Restart starts a new game with states reset
func (g *Game) Restart() {
	log.Printf("new wave: %d\n", g.HowMany)
	g.Count = g.HowMany
	g.Asteroids = NewAsteroids(g.Earth.Radius, g.HowMany)
	g.Entities[0] = g.Asteroids
	g.Earth.Impacted = false
	g.GameOver = false
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {

	if g.Loading {
		loadText := "LOADING..."
		loadTextF, _ := font.BoundString(g.FontFace, loadText)
		loadTextW := (loadTextF.Max.X - loadTextF.Min.X).Ceil() / 2
		loadTextH := (loadTextF.Max.Y - loadTextF.Min.Y).Ceil() / 2
		text.Draw(screen, loadText, g.FontFace, g.Width/2-loadTextW, g.Height/2-loadTextH, color.White)
		return
	}
	if !g.Loading && g.Wave == 0 {
		startText := "CLICK TO START"
		startTextF, _ := font.BoundString(g.FontFace, startText)
		startTextW := (startTextF.Max.X - startTextF.Min.X).Ceil() / 2
		startTextH := (startTextF.Max.Y - startTextF.Min.Y).Ceil() * 2
		text.Draw(screen, startText, g.FontFace, g.Width/2-startTextW, startTextH, color.White)
		creditsText := "By: Siôn le Roux www.sinisterstuf.org"
		creditsTextF, _ := font.BoundString(g.FontFace, creditsText)
		creditsTextW := (creditsTextF.Max.X - creditsTextF.Min.X).Ceil() / 2
		creditsTextH := (creditsTextF.Max.Y - creditsTextF.Min.Y).Ceil() * 2
		text.Draw(screen, creditsText, g.FontFace, g.Width/2-creditsTextW, g.Height-creditsTextH, color.White)
		titleText := "Lunar Defence"
		titleTextF, _ := font.BoundString(g.FontFace, titleText)
		titleTextW := (titleTextF.Max.X - titleTextF.Min.X).Ceil() / 2
		titleTextH := (titleTextF.Max.Y - titleTextF.Min.Y).Ceil() * 2
		text.Draw(screen, titleText, g.FontFace, g.Width/2-titleTextW, g.Height-titleTextH*2, color.White)
	}

	// Draw game objects
	for _, v := range g.Entities {
		v.Draw(screen)
	}

	if g.GameOver {
		screen.DrawImage(g.GOText.Image, g.GOText.Op)
	}

	// HUD and other text
	padding := 20
	f, _ := font.BoundString(g.FontFace, "00")
	h := (f.Max.Y - f.Min.Y).Ceil() * 2
	w := (f.Max.X - f.Min.X).Ceil() + padding
	text.Draw(screen, strconv.Itoa(g.Count), g.FontFace, padding, h, color.White)
	text.Draw(screen, strconv.Itoa(g.Wave), g.FontFace, g.Width-w, h, color.White)
	if g.Crosshair.CoolingDown && !g.Breathless { // TODO: this should be in Crosshair.Draw()
		missText := "MISSED: COOLING DOWN!"
		missTextF, _ := font.BoundString(g.FontFace, missText)
		missTextW := (missTextF.Max.X - missTextF.Min.X).Ceil() / 2
		text.Draw(screen, missText, g.FontFace, g.Width/2-missTextW, h, color.White)
	}
	if !g.GameOver && g.Breathless {
		tryAgain := fmt.Sprintf("WAVE %d", g.Wave)
		tryAgainF, _ := font.BoundString(g.FontFace, tryAgain)
		tryAgainW := (tryAgainF.Max.X - tryAgainF.Min.X).Ceil() / 2
		text.Draw(screen, tryAgain, g.FontFace, g.Width/2-tryAgainW, h, color.White)
	}
	if g.GameOver && !g.Breathless {
		tryAgain := "CLICK TO TRY AGAIN"
		tryAgainF, _ := font.BoundString(g.FontFace, tryAgain)
		tryAgainW := (tryAgainF.Max.X - tryAgainF.Min.X).Ceil() / 2
		text.Draw(screen, tryAgain, g.FontFace, g.Width/2-tryAgainW, h, color.White)
	}

	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

func applyConfigs() {
	cfg, err := ini.Load("lunar-defence.ini")
	log.Println(err)
	if err == nil {
		HowManyStart, _ = cfg.Section("").Key("HowManyStart").Int()
		EdgeOfScreenOffset, _ = cfg.Section("").Key("EdgeOfScreenOffset").Float64()
		DistanceVariance, _ = cfg.Section("").Key("DistanceVariance").Float64()
		TimeBetweenWaves, _ = cfg.Section("").Key("TimeBetweenWaves").Int()
		WaveMultiplier, _ = cfg.Section("").Key("WaveMultiplier").Int()
		RotationSpeed, _ = cfg.Section("").Key("RotationSpeed").Float64()
		MoonOrbitRatio, _ = cfg.Section("").Key("MoonOrbitRatio").Float64()
		MoonOrbitDistance, _ = cfg.Section("").Key("MoonOrbitDistance").Float64()
		AsteroidSpinRatio, _ = cfg.Section("").Key("AsteroidSpinRatio").Float64()
	}
}

func loadFont() font.Face {
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
	return fontface
}
