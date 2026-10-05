package ui

// Split divides total cells into two parts: the first gets ratio of the
// total, but neither part goes below its minimum while the total allows.
// When the total cannot hold both minimums, the first part gets everything
// and the second is 0, so screens collapse to one pane on small terminals.
func Split(total int, ratio float64, minA, minB int) (a, b int) {
	if total <= 0 {
		return 0, 0
	}
	if total < minA+minB {
		return total, 0
	}
	a = int(float64(total) * ratio)
	if a < minA {
		a = minA
	}
	if total-a < minB {
		a = total - minB
	}
	return a, total - a
}

// Stack divides a height into a fixed header, a fixed footer and the body
// between them, never negative.
func Stack(total, header, footer int) (body int) {
	return max(0, total-header-footer)
}
