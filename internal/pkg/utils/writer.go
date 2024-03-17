package utils

import (
	"fmt"
)

type ChunkedFunctionCaller interface {
	Write([]byte) (int, error)
	Flush() error
}

type chunkedFunctionCaller struct {
	f         func(data []byte) error
	chunk     []byte
	chunkSize int
}

func NewChunkedFunctionCaller(f func(data []byte) error, chunkSize int) ChunkedFunctionCaller {
	return &chunkedFunctionCaller{
		f:         f,
		chunk:     make([]byte, 0, chunkSize),
		chunkSize: chunkSize,
	}
}

func (m *chunkedFunctionCaller) do(data []byte) error {
	//log.Debug().Any("data_len", len(data)).Msg("writing to func")
	return m.f(data)
}

func (m *chunkedFunctionCaller) Write(data []byte) (int, error) {
	written := 0

	for len(m.chunk)+len(data) >= m.chunkSize {
		addL := m.chunkSize - len(m.chunk)         // вычисляем сколько надо добавить до полноты чанка
		m.chunk = append(m.chunk, data[0:addL]...) // добавляем в m.chunk до полноты
		data = data[addL:]                         // обрезаем то, что добавили

		if err := m.do(m.chunk); err != nil { // запускаемся с этими данными
			return written, fmt.Errorf("error in function: %w", err)
		}

		written += addL
		m.chunk = m.chunk[:0] // обнуляем чанк
	}
	m.chunk = append(m.chunk, data...)

	return written + len(data), nil
}

func (m *chunkedFunctionCaller) Flush() error {
	m.chunk = append(m.chunk, '\n')
	if err := m.f(m.chunk); err != nil {
		return fmt.Errorf("error in function: %w", err)
	}

	m.chunk = nil

	return nil
}
