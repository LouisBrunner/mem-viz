package viz

import (
	"fmt"
	"io"
	"math"
	"strings"
	"unicode"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/dustin/go-humanize"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type blockVisitor = func(depth int, block *contracts.MemoryBlock) error
type valueVisitor = func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue) error
type linkVisitor = func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error

func visitEachBlockAux(root *contracts.MemoryBlock, depth int, visitor blockVisitor) error {
	if err := visitor(depth, root); err != nil {
		return err
	}
	for _, child := range root.Content {
		if err := visitEachBlockAux(child, depth+1, visitor); err != nil {
			return err
		}
	}
	return nil
}

func visitEachBlock(root *contracts.MemoryBlock, visitor blockVisitor) error {
	return visitEachBlockAux(root, 0, visitor)
}

func visitEachValue(root *contracts.MemoryBlock, visitor valueVisitor) error {
	return visitEachBlock(root, func(depth int, block *contracts.MemoryBlock) error {
		for _, value := range block.Values {
			if err := visitor(depth+1, block, value); err != nil {
				return err
			}
		}
		return nil
	})
}

func visitEachLink(root *contracts.MemoryBlock, visitor linkVisitor) error {
	return visitEachValue(root, func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue) error {
		for _, link := range value.Links {
			if err := visitor(depth+1, block, value, link); err != nil {
				return err
			}
		}
		return nil
	})
}

type linkOrigin struct {
	block *contracts.MemoryBlock
	value *contracts.MemoryValue
	link  *contracts.MemoryLink
}

func (me linkOrigin) String() string {
	return fmt.Sprintf("%s.%s", me.block.Name, me.value.Name)
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

func (me *outputter) Text(m contracts.MemoryBlock) error {
	links := map[uintptr][]linkOrigin{}
	visitEachLink(&m, func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error {
		links[uintptr(link.TargetAddress)] = append(links[uintptr(link.TargetAddress)], linkOrigin{
			block: block,
			value: value,
			link:  link,
		})
		return nil
	})

	linksIndex := 0
	linksOrder := maps.Keys(links)
	slices.Sort(linksOrder)

	builder := stringBuilder{w: me.w}

	// TODO: should be in the checker
	lastAddresses := map[int]uintptr{}

	flushLinks := func(upTo uintptr) {
		for linksIndex < len(linksOrder) && upTo > linksOrder[linksIndex] {
			addr := linksOrder[linksIndex]
			origins := links[linksOrder[linksIndex]]
			originsText := make([]string, len(origins))
			for i, origin := range origins {
				originsText[i] = origin.String()
			}
			builder.Writef("%#016x%17s %9s <- %s\n", addr, "", "", strings.Join(originsText, ", "))
			linksIndex += 1
		}
	}

	err := visitEachBlock(&m, func(depth int, block *contracts.MemoryBlock) error {
		flushLinks(block.Address)

		linksSuffixes := []string{}
		for linksIndex < len(linksOrder) && uintptr(block.Address) == linksOrder[linksIndex] {
			origins := links[linksOrder[linksIndex]]
			for _, origin := range origins {
				linksSuffixes = append(linksSuffixes, origin.String())
			}
			linksIndex += 1
		}
		linksSuffix := ""
		if len(linksSuffixes) > 0 {
			linksSuffix = fmt.Sprintf(" <- %s", strings.Join(linksSuffixes, ", "))
		}

		details := ""
		if len(block.Values) > 0 {
			detailsList := make([]string, len(block.Values))
			for i, value := range block.Values {
				detailsList[i] = fmt.Sprintf("%s:%s", makeAcronym(value.Name), value.Value)
			}
			details = fmt.Sprintf(" {%s}", strings.Join(detailsList, ","))
		}
		size := block.GetSize()
		builder.Writef("%#016x-%#016x [%7s] %s%s%s%s\n", block.Address, block.Address+uintptr(size), humanize.Bytes(size), indent(depth, "  "), block.Name, details, linksSuffix)

		if lastAddresses[depth] > block.Address {
			return fmt.Errorf("memory blocks are not sorted")
		}
		lastAddresses[depth] = block.Address + uintptr(size)
		return nil
	})
	if err != nil {
		return err
	}

	flushLinks(math.MaxUint64)

	return builder.Close()
}
