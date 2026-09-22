package service

// computePosition returns a position for inserting an item at index within
// existing (already sorted ascending, and excluding the item being placed).
// New items get the midpoint of their neighbors, so reordering never
// requires rewriting sibling rows (the same trick Trello uses).
func computePosition(existing []float64, index int) float64 {
	if len(existing) == 0 {
		return 1
	}
	if index <= 0 {
		return existing[0] - 1
	}
	if index >= len(existing) {
		return existing[len(existing)-1] + 1
	}
	return (existing[index-1] + existing[index]) / 2
}
