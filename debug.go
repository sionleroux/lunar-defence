// Copyright 2020 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

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

	ebitenutil.DrawRect(
		screen,
		float64(g.Crosshair.Center.X-20),
		float64(g.Crosshair.Center.Y-20),
		40,
		40,
		color.RGBA{0, 255, 0, 255},
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

func fps(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("FPS: %.0f, Tick: %.0f\n", ebiten.CurrentFPS(), ebiten.CurrentTPS()),
	)
}
