package main

import (
	"io"
)

type Chunk struct {
	Start int64
	Size  int64
}

func createChunks(readerAt io.ReaderAt, fileSize int64, chunkSize int64) ([]Chunk, error) {
	var chunks []Chunk
	if fileSize == 0 {
		return chunks, nil
	}
	offset := int64(0)
	buf := make([]byte, chunkSize)
	for offset < fileSize {
		readSize := chunkSize
		if offset+readSize > fileSize {
			readSize = fileSize - offset
		}
		n, err := readerAt.ReadAt(buf[:readSize], offset)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		// Ищем последний '\n' в прочитанном блоке
		lastNewLine := -1
		for i := n - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				lastNewLine = i
				break
			}
		}
		var chunkEnd int64
		if lastNewLine >= 0 {
			chunkEnd = offset + int64(lastNewLine) + 1 // включая '\n'
		} else {
			// Нет перевода строки в блоке - дочитываем до следующего /n
			// Используем буферизированное чтение текущей позиции
			// Для простоты считаем, что строки не длиннее 10*chunkSize
			maxExtra := chunkSize * 10
			extraBuf := make([]byte, maxExtra)
			extraRead := int64(0)
			for extraRead < maxExtra {
				pos := offset + int64(n) + extraRead
				if pos >= fileSize {
					// Достигли конца файла
					chunkEnd = fileSize
					break
				}
				var readCount int
				readCount, err = readerAt.ReadAt(extraBuf[extraRead:extraRead+1], pos)
				if err != nil && err != io.EOF {
					return nil, err
				}
				if readCount == 0 {
					chunkEnd = pos
					break
				}
				if extraBuf[extraRead] == '\n' {
					chunkEnd = pos + 1
					break
				}
				extraRead++
			}
			if chunkEnd == 0 {
				chunkEnd = fileSize
			}
		}
		if chunkEnd <= offset {
			// Не удалось щпределить границу - берем все до конца
			chunkEnd = fileSize
		}
		chunks = append(chunks, Chunk{Start: offset, Size: chunkEnd - offset})
		offset = chunkEnd
	}
	return chunks, nil
}
