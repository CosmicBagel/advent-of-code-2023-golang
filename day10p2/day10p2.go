package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

/*
2d grid
info for each element:
- isPath
- flow
- originalCharacter
-

replace all none path elements with .

flood fill from edge to find outside edge of loop

follow loop, note direction flow, marking open slots as
	inside or outside the loop (if its not part of the loop)

add inside marks to list as you go

flood search from each inside mark (not infinitely obviously)

report count of found inside tiles


------

create blank grid with just the loop on it (using existing found path)

find left most part of the loop
	know left of this is out of the loop, right is in the loop

follow loop with this (adjusting perspective with bends / turns)
	when there's an empty tile, flood fill search for other inner tiles

*/

type Point struct {
	X int
	Y int
}

type NodeType int

const (
	Empty NodeType = iota
	Vertical
	Horizontal
	NEBend
	NWBend
	SWBend
	SEBend
	Start
)

type Node struct {
	north *Node
	east  *Node
	south *Node
	west  *Node

	nodeType       NodeType
	location       Point
	originalMarker rune
}

type Grid struct {
	height int
	width  int

	elements [][]byte
}

func makeGrid(width, height int) *Grid {
	g := Grid{
		width:    width,
		height:   height,
		elements: make([][]byte, height),
	}

	for y := 0; y < height; y++ {
		row := make([]byte, width)
		g.elements[y] = row
		for x := 0; x < width; x++ {
			row[x] = '.'
		}
	}
	return &g
}

func (g *Grid) getPoint(p Point) byte {
	return g.elements[p.Y][p.X]
}

func (g *Grid) setPoint(p Point, b byte) {
	g.elements[p.Y][p.X] = b
}

func (g *Grid) draw() {
	fmt.Printf("printing %d by %d grid\n", g.width, g.height)
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			fmt.Printf("%s", string(g.getPoint(Point{x, y})))
		}
		fmt.Print("\n")
	}
}

