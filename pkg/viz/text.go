package viz

import (
	"fmt"
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
	const thresholdsArrayTooBig = 10000
	const showLinks = true
	const showHiddenLinks = true
	const showProperties = true
	const showUnused = true

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
	formatNoAddr := "%18s"
	formatSize := "[%6s]"
	formatNoSize := "%8s"
	formatLink := fmt.Sprintf("%s %s %s <- %%s\n", formatAddr, formatNoAddr, formatNoSize)
	formatSkip := fmt.Sprintf("%s %s %s %%sSKIPPED (%%d items)\n", formatNoAddr, formatNoAddr, formatNoSize)
	formatMem := fmt.Sprintf("%s-%s %s %%s%%s%%s%%s\n", formatAddr, formatAddr, formatSize)
	formatUnused := fmt.Sprintf("%s-%s %s %%sUNUSED\n", formatAddr, formatAddr, formatSize)

	parentsEnd := map[int]uintptr{}
	hidden := map[int]bool{}

	flushUnused := func(from, to uintptr, depth int) {
		if !showUnused || from == 0 || from >= to {
			return
		}

		depths := maps.Keys(parentsEnd)
		slices.Sort(depths)
		for i := len(depths) - 1; i >= 0 && depths[i] >= depth; i-- {
			if hidden[i] {
				continue
			}
			idepth := depths[i]
			scopeEnd := parentsEnd[idepth]
			if from < scopeEnd && scopeEnd <= to {
				builder.Writef(formatUnused, from, scopeEnd, humanize.Bytes(uint64(scopeEnd-from)), indent(idepth+1, indentStr))
				from = scopeEnd
				parentsEnd[idepth] = from
			}
			if from >= to {
				return
			}
		}
		builder.Writef(formatUnused, from, to, humanize.Bytes(uint64(to-from)), indent(depth, indentStr))
	}

	flushEach := func(upTo, lastAddress uintptr, depth int) {
		lastUnused := lastAddress

		for linksIndex < len(linksOrder) && upTo > linksOrder[linksIndex] {
			if showLinks && showHiddenLinks {
				addr := linksOrder[linksIndex]

				// FIXME: depth is inaccurate
				flushUnused(lastUnused, addr, depth)

				origins := links[linksOrder[linksIndex]]
				originsText := make([]string, len(origins))
				for i, origin := range origins {
					originsText[i] = origin.String()
				}
				builder.Writef(formatLink, addr, "", "", strings.Join(originsText, ", "))
				lastUnused = addr
			}
			linksIndex += 1
		}

		flushUnused(lastUnused, upTo, depth)
	}

	lastAddress := uintptr(0)
	err := commons.VisitEachBlock(&m, func(ctx commons.VisitContext, block *contracts.MemoryBlock) error {
		size := block.GetSize()

		if ctx.Parent == nil || thresholdsArrayTooBig == 0 || (len(ctx.Parent.Content) < thresholdsArrayTooBig && !hidden[ctx.Depth-1]) {
			hidden[ctx.Depth] = false
			flushEach(block.Address, lastAddress, ctx.Depth)

			linksSuffix := ""
			if showLinks {
				linksSuffixes := []string{}
				for linksIndex < len(linksOrder) && uintptr(block.Address) == linksOrder[linksIndex] {
					origins := links[linksOrder[linksIndex]]
					for _, origin := range origins {
						linksSuffixes = append(linksSuffixes, origin.String())
					}
					linksIndex += 1
				}
				if len(linksSuffixes) > 0 {
					linksSuffix = fmt.Sprintf(" <- %s", strings.Join(linksSuffixes, ", "))
				}
			}

			details := ""
			if showProperties && len(block.Values) > 0 {
				detailsList := make([]string, len(block.Values))
				for i, value := range block.Values {
					detailsList[i] = fmt.Sprintf("%s:%s", makeAcronym(value.Name), value.Value)
				}
				details = fmt.Sprintf(" {%s}", strings.Join(detailsList, ","))
			}
			builder.Writef(formatMem, block.Address, block.Address+uintptr(size), humanize.Bytes(size), indent(ctx.Depth, indentStr), block.Name, details, linksSuffix)
			if thresholdsArrayTooBig != 0 && len(block.Content) > thresholdsArrayTooBig {
				builder.Writef(formatSkip, "", "", "", indent(ctx.Depth+1, indentStr), len(block.Content))
			}
		} else {
			hidden[ctx.Depth] = true
		}

		lastAddress = block.Address + uintptr(size)
		parentsEnd[ctx.Depth] = lastAddress
		return nil
	})
	if err != nil {
		return err
	}

	flushEach(m.Address+uintptr(m.Size), lastAddress, 0)

	return builder.Close()
}
