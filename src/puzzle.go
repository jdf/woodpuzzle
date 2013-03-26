package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
)

const BOARD_SIZE int = 8

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

func makeBoard() board {
	b := make(board, BOARD_SIZE)
	for i := range b {
		b[i] = make([]color, BOARD_SIZE)
	}
	return b
}

func (b board) String() string {
	var buf bytes.Buffer
	for _, row := range b {
		for _, c := range row {
			buf.WriteString(fmt.Sprintf("%06x ", c))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

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

type piece struct {
	placements       []placement
	c                color
	size             int
	currentPlacement int
	x, y             int
}

func (p piece) contains(pl placement) bool {
	for _, existing := range p.placements {
		if existing.equal(pl) {
			return true
		}
	}
	return false
}

func (p *piece) add(pl placement) {
	if !p.contains(pl) {
		p.placements = append(p.placements, pl)
	}
}

func (p *piece) reset() {
	p.currentPlacement = 0
	p.x = 0
	p.y = 0
}

func (p *piece) increment() bool {
	pl := p.placements[p.currentPlacement]
	if p.x+len(pl[0]) < BOARD_SIZE {
		p.x++
		return true
	}
	if p.y+len(pl) < BOARD_SIZE {
		p.x = 0
		p.y++
		return true
	}
	if p.currentPlacement+1 < len(p.placements) {
		p.x = 0
		p.y = 0
		p.currentPlacement++
		return true
	}
	return false
}

func (p piece) unwind(b board) {
	pl := p.placements[p.currentPlacement]

	for y := 0; y < len(pl); y++ {
		destRow := b[p.y+y]
		for x := 0; x < len(pl[0]); x++ {
			destX := p.x + x
			if destRow[destX] == p.c {
				destRow[destX] = 0
			}
		}
	}
}

func (p piece) place(b board) bool {
	pl := p.placements[p.currentPlacement]
	for y := 0; y < len(pl); y++ {
		srcRow := pl[y]
		destRow := b[p.y+y]
		for x := 0; x < len(srcRow); x++ {
			src := srcRow[x]
			if src == 0 {
				continue
			}
			destX := p.x + x
			if destRow[destX] == 0 {
				destRow[destX] = src
			} else {
				p.unwind(b)
				return false
			}
		}
	}

	return true
}

func (p piece) String() string {
	return fmt.Sprintf("<Piece color(%06x) currentPlacement(%d) x(%d) y(%d)>",
		p.c, p.currentPlacement, p.x, p.y)
}

var pieceChunk *regexp.Regexp = regexp.MustCompile(`[X.]+`)

func makePiece(col color, spec string) piece {
	size := 0
	p := placement{}
	for _, chunk := range pieceChunk.FindAllString(spec, -1) {
		row := colors{}
		for i := range chunk {
			if chunk[i] == 'X' {
				row = append(row, col)
				size++
			} else {
				row = append(row, 0)
			}
		}
		p = append(p, row)
	}
	result := piece{c: col, size: size}
	for i := 0; i < 4; i++ {
		result.add(p)
		result.add(p.mirror())
		p = p.rotate()
	}
	return result
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

type reverseBySize []piece

func (p reverseBySize) Len() int           { return len(p) }
func (p reverseBySize) Less(i, j int) bool { return p[j].size < p[i].size }
func (p reverseBySize) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var marker uint64 = 0

type floodboard [][]uint64

func makeFloodBoard() floodboard {
	b := make([][]uint64, BOARD_SIZE)
	for i := range b {
		b[i] = make([]uint64, BOARD_SIZE)
	}
	return b
}

func (f floodboard) String() string {
	var buf bytes.Buffer
	for _, row := range f {
		for _, c := range row {
			buf.WriteString(fmt.Sprintf("%03x ", c))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

var floodBoard floodboard = makeFloodBoard()

type node struct {
	x, y int
}

func (n node) String() string { return fmt.Sprintf("{%d,%d}", n.x, n.y) }

type stack struct {
	nodes []node
	top   int
}

func newStack() stack       { return stack{top: -1} }
func (s stack) empty() bool { return s.top == -1 }
func (s *stack) pop() node {
	n := s.nodes[s.top]
	s.top--
	return n
}
func (s *stack) push(n node) {
	s.top++
	if s.top >= len(s.nodes) {
		s.nodes = append(s.nodes, n)
	} else {
		s.nodes[s.top] = n
	}
}

// Floodfill empty squares and count them.
func fill(b board, x, y int) int {
	count := 0
	stack := newStack()
	stack.push(node{x, y})
	for !stack.empty() {
		n := stack.pop()
		if floodBoard[n.y][n.x] == marker {
			continue
		}
		floodBoard[n.y][n.x] = marker
		count++
		for yy := n.y - 1; yy <= n.y+1; yy++ {
			if yy < 0 || yy >= BOARD_SIZE {
				continue
			}
			for xx := n.x - 1; xx <= n.x+1; xx++ {
				if xx < 0 || xx >= BOARD_SIZE {
					continue
				}
				if !(xx == x && yy == y) && b[yy][xx] == 0 && floodBoard[yy][xx] != marker {
					stack.push(node{xx, yy})
				}
			}
		}
	}
	return count
}

// Use a flood-fill to find contiguous unfilled areas. If we find one that's smaller than
// the smallest piece, the board is unsolvable.
func isUnsolvable(b board, minSize int) bool {
	// Rather than clearing the floodBoard, just use a new "empty" marker.
	marker++

	for y, srcRow := range b {
		for x, srcCell := range srcRow {
			if srcCell != 0 {
				continue
			}
			if floodBoard[y][x] == marker {
				continue
			}
			if fill(b, x, y) < minSize {
				return true
			}
		}
	}
	return false
}

func main() {
	sort.Sort(reverseBySize(pieces))
	minSize := pieces[len(pieces)-1].size
	b := makeBoard()
	var iteration, solution uint64

	currentPiece := 0
	for true {
		iteration++
		if iteration%100000 == 0 {
			fmt.Print(".")
			if iteration%10000000 == 0 {
				for _, p := range pieces {
					fmt.Print(p.currentPlacement)
				}
				fmt.Println()
			}
		}
		if pieces[currentPiece].place(b) {
			if currentPiece == len(pieces)-1 {
				solution++;
				// Solution!
				fmt.Println()
				fmt.Println()
				fmt.Println(b.String())
				fmt.Println()
				fmt.Println()
				ioutil.WriteFile(fmt.Sprintf("solutions/%02d.txt", solution),
					[]byte(b.String()), 0644)
			} else if isUnsolvable(b, minSize) {
				pieces[currentPiece].unwind(b)
			} else {
				currentPiece++
				continue
			}
		}
		for !pieces[currentPiece].increment() {
			pieces[currentPiece].reset()
			currentPiece--
			if currentPiece < 0 {
				break
			}
			pieces[currentPiece].unwind(b)
		}
		if currentPiece < 0 {
			break
		}
	}
	fmt.Println("All done.")
}
