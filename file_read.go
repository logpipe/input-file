package main

import (
	"bufio"
	"os"
)

type FileReader struct {
	path     string
	delim    byte
	consumer func(string)
	stop     chan struct{}
	file     *os.File
}

func NewFileReader(path string, delim byte, consumer func(string)) *FileReader {
	stop := make(chan struct{})
	return &FileReader{path: path, delim: delim, consumer: consumer, stop: stop}
}

func (r *FileReader) Start() {
	file, err := os.OpenFile(r.path, os.O_RDONLY, os.ModeType)
	if err != nil {
		return
	}
	r.file = file
	go func() {
		defer r.file.Close()
		reader := bufio.NewReader(r.file)
		for {
			select {
			case <-r.stop:
				break
			default:
			}
			line, e := reader.ReadString(r.delim)
			if e == nil {
				r.consumer(line)
			}
		}
	}()
}

func (r *FileReader) Stop() {
	r.stop <- struct{}{}
}
