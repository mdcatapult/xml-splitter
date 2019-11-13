package main

import (
	"github.com/stretchr/testify/mock"
	"io"
)

type mockReader struct {
	data string
	done bool
	mock.Mock
}

func (m *mockReader) Read(bytes []byte) (int, error) {
	m.Called()
	copy(bytes, m.data)
	if m.done {
		return 0, io.EOF
	}
	m.done = true
	return len(m.data), nil
}

type mockWriter struct {
	mock.Mock
}

func (m *mockWriter) write(actions []ioAction) ([]ioAction, error) {
	args := m.Called()
	for len(actions) > 0 && actions[0].ready {
		actions = actions[1:]
	}
	return actions, args.Error(1)
}