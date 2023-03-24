package viz

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/dustin/go-humanize"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func (me *outputter) Text(m contracts.MemoryBlock) error {
	const indentStr = "  "

	indent := func(depth int, s string) string {
		return strings.Repeat(s, depth)
	}

	makeAcronym := func(s string) string {
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

	links := getLinks(&m)

	linksIndex := 0
	linksOrder := maps.Keys(links)
	slices.Sort(linksOrder)

	builder := stringBuilder{w: me.w}

	formatAddr := "%#016x"
	formatNoAddr := "%17s"
	formatSize := "[%6s]"
	formatNoSize := "%8s"
	formatLink := fmt.Sprintf("%s %s %s <- %%s\n", formatAddr, formatNoAddr, formatNoSize)
	formatMem := fmt.Sprintf("%s-%s %s %%s%%s%%s%%s\n", formatAddr, formatAddr, formatSize)
	formatUnused := fmt.Sprintf("%s-%s %s %%sUNUSED\n", formatAddr, formatAddr, formatSize)

	flushLinks := func(upTo, lastAddress uintptr, depth int) {
		lastUnused := lastAddress
		for linksIndex < len(linksOrder) && upTo > linksOrder[linksIndex] {
			addr := linksOrder[linksIndex]

			if lastUnused != 0 && lastUnused < addr {
				builder.Writef(formatUnused, lastUnused, addr, humanize.Bytes(uint64(addr-lastUnused)), indent(depth, indentStr))
			}

			origins := links[linksOrder[linksIndex]]
			originsText := make([]string, len(origins))
			for i, origin := range origins {
				originsText[i] = origin.String()
			}
			builder.Writef(formatLink, addr, "", "", strings.Join(originsText, ", "))
			linksIndex += 1

			lastUnused = addr
		}

		if lastUnused != 0 && lastUnused < upTo {
			builder.Writef(formatUnused, lastUnused, upTo, humanize.Bytes(uint64(upTo-lastUnused)), indent(depth, indentStr))
		}
	}

	lastAddress := uintptr(0)
	err := commons.VisitEachBlock(&m, func(depth int, block *contracts.MemoryBlock) error {
		// TODO: show unused when a parent is a certain size but it's not fully mapped

		flushLinks(block.Address, lastAddress, depth)

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
		builder.Writef(formatMem, block.Address, block.Address+uintptr(size), humanize.Bytes(size), indent(depth, indentStr), block.Name, details, linksSuffix)

		lastAddress = block.Address + uintptr(size)
		return nil
	})
	if err != nil {
		return err
	}

	flushLinks(math.MaxUint64, 0, 0)

	return builder.Close()
}
