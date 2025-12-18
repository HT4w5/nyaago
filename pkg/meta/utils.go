package meta

func getWidth(strs []string) int {
	width := 0
	for _, v := range strs {
		l := len(v)
		if l > width {
			width = l
		}
	}
	return width
}
