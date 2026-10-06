package ui

// TreePrefixes returns the tree lines for rows given in depth-first order
// with their depths (0 for roots): "├─ " or "└─ " at the row's own level and
// "│  " or blank for every open ancestor level.
func TreePrefixes(depths []int) []string {
	out := make([]string, len(depths))
	// hasNext reports whether a later row continues level l before the
	// subtree at that level closes.
	hasNext := func(k, l int) bool {
		for j := k + 1; j < len(depths); j++ {
			switch {
			case depths[j] == l:
				return true
			case depths[j] < l:
				return false
			}
		}
		return false
	}
	for k, d := range depths {
		p := ""
		for l := 1; l < d; l++ {
			if hasNext(k, l) {
				p += "│  "
			} else {
				p += "   "
			}
		}
		if d > 0 {
			if hasNext(k, d) {
				p += "├─ "
			} else {
				p += "└─ "
			}
		}
		out[k] = p
	}
	return out
}
