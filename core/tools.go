package core

func appendInt(s *[]int, v int) {
	*s = append(*s, v)
}

func removeInt(s *[]int, i int) {
	copy((*s)[i:], (*s)[i+1:])
	*s = (*s)[:len(*s)-1]
}
