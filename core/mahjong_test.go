package core

import "testing"

func TestCheckNormalWin(t *testing.T) {
	a := []struct {
		handTile []int
		m        int
		q        int
		ans      bool
	}{{[]int{0, 1, 2, 4, 5, 6, 8, 9, 10, 12, 13, 14, 16, 17}, 4, 1, true},
		{[]int{0, 1, 2, 4, 8, 12, 16, 20, 21, 24, 28, 32, 33, 34}, 4, 1, true},
		{[]int{0, 1, 2, 4, 8, 12, 16, 20, 24, 28, 32, 33, 34, 40}, 4, 1, false}}
	for i := range a {
		if checkNormalWin(a[i].handTile, a[i].m, a[i].q) != a[i].ans {
			t.Errorf("checkNormalWin(%v,%d,%d) = %v", a[i].handTile, a[i].m, a[i].q, !a[i].ans)
		} else {
			t.Logf("checkNormalWin(%v,%d,%d) passed", a[i].handTile, a[i].m, a[i].q)
		}
	}
}
