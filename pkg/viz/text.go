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
	// TODO: make this configurable on (mem|dsc)-viz
	const thresholdsArrayTooBig = 20

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

	parentsEnd := map[int]uintptr{}

	flushUnused := func(from, to uintptr, depth int) {
		if from == 0 || from >= to {
			return
		}

		scopeEnd := parentsEnd[depth]
		if from < scopeEnd && scopeEnd < to {
			builder.Writef(formatUnused, from, scopeEnd, humanize.Bytes(uint64(scopeEnd-from)), indent(depth+1, indentStr))
			from = scopeEnd
		}
		builder.Writef(formatUnused, from, to, humanize.Bytes(uint64(to-from)), indent(depth, indentStr))
	}

	flushLinks := func(upTo, lastAddress uintptr, depth int) {
		lastUnused := lastAddress
		for linksIndex < len(linksOrder) && upTo > linksOrder[linksIndex] {
			addr := linksOrder[linksIndex]

			flushUnused(lastUnused, addr, depth)

			origins := links[linksOrder[linksIndex]]
			originsText := make([]string, len(origins))
			for i, origin := range origins {
				originsText[i] = origin.String()
			}
			builder.Writef(formatLink, addr, "", "", strings.Join(originsText, ", "))
			linksIndex += 1

			lastUnused = addr
		}

		flushUnused(lastUnused, upTo, depth)
	}

	lastAddress := uintptr(0)
	err := commons.VisitEachBlock(&m, func(ctx commons.VisitContext, block *contracts.MemoryBlock) error {
		flushLinks(block.Address, lastAddress, ctx.Depth)

		size := block.GetSize()
		if ctx.Parent == nil || len(ctx.Parent.Content) < thresholdsArrayTooBig {
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
			builder.Writef(formatMem, block.Address, block.Address+uintptr(size), humanize.Bytes(size), indent(ctx.Depth, indentStr), block.Name, details, linksSuffix)
		}

		lastAddress = block.Address + uintptr(size)
		parentsEnd[ctx.Depth] = lastAddress
		return nil
	})
	if err != nil {
		return err
	}

	flushLinks(math.MaxUint64, 0, 0)

	return builder.Close()
}