func main() {
	fmt.Println("day 10 p 2")
	// file_name := "example_inputA.txt" // expecting 4
	// file_name := "example_inputB.txt" // expecting 4
	// file_name := "example_inputC.txt" // expecting 8
	file_name := "example_inputD.txt" // expecting 10
	// file_name := "example_inputE.txt" // expecting 1
	// file_name := "example_inputF.txt" // expecting 1
	// file_name := "example_inputG.txt" // expecting 1
	// file_name := "example_inputH.txt" // expecting 1
	// file_name := "input.txt"

	file, err := os.Open(file_name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	startingNode, gridWidth, gridHeight := parse(scanner)
	grid := makeGrid(gridWidth, gridHeight)
	writePathToGrid(startingNode, grid)

	grid.draw()
	// result := findFarthestDistance(startingNode)

	// fmt.Printf("enclosed tiles: %d\n", result)
}

func parse(scanner *bufio.Scanner) (*Node, int, int) {
	var startingNode *Node = nil

	scanner.Scan()
	currentLine := scanner.Text()
	gridWidth := len(currentLine)

	aboveNodes := make([]*Node, 0, gridWidth)
	currentNodes := make([]*Node, 0, gridWidth)

	for i := 0; i < gridWidth; i++ {
		n := Node{
			location:       Point{i, -1},
			originalMarker: '.',
			nodeType:       Empty,
		}
		aboveNodes = append(aboveNodes, &n)
	}

	y := 0
	for {
		// fmt.Println(currentLine)
		for x, c := range currentLine {
			loc := Point{x, y}
			var aboveNode = aboveNodes[x]
			var leftNode *Node = nil
			if x > 0 {
				leftNode = currentNodes[x-1]
			}

			n := Node{
				location:       loc,
				originalMarker: c,
				nodeType:       Empty,
			}

			ant := aboveNode.nodeType
			aboveHasSouth := ant == Vertical || ant == SEBend || ant == SWBend || ant == Start

			leftHasEast := false
			if leftNode != nil {
				lnt := leftNode.nodeType
				leftHasEast = lnt == Horizontal || lnt == NEBend || lnt == SEBend || lnt == Start
			}

			switch c {
			case '|':
				n.nodeType = Vertical
				if aboveHasSouth {
					n.north = aboveNode
					aboveNode.south = &n
				}
			case '-':
				n.nodeType = Horizontal
				if leftHasEast {
					n.west = leftNode
					leftNode.east = &n
				}
			case 'L':
				n.nodeType = NEBend
				if aboveHasSouth {
					n.north = aboveNode
					aboveNode.south = &n
				}
			case 'J':
				n.nodeType = NWBend
				if leftHasEast {
					n.west = leftNode
					leftNode.east = &n
				}
				if aboveHasSouth {
					n.north = aboveNode
					aboveNode.south = &n
				}
			case '7':
				n.nodeType = SWBend
				if leftHasEast {
					n.west = leftNode
					leftNode.east = &n
				}
			case 'F':
				n.nodeType = SEBend
				// must be connected to by an east node or south node
			case 'S':
				n.nodeType = Start
				startingNode = &n
				if leftHasEast {
					n.west = leftNode
					leftNode.east = &n
				}
				if aboveHasSouth {
					n.north = aboveNode
					aboveNode.south = &n
				}
			}

			currentNodes = append(currentNodes, &n)
		}

		// for _, n := range currentNodes {
		// 	fmt.Printf("%+v\n", n)
		// }

		// clear and swap lines
		temp := aboveNodes[:0]
		aboveNodes = currentNodes
		currentNodes = temp

		y += 1
		if !scanner.Scan() {
			if scanner.Err() != nil {
				log.Fatal(scanner.Err())
			}
			break
		}
		currentLine = scanner.Text()
	}

	if startingNode == nil {
		log.Fatal("did not find starting node")
	}

	return startingNode, gridWidth, y
}

type Dir int

const (
	North Dir = iota
	East
	South
	West
)

func dirToName(d Dir) string {
	name := "West"
	switch d {
	case North:
		name = "North"
	case East:
		name = "East"
	case South:
		name = "South"
	}

	return name
}
func (d Dir) oppositeDir() Dir {
	return Dir((d + 2) % 4)
}

func writePathToGrid(startingNode *Node, grid *Grid) {
	traveler := makeTravelFunc(startingNode)

	grid.setPoint(startingNode.location, byte(startingNode.originalMarker))

	for {
		n := traveler()

		if n.originalMarker == 'S' {
			break
		}

		grid.setPoint(n.location, byte(n.originalMarker))

	}

}

func makeTravelFunc(start *Node) func() *Node {
	lastDirTraveled := North
	currentNode := start

	return func() *Node {
		directions := []*Node{
			currentNode.north,
			currentNode.east,
			currentNode.south,
			currentNode.west,
		}

		entryPoint := lastDirTraveled.oppositeDir()
		for i, n := range directions {
			if Dir(i) != entryPoint && n != nil {
				lastDirTraveled = Dir(i)
				currentNode = n
				return currentNode
			}
		}

		log.Fatal("Failed to find valid direction (incomplete circle?)")
		return nil
	}
}

func findFarthestDistance(startingNode *Node) int {
	// fmt.Printf("starting node %+v\n", startingNode)

	// logNode := func(n Node) {
	// 	fmt.Printf("\tA: %s %+v\n", string(n.originalMarker), n.location)
	//
	// 	fmt.Printf("\t\t")
	// 	if n.north != nil {
	// 		fmt.Printf("N %s ", string(n.north.originalMarker))
	// 	} else {
	// 		fmt.Print("N x ")
	// 	}
	// 	if n.east != nil {
	// 		fmt.Printf("E %s ", string(n.east.originalMarker))
	// 	} else {
	// 		fmt.Print("E x ")
	// 	}
	// 	if n.south != nil {
	// 		fmt.Printf("S %s ", string(n.south.originalMarker))
	// 	} else {
	// 		fmt.Print("S x ")
	// 	}
	// 	if n.west != nil {
	// 		fmt.Printf("W %s ", string(n.west.originalMarker))
	// 	} else {
	// 		fmt.Print("W x ")
	// 	}
	// 	fmt.Print("\n")
	// }

	var travelerA *Node = nil
	var travelerB *Node = nil

	startingDirections := []*Node{
		startingNode.north,
		startingNode.east,
		startingNode.south,
		startingNode.west,
	}

	// fmt.Printf("startingDirs %+v\n", startingDirections)

	var lastDirTraveledA Dir
	var lastDirTraveledB Dir

	// fmt.Println("0")
	// logNode(*startingNode)
	for i, n := range startingDirections {
		if n != nil {
			if travelerA == nil {
				travelerA = n
				lastDirTraveledA = Dir(i)
				// fmt.Printf("\t\tA taking %s\n", dirToName(lastDirTraveledA))
			} else {
				travelerB = n
				lastDirTraveledB = Dir(i)
				// fmt.Printf("\t\tB taking %s\n", dirToName(lastDirTraveledB))
			}
		}
		if travelerA != nil && travelerB != nil {
			break
		}
	}
	// fmt.Printf("A %+v | B %+v\n", travelerA, travelerB)

	travel := func(start *Node, lastDirTraveled Dir) (*Node, Dir) {
		directions := []*Node{
			start.north,
			start.east,
			start.south,
			start.west,
		}

		entryPoint := lastDirTraveled.oppositeDir()
		for i, n := range directions {
			if Dir(i) != entryPoint && n != nil {
				return n, Dir(i)
			}
		}

		log.Fatal("Failed to find valid direction (incomplete circle?)")
		return nil, 0
	}

	count := 1
	for {
		// fmt.Println(count)

		// logNode(*travelerA)
		travelerA, lastDirTraveledA = travel(travelerA, lastDirTraveledA)
		// fmt.Printf("\t\tA taking %s\n", dirToName(lastDirTraveledA))

		/////

		// logNode(*travelerB)
		travelerB, lastDirTraveledB = travel(travelerB, lastDirTraveledB)
		// fmt.Printf("\t\tB taking %s\n", dirToName(lastDirTraveledB))

		/////

		count += 1
		if travelerA == travelerB {
			break
		}

	}
	// fmt.Println(count)
	// fmt.Printf("\tA: %s %+v\n", string(travelerA.originalMarker), travelerA.location)
	// fmt.Printf("\t\tN %v E %v S %v W %v\n",
	// 	travelerA.north != nil, travelerA.east != nil, travelerA.south != nil, travelerA.west != nil)
	//
	// fmt.Printf("\tB: %s %+v\n", string(travelerB.originalMarker), travelerB.location)
	// fmt.Printf("\t\tN %v E %v S %v W %v\n",
	// 	travelerB.north != nil, travelerB.east != nil, travelerB.south != nil, travelerB.west != nil)

	return count
}
