package main

import (
	"github.com/logpipe/logpipe/config"
	"github.com/logpipe/logpipe/core"
	"github.com/logpipe/logpipe/plugin"
	"strings"
)

func Register() {
	plugin.RegisterInputBuilder(&FileInputBuilder{})

}

type FileInputSpec struct {
	Path    string
	Delim   byte
	Include string
	Exclude string
}

type FileInput struct {
	core.BaseInput
	spec     FileInputSpec
	stop     chan struct{}
	consumer func(event core.Event)
	scanner  *FileScanner
	readers  map[string]*FileReader
}

func NewFileInput(name string, spec FileInputSpec) *FileInput {
	readers := make(map[string]*FileReader)
	return &FileInput{spec: spec, readers: readers}
}

func (i *FileInput) Start(consumer func(event core.Event)) error {
	i.consumer = consumer
	spec := i.spec
	paths := strings.Split(spec.Path, ";")

	scanner := NewFileScanner(paths, 10, spec.Include, spec.Exclude)
	scanner.Start()

	i.scanner = scanner
	go func() {
		for {
			select {
			case <-i.stop:
				return
			case f := <-scanner.Files:
				if _, ok := i.readers[f]; !ok {
					reader := NewFileReader(f, spec.Delim, i.accept)
					reader.Start()
					i.readers[f] = reader
				}
			}
		}
	}()

	return nil
}

func (i *FileInput) accept(line string) {
	if i.consumer != nil {
		var source interface{} = line
		if i.Codec() != nil {
			event, e := i.Codec().Decode(source)
			if e == nil {
				i.consumer(event)
			}
		} else {
			event := core.NewEvent(source)
			i.consumer(event)
		}
	}
}

func (i *FileInput) Stop() error {
	i.stop <- struct{}{}
	i.scanner.Stop()
	for _, r := range i.readers {
		r.Stop()
	}
	return nil
}

type FileInputBuilder struct {
}

func (b *FileInputBuilder) Kind() string {
	return "file"
}

func (b *FileInputBuilder) Build(name string, specValue config.Value) core.Input {
	spec := FileInputSpec{}
	specValue.Parse(&spec)

	return NewFileInput(name, spec)
}
