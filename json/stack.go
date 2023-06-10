package json

type stack[T any] []T

func (s stack[T]) empty() bool {
	return len(s) == 0
}

func (s *stack[T]) pop() T {
	n := len(*s) - 1
	v := (*s)[n]
	*s = (*s)[:n]
	return v
}

func (s *stack[T]) push(v T) {
	*s = append(*s, v)
}

func (s stack[T]) top() T {
	if len(s) > 0 {
		return s[len(s)-1]
	}
	var zero T
	return zero
}

func stackContains[T comparable](s stack[T], v T) bool {
	for _, elem := range s {
		if elem == v {
			return true
		}
	}
	return false
}
