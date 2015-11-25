// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package main

import (
	"image"
	"log"
	"math/rand"

	_ "image/png"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
)

const (
	tileWidth, tileHeight = 16, 16 // width and height of each tile
	tilesX, tilesY        = 16, 16 // number of horizontal tiles

	gopherTile = 1 // which tile the gopher is standing on (0-indexed)

	initScrollV = 1     // initial scroll velocity
	scrollA     = 0.001 // scroll accelleration

	groundChangeProb = 5 // 1/probability of ground height change
	groundMin        = tileHeight * (tilesY - 2*tilesY/5)
	groundMax        = tileHeight * tilesY
	initGroundY      = tileHeight * (tilesY - 1)
)

type Game struct {
	scroll struct {
		x float32 // x-offset
		v float32 // velocity
	}
	groundY  [tilesX + 3]float32 // ground y-offsets
	lastCalc clock.Time          // when we last calculated a frame
}

func NewGame() *Game {
	var g Game
	g.reset()
	return &g
}

func (g *Game) reset() {
	g.scroll.x = 0
	g.scroll.v = initScrollV
	for i := range g.groundY {
		g.groundY[i] = initGroundY
	}
}

func (g *Game) Scene(eng sprite.Engine) *sprite.Node {
	texs := loadTextures(eng)

	scene := &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	newNode := func(fn arrangerFunc) {
		n := &sprite.Node{Arranger: arrangerFunc(fn)}
		eng.Register(n)
		scene.AppendChild(n)
	}

	// The ground.
	for i := range g.groundY {
		i := i
		// The top of the ground.
		newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, texs[texGround])
			eng.SetTransform(n, f32.Affine{
				{tileWidth, 0, float32(i)*tileWidth - g.scroll.x},
				{0, tileHeight, g.groundY[i]},
			})
		})
		// The earth beneath.
		newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, texs[texEarth])
			eng.SetTransform(n, f32.Affine{
				{tileWidth, 0, float32(i)*tileWidth - g.scroll.x},
				{0, tileHeight * tilesY, g.groundY[i] + tileHeight},
			})
		})
	}

	// The gopher.
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, texs[texGopher])
		eng.SetTransform(n, f32.Affine{
			{tileWidth, 0, tileWidth * gopherTile},
			{0, tileHeight, 0},
		})
	})

	return scene
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }

const (
	texGopher = iota
	texGround
	texEarth
)

func loadTextures(eng sprite.Engine) []sprite.SubTex {
	a, err := asset.Open("placeholder-sprites.png")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	m, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(m)
	if err != nil {
		log.Fatal(err)
	}

	const n = 128
	return []sprite.SubTex{
		texGopher: sprite.SubTex{t, image.Rect(1+0, 0, n-1, n)},
		texGround: sprite.SubTex{t, image.Rect(1+n*3, 0, n*4-1, n)},
		texEarth:  sprite.SubTex{t, image.Rect(1+n*4, 0, n*5-1, n)},
	}
}

func (g *Game) Update(now clock.Time) {
	// Compute game states up to now.
	for ; g.lastCalc < now; g.lastCalc++ {
		g.calcFrame()
	}
}

func (g *Game) calcFrame() {
	g.calcScroll()
}

func (g *Game) calcScroll() {
	// Compute velocity.
	g.scroll.v += scrollA

	// Compute offset.
	g.scroll.x += g.scroll.v

	// Create new ground tiles if we need to.
	for g.scroll.x > tileWidth {
		g.newGroundTile()
	}
}

func (g *Game) newGroundTile() {
	// Compute next ground y-offset.
	next := g.nextGroundY()

	// Shift ground tiles to the left.
	g.scroll.x -= tileWidth
	copy(g.groundY[:], g.groundY[1:])
	g.groundY[len(g.groundY)-1] = next
}

func (g *Game) nextGroundY() float32 {
	prev := g.groundY[len(g.groundY)-1]
	if change := rand.Intn(groundChangeProb) == 0; change {
		return (groundMax-groundMin)*rand.Float32() + groundMin
	}
	return prev
}
