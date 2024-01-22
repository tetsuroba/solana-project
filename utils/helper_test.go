package utils

import "testing"

func TestHelperFunctions(t *testing.T) {
	t.Run("Finds the correct item from int array", func(t *testing.T) {
		arr := []int{0, 1, 2, 3, 4, 5}
		index := Find(arr, 3)
		if index != 3 {
			t.Errorf("Incorrect index returned %d should be %d", index, 3)
		}
	})
	t.Run("Finds the correct item from string array", func(t *testing.T) {
		arr := []string{"0", "1", "2", "3", "4", "5"}
		index := Find(arr, "3")
		if index != 3 {
			t.Errorf("Incorrect index returned %d should be %d", index, 3)
		}
	})
}
