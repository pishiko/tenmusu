package util

type Stack[T any] struct {
	elems []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(v T) {
	s.elems = append(s.elems, v)
}

func (s *Stack[T]) Pop() T {
	if len(s.elems) == 0 {
		return *new(T)
	}
	idx := len(s.elems) - 1
	v := s.elems[idx]
	s.elems = s.elems[:idx]
	return v
}

func (s *Stack[T]) Peek() T {
	if len(s.elems) == 0 {
		return *new(T)
	}
	return s.elems[len(s.elems)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elems) == 0
}
