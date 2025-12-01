package cmp

func String(left, right string) Result {
	if left < right {
		return Less
	} else if left == right {
		return Equal
	} else {
		return Greater
	}
}
