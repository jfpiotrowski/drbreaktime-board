package drbreakboard

import (
	"errors"

	"github.com/rs/zerolog/log"
)

type SpaceContent int
type SpaceLinkage int
type SpaceColor int
type NextIteration int

const (
	Empty SpaceContent = iota
	Virus
	Pill
)

const (
	Unlinked SpaceLinkage = iota
	Up
	Down
	Left
	Right
)

const (
	Uncolored SpaceColor = iota
	Red
	Blue
	Yellow
)

const (
	NoAction NextIteration = iota
	Clear
	Fall
)

type Space struct {
	Content SpaceContent
	Linkage SpaceLinkage
	Color   SpaceColor
}

type Coordinate struct {
	y int
	x int
}

// Playfield for drbreaktime game
// can be arbitrarily sized
type PlayField struct {
	spaces [][]Space
}

// return an empty playfield
func NewPlayField(x int, y int) *PlayField {
	// initialize playfield to empty
	field := new(PlayField)
	field.spaces = make([][]Space, y)
	for i := range field.spaces {
		field.spaces[i] = make([]Space, x)
	}

	return field
}

func (field *PlayField) GetHeight() int {
	return len(field.spaces)
}

func (field *PlayField) GetWidth() int {
	return len(field.spaces[0])
}

// get the space at a given coordinate
func (field *PlayField) GetSpaceAtCoordinate(y int, x int) (Space, error) {
	// check coordinate in bounds
	if err := field.checkCoordinateInBounds(y, x); err != nil {
		return Space{}, err
	}

	return field.spaces[y][x], nil
}

// put space content in the play field if it leaves a legal board
// a legal board has no linked pills without a matching link adjacent.
// This means no putting pieces with a linkage, use PutLinkedSpaces for that
func (field *PlayField) PutSpaceAtCoordinateIfEmpty(y int, x int, space Space) error {
	if space.Linkage != Unlinked {
		return errors.New("cannot put a single linked space")
	}

	err := field.checkCoordinateInBoundsAndEmpty(y, x)

	if err != nil {
		return err
	}

	return field.putSpaceAtCoordinate(y, x, space)
}

// put space content in the play field if it leaves a legal board
// a legal board has no linked pills without a matching link adjacent.
// This means no putting pieces with a linkage, use PutLinkedSpaces for that
func (field *PlayField) PutTwoLinkedSpacesAtCoordinate(y int, x int, coordSpace Space, linkedSpace Space) error {
	// verify linkage is not unlinked and that spaces are properly linked
	if coordSpace.Linkage == Unlinked {
		return errors.New("cannot put an unlinked space")
	}

	if coordSpace.Content != Pill || linkedSpace.Content != Pill {
		return errors.New("space input was not pill as required")
	}

	switch coordSpace.Linkage {
	case Up:
		if linkedSpace.Linkage != Down {
			return errors.New("piece linkage was invalid")
		}
	case Down:
		if linkedSpace.Linkage != Up {
			return errors.New("piece linkage was invalid")
		}
	case Left:
		if linkedSpace.Linkage != Right {
			return errors.New("piece linkage was invalid")
		}
	case Right:
		if linkedSpace.Linkage != Left {
			return errors.New("piece linkage was invalid")
		}
	}

	// check coords are in bounds and empty
	err := field.checkCoordinateInBoundsAndEmpty(y, x)
	if err != nil {
		return err
	}

	// get linked coord and ensure it's in bounds and empty
	linkedY, linkedX, err := GetLinkedCoordinate(y, x, coordSpace.Linkage)
	if err != nil {
		return err
	}

	err = field.checkCoordinateInBoundsAndEmpty(linkedY, linkedX)
	if err != nil {
		return err
	}

	// pieces are a pill with correct linkage in each constituent space
	// put the pieces into the field
	err = field.putSpaceAtCoordinate(y, x, coordSpace)

	if err != nil {
		return err
	}

	return field.putSpaceAtCoordinate(linkedY, linkedX, linkedSpace)
}

// Clear the board
func (field *PlayField) ClearBoard() {
	for _, row := range field.spaces {
		for x := range row {
			row[x] = Space{}
		}
	}
}

func (field *PlayField) GetBottomRowIndex() int {
	return len(field.spaces) - 1
}

