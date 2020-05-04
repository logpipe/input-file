package main

import "testing"
import "github.com/logpipe/logpipe/engine"

func TestFileInput(t *testing.T) {
	input := &FileInput{path: "data.json", delim: '\n'}
	engine.DebugInput(input)
	engine.Wait()
}
