package main

import (
	"fmt"
	"regexp"
)

type color uint32
type colors []color

func (a colors) equal(b colors) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

type board [][]color

type placement []colors

func (p placement) equal(q placement) bool {
	if len(p) != len(q) {
		return false
	}
	for i, c := range p {
		if !c.equal(q[i]) {
			return false
		}
	}
	return true
}

func (p placement) mirror() placement {
	q := make([]colors, len(p))
	for i, line := range p {
		for j := range line {
			q[i] = append(q[i], line[len(line)-j-1])
		}
	}
	return q
}

func (p placement) rotate() placement {
	q := make([]colors, len(p[0]))
	for y := 0; y < len(p[0]); y++ {
		for x := len(p) - 1; x >= 0; x-- {
			q[y] = append(q[y], p[x][y])
		}
	}
	return q
}

func (p placement) debug() {
	for _, line := range p {
		for _, c := range line {
			if c == 0 {
				fmt.Print(".")
			} else {
				fmt.Print("X")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

type piece []placement

func (p piece) contains(pl placement) bool {
	for _, existing := range p {
		if existing.equal(pl) {
			return true
		}
	}
	return false
}

func (p *piece) add(pl placement) {
	if !p.contains(pl) {
		*p = append(*p, pl)
	}
}

var pieceChunk *regexp.Regexp = regexp.MustCompile(`[X.]+`)

func makePiece(col color, spec string) piece {
	p := placement{}
	for _, chunk := range pieceChunk.FindAllString(spec, -1) {
		row := colors{}
		for i := range chunk {
			if chunk[i] == 'X' {
				row = append(row, col)
			} else {
				row = append(row, 0)
			}
		}
		p = append(p, row)
	}
	placements := piece{p}
	for i := 0; i < 4; i++ {
		placements.add(p)
		placements.add(p.mirror())
		p = p.rotate()
	}
	return placements
}

var pieces = []piece{
	makePiece(0xFF4500,
		`.X.
		 XXX
		 .X.`),
	makePiece(0x0000FF,
		`XX
		 X.
		 X.
		 X.`),
	makePiece(0x4B0082,
		`.XX
		 XX.
		 X..`),
	makePiece(0x000001,
		`XX
		 XX
		 X.`),
	makePiece(0x61B329,
		`XX
		 .X
		 .X
		 .X
		 XX`),
	makePiece(0x42C0FB,
		`XX.
		 .X.
		 XXX`),
	makePiece(0x215E21,
		`.XX
		 XXX
		 XX.`),
	makePiece(0xDDDDDD,
		`XXX
		 .X.
		 .X.`),
	makePiece(0xFF7722,
		`X..
		 XXX
		 XXX`),
	makePiece(0x6B4226,
		`XXXX
		 X.XX
		 X...`),
	makePiece(0xFFE600,
		`XXX
		 ..X
		 ..X`),
}

func main() {
	for _, p := range pieces {
		for _, pl := range p {
			pl.debug()
		}
	}
}