func (field *PlayField) EvaluateBoardIteration() ([][]NextIteration, NextIteration) {
	// initialize board iteration field to empty
	nextIterationField := make([][]NextIteration, len(field.spaces))
	for i := range nextIterationField {
		nextIterationField[i] = make([]NextIteration, len(field.spaces[i]))
	}

	dockedField := field.generateDockedField()

	undockedPieceFound := false

	// look for falling pieces
	for y := range dockedField {
		for x, docked := range dockedField[y] {
			if field.spaces[y][x].Content == Pill && !docked {
				// undocked pill found
				// mark undocked piece found and mark it as fall in the
				// next iteration field
				undockedPieceFound = true
				nextIterationField[y][x] = Fall
			}
		}
	}

	// if we found a falling piece, we're done
	// return the field and that the board has movement
	if undockedPieceFound {
		return nextIterationField, Fall
	}

	// no falling pieces, check for clears
	// first horizontal

	currentStreak := 0
	currentColor := Uncolored
	nextIter := NoAction

	// look for rows with 4 or more consecutive color matches
	for y := range field.spaces {
		x := 0
		for {
			if field.checkCoordinateInBounds(y, x) != nil {
				// out of bounds
				if currentStreak >= 4 && currentColor != Uncolored {
					// have match stored, mark it
					nextIter = Clear
					for i := x - currentStreak; i < x; i++ {
						nextIterationField[y][i] = Clear
					}
				}

				// clear the vars for new row
				currentStreak = 0
				currentColor = Uncolored

				// done with this row, break out of forever loop
				break
			} else {
				// still in bounds, check next piece in row
				if field.spaces[y][x].Color == currentColor {
					// same as previous, continue the streak
					currentStreak += 1
				} else {
					// next piece is different
					// check if we have a row of 4+
					if currentStreak >= 4 && currentColor != Uncolored {
						// have match stored, mark it
						nextIter = Clear
						for i := x - currentStreak; i < x; i++ {
							nextIterationField[y][i] = Clear
						}
					}

					// set the vars for the new streak
					currentStreak = 1
					currentColor = field.spaces[y][x].Color
				}
			}

			// move x index to next space
			x += 1
		}
	}

	// look for cols with 4 or more consecutive color matches
	for x := range field.spaces[0] {
		y := 0

		// clear the vars for new row
		currentStreak = 0
		currentColor = Uncolored

		for {
			if field.checkCoordinateInBounds(y, x) != nil {
				// out of bounds
				if currentStreak >= 4 && currentColor != Uncolored {
					// have match stored, mark it
					nextIter = Clear
					for i := y - currentStreak; i < y; i++ {
						nextIterationField[i][x] = Clear
					}
				}
				// done with this row, break out of forever loop
				break
			} else {
				// still in bounds, check next piece in row
				if field.spaces[y][x].Color == currentColor {
					// same as previous, continue the streak
					currentStreak += 1
				} else {
					// next piece is different
					// check if we have a row of 4+
					if currentStreak >= 4 && currentColor != Uncolored {
						// have match stored, mark it
						nextIter = Clear
						for i := y - currentStreak; i < y; i++ {
							nextIterationField[i][x] = Clear
						}
					}

					// set the vars for the new streak
					currentStreak = 1
					currentColor = field.spaces[y][x].Color
				}
			}

			// move y index to next space
			y += 1
		}
	}

	return nextIterationField, nextIter
}

// iterate changes through the board
// error means something is semantically wrong with the board
// and should cause a panic level reaction
func (field *PlayField) IterateBoard() error {
	log.Trace().Msg("Entering IterateBoard()")
	iterField, nextIter := field.EvaluateBoardIteration()

	// no changes means nothing to iterate
	if nextIter == NoAction {
		log.Debug().Msg("no action needed for iterate")
		return nil
	}

	// clear pieces if nextIter is Clear
	if nextIter == Clear {
		log.Debug().Msg("Clear is next iteration")
		// next update is removing cleared pieces from the board
		// remove pieces and make linked spaces singles
		for y, row := range field.spaces {
			for x := range row {
				if iterField[y][x] == Clear {
					err := field.clearSpace(y, x)
					if err != nil {
						// something has gone horribly wrong
						log.Error().Msg("clearSpace failed when clearing")
						return err
					}
				}
			}
		}

		return nil
	}

	log.Debug().Msg("Fall is next iteration")
	// nextIter is Fall, so drop pieces
	// work bottom to top to not overwrite pieces
	for y := field.GetBottomRowIndex() - 1; y >= 0; y-- {
		for x, space := range field.spaces[y] {
			if iterField[y][x] == Fall {
				err := field.putSpaceAtCoordinate(y+1, x, space)
				if err != nil {
					// something has gone horribly wrong
					return err
				}
				// replace existing space with empty space
				field.putSpaceAtCoordinate(y, x, Space{})
			}
		}
	}

	return nil
}

func (field *PlayField) GetVirusCount() int {
	viruses := 0
	for _, row := range field.spaces {
		for _, space := range row {
			if space.Content == Virus {
				viruses += 1
			}
		}
	}
	return viruses
}

// maker function for linked pill spaces
func MakeLinkedPillSpaces(linkage SpaceLinkage, coordColor SpaceColor,
	linkedColor SpaceColor) (Space, Space, error) {
	if linkage == Unlinked {
		return Space{}, Space{}, errors.New("linked pill cannot have unlinked linkage")
	}

	if coordColor == Uncolored || linkedColor == Uncolored {
		return Space{}, Space{}, errors.New("spaces must have a color")
	}

	return Space{Pill, linkage, coordColor}, Space{Pill, getOpposingLinkage(linkage), linkedColor}, nil
}

// maker function for virus
func MakeVirus(color SpaceColor) (Space, error) {
	if color == Uncolored {
		return Space{}, errors.New("virus must have color")
	}

	return Space{Virus, Unlinked, color}, nil
}

