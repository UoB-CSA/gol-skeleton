package util

import (
	"fmt"
	"log"
	"strings"
)

func VisualiseMatrix(given [][]uint8, width, height int) {
	log.Print(matricesToString(given, nil, width, height))
}

func (c1 Cell) in(slice []Cell) bool {
	for _, c2 := range slice {
		if c1 == c2 {
			return true
		}
	}
	return false
}

func AliveCellsToString(given, expected []Cell, width, height int) string {
	givenMatrix := make([][]byte, height)
	for i := range givenMatrix {
		givenMatrix[i] = make([]byte, width)
	}
	expectedMatrix := make([][]byte, height)
	for i := range expectedMatrix {
		expectedMatrix[i] = make([]byte, width)
	}
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if (Cell{j, i}).in(given) {
				givenMatrix[i][j] = 0xFF
			}
		}
	}
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if (Cell{j, i}).in(expected) {
				expectedMatrix[i][j] = 0xFF
			}
		}
	}
	var output []string
	output = append(output, "  Your alive cells:                      Expected alive cells:\n")
	output = append(output, squaresToStrings(givenMatrix, expectedMatrix, width, height)...)
	return strings.Join(output, "")
}

func matricesToString(given, expected [][]uint8, width, height int) string {
	var output []string
	output = append(output, "  Your world matrix:                     ")
	if expected != nil {
		output = append(output, "Expected world matrix:\n")
	} else {
		output = append(output, "\n")
	}
	output = append(output, squaresToStrings(given, expected, width, height)...)
	return strings.Join(output, "")
}

func getHorizontalBorder(start, middle, end string, width int) string {
	border := start
	for i := 0; i < width*2; i++ {
		border += middle
	}
	border += end
	return border
}

func squaresToStrings(given, expected [][]uint8, width, height int) []string {
	var output []string
	output = append(output, getHorizontalBorder("  ┌", "─", "┐ ", width))
	if expected != nil {
		output = append(output, getHorizontalBorder("    ┌", "─", "┐", width))
	}
	output = append(output, "\n")

	for i := 0; i < height; i++ {
		output = append(output, fmt.Sprintf("%2d│", i))
		for j := 0; j < width; j++ {
			if given[i][j] == 0xFF {
				output = append(output, "██")
			} else if given[i][j] == 0x00 {
				output = append(output, "  ")
			}
		}

		if expected != nil {
			output = append(output, fmt.Sprintf("│   %2d│", i))
			for j := 0; j < width; j++ {
				if expected[i][j] == 0xFF {
					output = append(output, "██")
				} else if expected[i][j] == 0x00 {
					output = append(output, "  ")
				}
			}
		}
		output = append(output, "│\n")
	}
	output = append(output, getHorizontalBorder("  └", "─", "┘ ", width))
	if expected != nil {
		output = append(output, getHorizontalBorder("    └", "─", "┘", width))
	}
	output = append(output, "\n")

	return output
}

func Green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func Yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func Red(text string) string {
	return "\033[31m" + text + "\033[0m"
}
