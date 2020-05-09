package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type FileState struct {
	Name   string
	Path   string
	Offset int64
	Line   int64
	Mtime  int64
}

func NewFileState(path string) *FileState {
	stat, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return &FileState{Name: stat.Name(), Path: path, Offset: 0, Line: 0, Mtime: stat.ModTime().Unix()}
}

type FileStateSyncer struct {
	path     string
	stateMap map[string]*FileState
	interval uint32
	stop     chan struct{}
	lock     sync.RWMutex
}

func NewFileStateSyncer(path string, interval uint32) *FileStateSyncer {
	return &FileStateSyncer{path: path}
}

func (s *FileStateSyncer) Get(path string) *FileState {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if len(s.stateMap) > 0 {
		return s.stateMap[path]
	}
	return nil
}

func (s *FileStateSyncer) Set(state *FileState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.stateMap != nil {
		s.stateMap[state.Path] = state
	}
}

func (s *FileStateSyncer) Start() {
	s.load()
	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
			}
			if len(s.stateMap) > 0 {
				s.dump()
			}
			time.Sleep(time.Duration(s.interval) * time.Second)
		}
	}()
}

func (s *FileStateSyncer) Stop() {
	s.stop <- struct{}{}
}

func (s *FileStateSyncer) load() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stateMap = make(map[string]*FileState)
	states := make([]*FileState, 0)
	bytes, err := ioutil.ReadFile(s.path)
	if err != nil {
		return
	}
	json.Unmarshal(bytes, &states)
	if len(states) > 0 {
		for _, st := range states {
			s.stateMap[st.Path] = st
		}
	}
}

func (s *FileStateSyncer) dump() {
	s.lock.RLock()
	defer s.lock.RUnlock()
	states := make([]*FileState, len(s.stateMap))
	i := 0
	for _, st := range s.stateMap {
		states[i] = st
	}
	bytes, err := json.Marshal(states)
	if err != nil {
		return
	}
	_ = ioutil.WriteFile(s.path, bytes, os.ModeType)
}
