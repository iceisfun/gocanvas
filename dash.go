package gocanvas

import "math"

// applyDash breaks sub-paths into dashed segments according to the given pattern.
// Even indices in pattern are draw lengths, odd indices are gap lengths.
// The offset shifts the starting point along the pattern.
func applyDash(subPaths [][]Point, pattern []float64, offset float64) [][]Point {
	if len(pattern) == 0 {
		return subPaths
	}

	// Normalize: if odd length, double it (per HTML5 Canvas spec).
	if len(pattern)%2 != 0 {
		pattern = append(pattern, pattern...)
	}

	// Compute total pattern length.
	var patternLen float64
	for _, v := range pattern {
		patternLen += v
	}
	if patternLen <= 0 {
		return subPaths
	}

	// Normalize offset into [0, patternLen).
	offset = math.Mod(offset, patternLen)
	if offset < 0 {
		offset += patternLen
	}

	// Find starting position in the pattern.
	dashIdx := 0
	dashRemaining := pattern[0]
	for offset > 0 {
		if offset < dashRemaining {
			dashRemaining -= offset
			break
		}
		offset -= dashRemaining
		dashIdx = (dashIdx + 1) % len(pattern)
		dashRemaining = pattern[dashIdx]
	}

	var result [][]Point

	for _, sp := range subPaths {
		if len(sp) < 2 {
			continue
		}

		// Reset dash state per sub-path to maintain continuity within
		// a single sub-path, but the caller decides if state carries over.
		di := dashIdx
		dr := dashRemaining
		var current []Point
		drawing := di%2 == 0

		for i := 0; i < len(sp)-1; i++ {
			p0 := sp[i]
			p1 := sp[i+1]

			dx := p1.X - p0.X
			dy := p1.Y - p0.Y
			segLen := math.Sqrt(dx*dx + dy*dy)
			if segLen < 1e-10 {
				continue
			}

			consumed := 0.0

			for consumed < segLen {
				remaining := segLen - consumed
				take := dr
				if take > remaining {
					take = remaining
				}

				t := (consumed + take) / segLen
				pt := Point{
					X: p0.X + dx*t,
					Y: p0.Y + dy*t,
				}

				if drawing {
					if len(current) == 0 {
						// Start a new dash segment.
						t0 := consumed / segLen
						start := Point{
							X: p0.X + dx*t0,
							Y: p0.Y + dy*t0,
						}
						current = append(current, start)
					}
					current = append(current, pt)
				} else {
					// Ending a gap — flush any current dash.
					if len(current) >= 2 {
						result = append(result, current)
						current = nil
					} else {
						current = nil
					}
				}

				consumed += take
				dr -= take

				if dr <= 1e-10 {
					// Advance to next dash element.
					if drawing && len(current) >= 2 {
						result = append(result, current)
						current = nil
					}
					di = (di + 1) % len(pattern)
					dr = pattern[di]
					drawing = di%2 == 0
				}
			}
		}

		// Flush remaining dash segment.
		if len(current) >= 2 {
			result = append(result, current)
		}
	}

	return result
}
