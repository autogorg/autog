package rag

import (
	"fmt"
	"autog"
)

const (
	DefaultChunkSize = 1500
)

type TextSplitter struct {
	ChunkSize int
}

func (ts *TextSplitter) GetParser() autog.ParserFunction {
	if ts.ChunkSize <= 0 {
		ts.ChunkSize = DefaultChunkSize
	}
	n := ts.ChunkSize
	parser := func (path string, payload interface{}) ([]autog.Chunk, error) {
		var chunks []autog.Chunk
		if path == autog.DOCUMENT_PATH_NONE {
			return chunks, fmt.Errorf("Document path is empty!")
		}
		content, ok := payload.(string)
		if !ok {
			return chunks, fmt.Errorf("Payload is not string type!")
		}
		runes := []rune(content)
		i := 0
		for i < len(runes) {
			//  0.5 * n -- 1.5 * n 
			j := min(i+int(1.5*float64(n)), len(runes))
			found := false
			for j > i+int(0.5*float64(n)) {
				chunk := string(runes[i:j])
				if chunk[len(chunk)-1] == '.' || chunk[len(chunk)-1] == '\n' {
					found = true
					break
				}
				j--
			}
			// n
			if !found {
				j = min(i+n, len(runes))
			}
			query := string(runes[i:j])
			chunk := &MemChunk{
				Index     : len(chunks),
				Path      : path,
				Query     : query,
				Content   : query,
				ByteStart : i,
				ByteEnd   : j,
				Payload   : "",
			}
			chunks = append(chunks, chunk)
			i = j
		}
		return chunks, nil
	}
	return parser
}