package util

import "errors"

var (
	ErrStackEmpty = errors.New("stack is empty")
	ErrNotFound = errors.New("")
)

type Stack[T any] struct {
	data []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{data: make([]T, 0)}
}

func (s *Stack[T]) Push(data T) {
	s.data = append(s.data, data)
}

func (s *Stack[T]) Pop() (T, error) {
	var data T
	if s.IsEmpty() {
		return data, ErrStackEmpty
	}

	data = s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	return data, nil
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

func (s *Stack[T]) Peek() (T, error) {
	var data T

	if s.IsEmpty() {
		return data, ErrStackEmpty
	}

	data = s.data[len(s.data)-1]
	return data, nil
}

func (s *Stack[T]) Size() int {
	return len(s.data)
}

func (s *Stack[T]) Get(index int) (T, error) {
	var val T
	var i int
	for i, val = range s.data {
		if i == index {
			return val, nil
		}
	}

	return val, ErrNotFound
}
