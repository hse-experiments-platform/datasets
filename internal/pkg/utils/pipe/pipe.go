package pipe

import (
	"fmt"
	"io"
	"sync"
)

type PipeReadWriter struct {
	sync.Mutex             // Для синхронизации
	cond       *sync.Cond  // Условная переменная для блокировки/разблокировки
	data       []byte      // Буфер для данных
	chunkChan  chan []byte // Канал для получения чанков
	closed     bool        // Флаг закрытия канала
}

func NewPipeReadWriter() *PipeReadWriter {
	cp := &PipeReadWriter{
		chunkChan: make(chan []byte),
	}
	cp.cond = sync.NewCond(&cp.Mutex)

	go cp.processChunks()

	return cp
}

func (cp *PipeReadWriter) Close() {
	cp.Lock()
	defer cp.Unlock()

	if cp.closed == true {
		return
	}

	close(cp.chunkChan)
	cp.closed = true
	cp.cond.Broadcast() // Разблокируем ожидающие чтения
}

func (cp *PipeReadWriter) Write(p []byte) (n int, err error) {
	cp.Lock()
	defer cp.Unlock()

	if cp.closed == true {
		return 0, fmt.Errorf("writer is closed")
	}

	cp.chunkChan <- p
	return len(p), nil
}

func (cp *PipeReadWriter) Read(p []byte) (n int, err error) {
	cp.Lock()
	defer cp.Unlock()

	for len(cp.data) < len(p) { // Ожидаем пока в буфере не накопится достаточно данных
		if cp.closed {
			break // Выходим из цикла, если канал данных закрыт
		}
		cp.cond.Wait() // Блокируемся, ожидая новых данных или закрытия канала
	}

	// Если данные есть в буфере, скопируем столько, сколько можем
	if len(cp.data) > 0 {
		n = copy(p, cp.data)
		cp.data = cp.data[n:]
	}

	if len(cp.data) == 0 && cp.closed {
		// Если данных больше нет и канал закрыт, сообщаем о конце файла
		return n, io.EOF
	}

	// Если данных меньше запрошенного количества, но канал открыт, ожидаем продолжения
	if n < len(p) && !cp.closed {
		return n, nil
	}

	return n, nil // Вернуть количество скопированных байт и nil ошибка
}

func (cp *PipeReadWriter) processChunks() {
	for chunk := range cp.chunkChan {
		cp.Lock()
		cp.data = append(cp.data, chunk...)
		cp.cond.Broadcast() // Сигнализируем о новых данных
		cp.Unlock()
	}
}