func getOpposingLinkage(linkage SpaceLinkage) SpaceLinkage {
	switch linkage {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	}

	return Unlinked
}

// get the linked coordinate
// makes no guarantee for coordinate being in bounds
func GetLinkedCoordinate(y int, x int, linkage SpaceLinkage) (int, int, error) {
	switch linkage {
	case Up:
		return y - 1, x, nil
	case Down:
		return y + 1, x, nil
	case Left:
		return y, x - 1, nil
	case Right:
		return y, x + 1, nil
	default:
		return 0, 0, errors.New("could not get linked coordinate for unlinked space")
	}
}

// raw put into the field with no board integrity check
func (field *PlayField) putSpaceAtCoordinate(y int, x int, space Space) error {

	// check coordinate in bounds
	if err := field.checkCoordinateInBounds(y, x); err != nil {
		return err
	}

	field.spaces[y][x] = space
	return nil
}

func (field *PlayField) checkCoordinateInBounds(y int, x int) error {
	if y < 0 || y >= len(field.spaces) {
		return errors.New("y was out of bounds")
	}

	row := field.spaces[y]

	if x < 0 || x >= len(row) {
		return errors.New("x was out of bounds")
	}

	// no error, return nil
	return nil
}

func (field *PlayField) checkCoordinateInBoundsAndEmpty(y int, x int) error {
	// check coordinate in bounds
	if err := field.checkCoordinateInBounds(y, x); err != nil {
		return err
	}

	if field.spaces[y][x].Content != Empty {
		return errors.New("space was not empty")
	}

	return nil
}

// clear a space and unlink any linked space
func (field *PlayField) clearSpace(y int, x int) error {
	// check coordinate in bounds
	space, err := field.GetSpaceAtCoordinate(y, x)
	if err != nil {
		return err
	}

	if space.Linkage == Unlinked {
		// no linkage, just clear the space
		field.putSpaceAtCoordinate(y, x, Space{})
		return nil
	}

	// space is linked, have to set the linked space to single
	linkedY, linkedX, err := GetLinkedCoordinate(y, x, space.Linkage)
	if err != nil {
		return err
	}

	linkedSpace, err := field.GetSpaceAtCoordinate(linkedY, linkedX)
	if err != nil {
		return err
	}

	// set linked space to unlinked and put linked space and empty space into board
	linkedSpace.Linkage = Unlinked
	field.putSpaceAtCoordinate(linkedY, linkedX, linkedSpace)
	field.putSpaceAtCoordinate(y, x, Space{})

	return nil
}

func (field *PlayField) generateDockedField() [][]bool {
	// initialize docked field to empty
	dockedField := make([][]bool, len(field.spaces))
	for i := range dockedField {
		dockedField[i] = make([]bool, len(field.spaces[i]))
	}

	// create a queue for pieces to evaluate for dockedness
	dockedCheckQueue := make([]Coordinate, 0)

	// seed queue with virii and pieces on bottom
	for y := range dockedField {
		for x := range dockedField[y] {
			// iterate through all pieces looking for virii or bottom row
			space, _ := field.GetSpaceAtCoordinate(y, x)
			if space.Content == Virus {
				// all virii are docked
				dockedCheckQueue = append(dockedCheckQueue, Coordinate{y, x})
			} else if y == field.GetBottomRowIndex() && space.Content != Empty {
				// piece resting on bottom
				dockedCheckQueue = append(dockedCheckQueue, Coordinate{y, x})
			}
		}
	}

	for len(dockedCheckQueue) > 0 {
		// get the space and remove it from the queue
		dockedSpace := dockedCheckQueue[0]
		dockedCheckQueue = dockedCheckQueue[1:]

		// check coordinate in bounds
		if field.checkCoordinateInBounds(dockedSpace.y, dockedSpace.x) != nil {
			continue
		}

		// if already docked, ignore
		if dockedField[dockedSpace.y][dockedSpace.x] {
			continue
		}

		// if this space is empty, ignore it
		if field.spaces[dockedSpace.y][dockedSpace.x].Content == Empty {
			continue
		}

		// there is a piece here that is docked, mark it as docked
		dockedField[dockedSpace.y][dockedSpace.x] = true

		// if a piece is docked, the space above it and
		// a space linked with this space is also docked
		if field.checkCoordinateInBounds(dockedSpace.y-1, dockedSpace.x) == nil {
			// space is in bounds, add it to the queue
			dockedCheckQueue = append(dockedCheckQueue, Coordinate{dockedSpace.y - 1, dockedSpace.x})
		}

		if field.spaces[dockedSpace.y][dockedSpace.x].Linkage != Unlinked {
			// this space is linked, mark its link partner as docked
			linkedY, linkedX, _ := GetLinkedCoordinate(dockedSpace.y, dockedSpace.x, field.spaces[dockedSpace.y][dockedSpace.x].Linkage)
			dockedCheckQueue = append(dockedCheckQueue, Coordinate{linkedY, linkedX})
		}
	}

	return dockedField
}
