package rag

import (
	"fmt"
	"github.com/autogorg/autog"
)

const (
	DefaultChunkSize  = 512
	DefaultMinOverlap = 0.01
	DefaultMaxOverlap = 0.99
)

type TextSplitter struct {
	ChunkSize int
	Overlap float64
	BreakStartChars []rune
	BreakEndChars   []rune
}

func NewTextSplitter(chunkSize int) *TextSplitter {
	return &TextSplitter{ChunkSize: chunkSize}
}

func (ts *TextSplitter) GetParser() autog.ParserFunction {
	if ts.ChunkSize <= 0 {
		ts.ChunkSize = DefaultChunkSize
	}
	if ts.Overlap <= DefaultMinOverlap{
		ts.Overlap = DefaultMinOverlap
	}
	if ts.Overlap >= DefaultMaxOverlap {
		ts.Overlap = DefaultMaxOverlap
	}

	size := ts.ChunkSize
	step := int((1.0-ts.Overlap)*float64(ts.ChunkSize))
	check := int(ts.Overlap*float64(ts.ChunkSize))
	check = min(int(float64(step)*0.5), check)

	NeedCheckBreak := func (start bool) bool {
		if start {
			return len(ts.BreakStartChars) > 0
		}
		return len(ts.BreakEndChars) > 0
	}

    CheckBreakChar := func(start bool, c rune) bool {
		chars := ts.BreakEndChars
		if start {
			chars = ts.BreakStartChars
		}
        for _, bc := range chars {
            if c == bc {
                return true
            }
        }
        return false
    }

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
			j := min(i+size, len(runes))
			if NeedCheckBreak(false) {
				j = min(i+size+check, len(runes))
				found := false
				for j > i+size && j > 0 {
					if CheckBreakChar(false, runes[j-1]) {
						found = true
						break
					}
					j--
				}
				if !found {
					j = min(i+size, len(runes))
				}
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

			nexti := i + step
			if NeedCheckBreak(true) {
				nexti = max(i+step-check, i + 1)
				found := false
				for nexti < i+step && nexti < len(runes) {
					if CheckBreakChar(true, runes[nexti]) {
						found = true
						break
					}
					nexti++
				}
				if !found {
					nexti = i + step
				}
			}
			i = nexti
		}
		return chunks, nil
	}
	return parser
}