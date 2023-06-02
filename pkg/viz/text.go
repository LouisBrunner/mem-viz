package viz

import (
	"fmt"
	"strings"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/dustin/go-humanize"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func (me *outputter) Text(m contracts.MemoryBlock) error {
	// TODO: make this configurable on (.*)-viz
	const thresholdsArrayTooBig = 10000
	const showLinks = true
	const showHiddenLinks = true
	const showProperties = true
	const showUnused = true

	const indentStr = "  "

	links := getLinks(&m)

	linksOrder := maps.Keys(links)
	slices.Sort(linksOrder)
	nextLinkIndex := 0

	builder := stringBuilder{w: me.w}

	formatAddr := "%#016x"
	formatNoAddr := "%18s"
	formatSize := "[%6s]"
	formatNoSize := "%8s"
	formatLink := fmt.Sprintf("%s %s %s <- %%s\n", formatAddr, formatNoAddr, formatNoSize)
	formatSkip := fmt.Sprintf("%s %s %s %%sSKIPPED (%%d items)\n", formatNoAddr, formatNoAddr, formatNoSize)
	formatMem := fmt.Sprintf("%s-%s %s %%s%%s%%s%%s\n", formatAddr, formatAddr, formatSize)
	formatUnused := fmt.Sprintf("%s-%s %s %%sUNUSED\n", formatAddr, formatAddr, formatSize)

	writeBlock := func(ctx commons.VisitContext, block *contracts.MemoryBlock) {
		linksSuffix := ""
		if showLinks {
			linksSuffixes := []string{}
			for _, origin := range links[uintptr(block.Address)] {
				linksSuffixes = append(linksSuffixes, origin.String())
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
		size := block.GetSize()
		builder.Writef(formatMem, block.Address, block.Address+uintptr(size), humanize.Bytes(size), indent(ctx.Depth, indentStr), block.Name, details, linksSuffix)
	}

	flushOnlyUnused := func(from, to uintptr, depth int) {
		if !showUnused || from == 0 || from >= to {
			return
		}
		builder.Writef(formatUnused, from, to, humanize.Bytes(uint64(to-from)), indent(depth, indentStr))
	}

	flushUnused := func(from, to uintptr, depth int) {
		if !showUnused || from == 0 || from >= to {
			return
		}
		if !showHiddenLinks {
			flushOnlyUnused(from, to, depth)
			return
		}
		lastUnused := from
		for nextLinkIndex < len(linksOrder) {
			addr := linksOrder[nextLinkIndex]
			if addr <= from {
				nextLinkIndex += 1
				continue
			}
			if addr >= to {
				break
			}

			flushOnlyUnused(lastUnused, addr, depth)
			origins := links[addr]
			originsText := make([]string, len(origins))
			for i, origin := range origins {
				originsText[i] = origin.String()
			}
			builder.Writef(formatLink, addr, "", "", strings.Join(originsText, ", "))
			lastUnused = addr
			nextLinkIndex += 1
		}
		flushOnlyUnused(lastUnused, to, depth)
	}

	flushEmptyBlock := func(from, to uintptr, depth int) {
		if !showHiddenLinks {
			return
		}
		for i := nextLinkIndex; i < len(linksOrder); i += 1 {
			addr := linksOrder[i]
			if from >= addr {
				continue
			}
			if addr >= to {
				break
			}
			if from < addr && addr < to {
				flushUnused(from, to, depth)
				return
			}
		}
	}

	lastChildrenEnd := uintptr(0)
	err := commons.VisitEachBlockAdvanced(&m, commons.VisitorSetup{
		BeforeChildren: func(ctx commons.VisitContext, block *contracts.MemoryBlock) error {
			if ctx.PreviousSibling != nil {
				flushUnused(lastChildrenEnd, block.Address, ctx.Depth)
			}
			skipChildren := thresholdsArrayTooBig != 0 && len(block.Content) > thresholdsArrayTooBig
			writeBlock(ctx, block)
			end := block.Address + uintptr(block.GetSize())
			if skipChildren {
				builder.Writef(formatSkip, "", "", "", indent(ctx.Depth+1, indentStr), len(block.Content))
				// Delete extra links we won't be showing
				for i := nextLinkIndex; i < len(linksOrder); i += 1 {
					if linksOrder[i] >= end {
						nextLinkIndex = i
						break
					}
				}
			} else if len(block.Content) > 0 {
				flushUnused(block.Address, block.Content[0].Address, ctx.Depth+1)
			}
			ctx.OutBeforeChildrenSkip = skipChildren
			// each child will overwrite this (if any) and we will know where they end this way
			lastChildrenEnd = end
			return nil
		},
		AfterChildren: func(ctx commons.VisitContext, block *contracts.MemoryBlock) error {
			end := block.Address + uintptr(block.GetSize())
			if len(block.Content) == 0 {
				flushEmptyBlock(block.Address, end, ctx.Depth+1)
			} else {
				flushUnused(lastChildrenEnd, end, ctx.Depth+1)
			}
			lastChildrenEnd = end
			return nil
		},
	})
	if err != nil {
		return err
	}

	return builder.Close()
}
