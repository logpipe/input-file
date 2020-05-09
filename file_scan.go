package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type FileScanner struct {
	paths    []string
	Files    chan string
	interval int
	include  string
	exclude  string
	stop     chan struct{}
}

func NewFileScanner(paths []string, interval int, include, exclude string) *FileScanner {
	files := make(chan string)
	stop := make(chan struct{})
	return &FileScanner{paths: paths, Files: files, interval: interval, include: include, exclude: exclude, stop: stop}
}

func (s *FileScanner) Start() {
	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
			}
			for _, p := range s.paths {
				s.scan(p)
			}
			time.Sleep(time.Duration(s.interval) * time.Second)
		}
	}()
}

func (s *FileScanner) Stop() {
	s.stop <- struct{}{}
}

func (s *FileScanner) scan(path string) error {
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}
	if matches != nil {
		for _, p := range matches {
			s.walk(p)
		}
	}
	return nil
}

func (s *FileScanner) walk(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}
	if stat.IsDir() {
		dir, err := ioutil.ReadDir(path)
		if err == nil {
			for _, st := range dir {
				if st.IsDir() {
					s.walk(filepath.Join(path, st.Name()))
				}
			}
		}
	} else {
		if filepath.IsAbs(path) && stat.Mode().IsRegular() {
			s.accept(path)
		}
	}
}

func (s *FileScanner) accept(path string) {
	if s.exclude != "" {
		excluded, err := filepath.Match(s.exclude, path)
		if err != nil || excluded {
			return
		}
	}
	if s.include != "" {
		included, err := filepath.Match(s.include, path)
		if err == nil && included {
			s.Files <- path
		}
	} else {
		s.Files <- path
	}

}
