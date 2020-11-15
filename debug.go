package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func debug(screen *ebiten.Image, g *Game) {
	ebitenutil.DrawRect(
		screen,
		float64(g.Width)/2-20,
		float64(g.Height)/2-20,
		40,
		40,
		color.RGBA{255, 255, 0, 255},
	)

	ebitenutil.DrawLine(
		screen,
		float64(g.Earth.Center.X),
		float64(g.Earth.Center.Y),
		float64(g.Earth.Center.X)+g.Earth.Radius+g.Asteroid.Distance,
		float64(g.Earth.Center.Y),
		color.White,
	)

	ebitenutil.DrawLine(
		screen,
		float64(g.Earth.Center.X),
		float64(g.Earth.Center.Y),
		float64(g.Earth.Center.X)+g.Earth.Radius,
		float64(g.Earth.Center.Y),
		color.RGBA{255, 0, 0, 255},
	)

	mx, my := ebiten.CursorPosition()
	ebitenutil.DrawLine(
		screen,
		float64(g.Earth.Center.X),
		float64(g.Earth.Center.Y),
		float64(mx),
		float64(my),
		color.RGBA{0, 255, 255, 255},
	)

	mdx := mx - g.Width/2
	mdy := my - g.Height/2
	ebitenutil.DebugPrint(screen, fmt.Sprintf("(%v, %v) d%.0f\n",
		mdx, mdy,
		math.Sqrt(math.Pow(float64(mdx), 2)+math.Pow(float64(mdy), 2)),
	))
}
