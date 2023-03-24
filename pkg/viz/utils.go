package viz

import (
	"fmt"
	"io"

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

func getLinks(block *contracts.MemoryBlock) map[uintptr][]linkOrigin {
	links := map[uintptr][]linkOrigin{}
	commons.VisitEachLink(block, func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error {
		links[uintptr(link.TargetAddress)] = append(links[uintptr(link.TargetAddress)], linkOrigin{
			block: block,
			value: value,
			link:  link,
		})
		return nil
	})
	return links
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

func (me *stringBuilder) Writef(format string, args ...interface{}) {
	if me.err != nil {
		return
	}
	_, me.err = me.w.Write([]byte(fmt.Sprintf(format, args...)))
}

func (me *stringBuilder) Close() error {
	return me.err
}
