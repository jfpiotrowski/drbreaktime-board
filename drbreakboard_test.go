package drbreakboard

import (
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestBoardCreation(t *testing.T) {
	board := NewPlayField(8, 16)

	topleft, ok := board.GetSpaceAtCoordinate(0, 0)

	if ok != nil {
		t.Fatalf("Unexpected top left space error %v", ok)
	}

	if topleft.Content != Empty {
		t.Fatalf("top left space in new board was not empty, %v", topleft.Content)
	}

	_, ok = board.GetSpaceAtCoordinate(50, 0)

	if ok == nil {
		t.Fatal("out of bounds y space retrieval worked")
	}

	_, ok = board.GetSpaceAtCoordinate(0, 50)

	if ok == nil {
		t.Fatal("out of bounds x space retrieval worked")
	}
}

func TestClearChecking(t *testing.T) {
	field := NewPlayField(8, 16)

	DrawBoard(field)

	bottomRow := field.GetBottomRowIndex()

	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 0, Space{Pill, Unlinked, Blue})
	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 1, Space{Pill, Unlinked, Blue})
	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 2, Space{Pill, Unlinked, Blue})
	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 3, Space{Pill, Unlinked, Blue})

	DrawBoard(field)

	iterField, nextIter := field.EvaluateBoardIteration()

	DrawNextIteration(iterField)

	if nextIter != Clear {
		t.Fatal("clear iteration returned no changes")
	}

	if iterField[bottomRow][0] != Clear {
		t.Fatal("Cleared piece not showing clear")
	}

	field.PutSpaceAtCoordinateIfEmpty(bottomRow-1, 0, Space{Pill, Unlinked, Blue})
	field.PutSpaceAtCoordinateIfEmpty(bottomRow-2, 0, Space{Pill, Unlinked, Blue})
	field.PutSpaceAtCoordinateIfEmpty(bottomRow-3, 0, Space{Pill, Unlinked, Blue})

	iterField, nextIter = field.EvaluateBoardIteration()

	DrawNextIteration(iterField)

	if nextIter != NoAction {
		t.Fatal("clear iteration returned no changes")
	}

	if iterField[bottomRow-1][0] != Clear {
		t.Fatal("Cleared piece not showing clear")
	}
}

func TestSinglePillStackFallChecking(t *testing.T) {
	field := NewPlayField(8, 16)

	DrawBoard(field)

	bottomRow := field.GetBottomRowIndex()

	field.PutSpaceAtCoordinateIfEmpty(bottomRow-1, 0, Space{Pill, Unlinked, Blue})

	DrawBoard(field)

	iterField, nextIter := field.EvaluateBoardIteration()

	DrawNextIteration(iterField)

	if nextIter != Fall {
		t.Fatal("fall iteration returned no changes")
	}

	if iterField[bottomRow-1][0] != Fall {
		t.Fatal("falling piece not showing Fall")
	}

	field.PutSpaceAtCoordinateIfEmpty(bottomRow-2, 0, Space{Pill, Unlinked, Blue})

	DrawBoard(field)

	iterField, nextIter = field.EvaluateBoardIteration()

	DrawNextIteration(iterField)

	if nextIter != Fall {
		t.Fatal("fall iteration returned no changes")
	}

	if iterField[bottomRow-2][0] != Fall {
		t.Fatal("falling piece not showing Fall")
	}

	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 0, Space{Pill, Unlinked, Blue})

	DrawBoard(field)

	iterField, nextIter = field.EvaluateBoardIteration()

	DrawNextIteration(iterField)

	if nextIter != Fall {
		t.Fatal("fall iteration returned no changes")
	}
}

