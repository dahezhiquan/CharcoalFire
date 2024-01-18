package utils

// 算法

// BinarySearch 二分查找，判断一个字符串是否在一个字符串切片中存在（前提是这个字符串切片是排序后的）
func BinarySearch(slice []string, target string) bool {
	left := 0
	right := len(slice) - 1

	for left <= right {
		mid := (left + right) / 2
		if slice[mid] == target {
			return true
		} else if slice[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return false
}
