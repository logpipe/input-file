package main

import (
	"bufio"
	"io"
	"os"
)

type FileReader struct {
	path     string
	delim    byte
	syncer   *FileStateSyncer
	consumer func(string)
	stop     chan struct{}
	file     *os.File
	state    *FileState
}

func NewFileReader(path string, delim byte, syncer *FileStateSyncer, consumer func(string)) *FileReader {
	stop := make(chan struct{})
	return &FileReader{path: path, delim: delim, syncer: syncer, consumer: consumer, stop: stop}
}

func (r *FileReader) Start() {
	r.state = r.syncer.Get(r.path)
	if r.state == nil {
		r.state = NewFileState(r.path)
		r.syncer.Set(r.state)
	}
	file, err := os.OpenFile(r.path, os.O_RDONLY, os.ModeType)
	if err != nil {
		return
	}
	r.file = file
	go func() {
		defer r.file.Close()
		if r.state.Offset > 0 {
			r.file.Seek(r.state.Offset, io.SeekCurrent)
		}
		offset, _ := r.file.Seek(0, io.SeekStart)
		r.state.Offset = offset
		reader := bufio.NewReader(r.file)
		for {
			select {
			case <-r.stop:
				break
			default:
			}
			line, e := reader.ReadString(r.delim)
			if e == nil {
				length := len(line)
				r.state.Offset += int64(length)
				r.state.Line += 1

				r.consumer(line)
			} else if e == io.EOF {
				stat, _ := r.file.Stat()
				if stat.Size() < r.state.Offset {
					r.file.Seek(0, io.SeekStart)
					r.state.Offset = 0
				}
			}
			stat, err := r.file.Stat()
			if err != nil {
				break
			}
			r.state.Mtime = stat.ModTime().Unix()
			r.syncer.Set(r.state)
		}
	}()
}

func (r *FileReader) Stop() {
	r.stop <- struct{}{}
}
