package drbreakboard

import (
	"fmt"
	"strings"
)

func DrawBoard(field *PlayField) {
	for y := 0; y < field.GetHeight(); y++ {
		for x := 0; x < field.GetWidth(); x++ {
			space, _ := field.GetSpaceAtCoordinate(y, x)
			fmt.Print(generateRawSpaceString(space))
			fmt.Print(" ")
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func DrawNextIteration(field [][]NextIteration) {
	for _, row := range field {
		for _, space := range row {
			switch space {
			case NoAction:
				fmt.Print("X")
			case Clear:
				fmt.Print("C")
			case Fall:
				fmt.Print("F")
			}
		}
		fmt.Println("")
	}
	fmt.Println("")
}

func generateRawSpaceString(space Space) string {
	var sb strings.Builder

	switch space.Content {
	case Empty:
		sb.WriteString("X")
	case Virus:
		sb.WriteString("V")
	case Pill:
		sb.WriteString("P")
	}

	switch space.Color {
	case Uncolored:
		sb.WriteString("X")
	case Red:
		sb.WriteString("R")
	case Blue:
		sb.WriteString("B")
	case Yellow:
		sb.WriteString("Y")
	}

	switch space.Linkage {
	case Unlinked:
		sb.WriteString("X")
	case Up:
		sb.WriteString("U")
	case Down:
		sb.WriteString("D")
	case Left:
		sb.WriteString("L")
	case Right:
		sb.WriteString("R")
	}

	return sb.String()
}
