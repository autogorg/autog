package rag_test

import (
	"fmt"
	"github.com/autogorg/autog/rag"
)

var text string = `<aaa> aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa </aaa>
{ aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa }
<bbb> bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb </bbb>
{ bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb }
`

func ExampleTextSplitter() {
	splitter := &rag.TextSplitter{
		ChunkSize:       80,
		Overlap:         0.25,
		BreakStartChars: []rune { '<', '{' },
		BreakEndChars:   []rune { '>', '}' },
	}

	parser := splitter.GetParser()
	chunks, err := parser("/doc", text)
	if err != nil {
		fmt.Println(err)
	}
	for i, chunk := range chunks {
		fmt.Printf("%d ->\n%s\n", i, chunk.GetContent())
	}

	// Output:
	// 0 ->
	// <aaa> aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa </aaa>
	// { aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa }
	// <bbb>
	// 1 ->
	// </aaa>
	// { aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa }
	// <bbb> bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb </bbb>
	// 2 ->
	// <bbb> bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb </bbb>
	// { bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb }
	// 3 ->
	// </bbb>
	// { bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb }
	// 
}