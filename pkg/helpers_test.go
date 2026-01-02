package pkg

// countScaled counts how many ScaleInfo items have Scaled = true
func countScaled(infos []ScaleInfo) int {
	count := 0
	for _, info := range infos {
		if info.Scaled {
			count++
		}
	}
	return count
}