func TestLinkedSpaceIteration(t *testing.T) {
	field := NewPlayField(8, 16)

	DrawBoard(field)

	bottomRow := field.GetBottomRowIndex()

	// put linked spaces one row over bottom
	space, linkedSpace, err := MakeLinkedPillSpaces(Up, Red, Blue)
	if err != nil {
		t.Fatalf("make linked pill err: %v", err)
	}

	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow-1, 0, space, linkedSpace)
	if err != nil {
		t.Fatalf("put linked pill err: %v", err)
	}
	DrawBoard(field)
	iterField, nextIter := field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Fall {
		t.Fatal("fall iteration returned no changes")
	}

	// test horizontal linking
	space, linkedSpace, err = MakeLinkedPillSpaces(Right, Red, Blue)
	if err != nil {
		t.Fatalf("make linked pill err: %v", err)
	}
	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow-3, 0, space, linkedSpace)
	if err != nil {
		t.Fatalf("put linked pill err: %v", err)
	}
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Fall {
		t.Fatal("fall iteration returned no changes")
	}

	// put piece in bottom row, all pieces should be docked
	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow, 0, space, linkedSpace)
	if err != nil {
		t.Fatalf("put linked pill err: %v", err)
	}
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != NoAction {
		t.Fatal("fall iteration should have no changes")
	}

	// put blue col of 4 in col 2
	space, linkedSpace, err = MakeLinkedPillSpaces(Up, Blue, Blue)
	if err != nil {
		t.Fatalf("could not create piece %v", err)
	}
	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow-1, 1, space, linkedSpace)
	if err != nil {
		t.Fatalf("could not place piece %v", err)
	}
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Clear {
		t.Fatal("col should be detected")
	}

	if iterField[bottomRow][1] != Clear {
		t.Fatal("Piece should have been cleared")
	}
}

func TestVirusSpaceIteration(t *testing.T) {
	field := NewPlayField(8, 16)

	virus, _ := MakeVirus(Blue)

	bottomRow := field.GetBottomRowIndex()

	// place viruses
	field.PutSpaceAtCoordinateIfEmpty(0, 3, virus)
	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 3, virus)

	// place horizontal pill over virus
	space, linkedSpace, err := MakeLinkedPillSpaces(Left, Blue, Blue)
	if err != nil {
		t.Fatalf("could not create piece %v", err)
	}
	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow-1, 3, space, linkedSpace)
	if err != nil {
		t.Fatalf("could not place piece %v", err)
	}

	// make sure there are no changes
	DrawBoard(field)
	iterField, nextIter := field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != NoAction {
		t.Fatal("no changes should have been found")
	}

	// test virus clear
	space, linkedSpace, _ = MakeLinkedPillSpaces(Up, Blue, Blue)
	err = field.PutTwoLinkedSpacesAtCoordinate(bottomRow-2, 3, space, linkedSpace)
	if err != nil {
		t.Fatalf("could not place linked piece %v", err)
	}
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Clear {
		t.Fatal("clear should have caused change")
	}
	if iterField[bottomRow][3] != Clear {
		t.Fatal("Piece should have been cleared")
	}
}

func TestIterationExecution(t *testing.T) {
	field := NewPlayField(8, 16)

	space, linkedSpace, _ := MakeLinkedPillSpaces(Right, Blue, Blue)
	virus, _ := MakeVirus(Blue)

	bottomRow := field.GetBottomRowIndex()

	field.PutSpaceAtCoordinateIfEmpty(bottomRow, 3, virus)
	field.PutSpaceAtCoordinateIfEmpty(bottomRow-1, 3, virus)
	field.PutTwoLinkedSpacesAtCoordinate(bottomRow-2, 3, space, linkedSpace)
	field.PutTwoLinkedSpacesAtCoordinate(bottomRow-3, 3, space, linkedSpace)

	DrawBoard(field)
	iterField, nextIter := field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Clear {
		t.Fatal("clear should be next iteration")
	}
	err := field.IterateBoard()
	if err != nil {
		t.Fatalf("board iterate errored, %v", err)
	}

	// iterate should leave two uncleared stacked blocks to fall 2 spots
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Fall {
		t.Fatal("fall should be next iteration")
	}
	err = field.IterateBoard()
	if err != nil {
		t.Fatalf("board iterate errored, %v", err)
	}

	// run second fall
	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != Fall {
		t.Fatal("fall should be next iteration")
	}
	err = field.IterateBoard()
	if err != nil {
		t.Fatalf("board iterate errored, %v", err)
	}

	DrawBoard(field)
	iterField, nextIter = field.EvaluateBoardIteration()
	DrawNextIteration(iterField)
	if nextIter != NoAction {
		t.Fatal("no action should be next iteration")
	}
}
