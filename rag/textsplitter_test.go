package rag_test

import (
	"fmt"
	"github.com/autogorg/autog"
)

var text string = `
<aaa> aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa </aaa>
{ aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa }
<bbb> bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb </bbb>
{ bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb }
<ccc> ccccccccccccccccccccccccccccccccccccc </ccc>
{ ccccccccccccccccccccccccccccccccccccc }
<ddd> ddddddddddddddddddddddddddddddddddddd </ddd>
{ ddddddddddddddddddddddddddddddddddddd }
<eee> eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee </eee>
{ eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee }
<fff> fffffffffffffffffffffffffffffffffffff </fff>
{ fffffffffffffffffffffffffffffffffffff }
`

func ExampleTextSplitter() {
	splitter := &TextSplitter{
		ChunkSize:       80,
		Overlap:         0.25,
		BreakStartChars: []rune { "<", "{" },
		BreakEndChars:   []rune { ">", "}" },
	}

	parser := splitter.GetParser()
	chunks, err := parser("", text)
	if err != nil {
		fmt.Println(err)
	}
	for _, chunk := range chunks {
		fmt.Println(chunk.GetContent())
	}

	// Output:
	// 
}