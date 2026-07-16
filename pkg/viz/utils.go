package viz

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

type linkOrigin struct {
	block *contracts.MemoryBlock
	value *contracts.MemoryValue
	link  *contracts.MemoryLink
}

func (me linkOrigin) String() string {
	return fmt.Sprintf("%s.%s", me.block.Name, me.value.Name)
}

func getLinks(block *contracts.MemoryBlock) (map[uintptr][]linkOrigin, error) {
	links := map[uintptr][]linkOrigin{}
	err := commons.VisitEachLink(block, func(ctx commons.VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) {
		links[uintptr(link.TargetAddress)] = append(links[uintptr(link.TargetAddress)], linkOrigin{
			block: block,
			value: value,
			link:  link,
		})
	})
	if err != nil {
		return nil, err
	}
	return links, nil
}

type stringBuilder struct {
	w   io.Writer
	err error
}

func (me *stringBuilder) WriteString(s string) {
	if me.err != nil {
		return
	}
	_, me.err = me.w.Write([]byte(s))
}

func (me *stringBuilder) Writef(format string, args ...any) {
	if me.err != nil {
		return
	}
	_, me.err = fmt.Fprintf(me.w, format, args...)
}

func (me *stringBuilder) Close() error {
	return me.err
}

func indent(depth int, s string) string {
	return strings.Repeat(s, depth)
}

func makeAcronym(s string) string {
	uppers := []rune{}
	for _, c := range s {
		if unicode.IsUpper(c) {
			uppers = append(uppers, c)
		}
	}
	if len(uppers) == 0 {
		return s
	}
	return string(uppers)
}
