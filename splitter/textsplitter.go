package splitter

import (
	"autog"
)

type TextSplitter struct {

}

func (ts *TextSplitter) GetParser() autog.ParserFunction {
	parser := func (path string, payload interface{}) ([]autog.Chunk, error) {
		return []autog.Chunk, nil
	}
	return parser
}