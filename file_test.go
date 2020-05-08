package main

import "testing"
import "github.com/logpipe/logpipe/engine"

func TestFileInput(t *testing.T) {
	spec := FileInputSpec{Path: "*.json", Delim: '\n'}
	input := NewFileInput("file", spec)
	engine.DebugInput(input)
	engine.Wait()
}
