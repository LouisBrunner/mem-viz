package macho

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type parser struct {
	logger    *logrus.Logger
	allBlocks map[uintptr]*[]*contracts.MemoryBlock
}

func Parse(logger *logrus.Logger, file string) (*contracts.MemoryBlock, error) {
	p := &parser{
		logger:    logger,
		allBlocks: make(map[uintptr]*[]*contracts.MemoryBlock),
	}
	return p.parse(file)
}

func (me *parser) parse(file string) (*contracts.MemoryBlock, error) {
	st, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	var m *macho.File
	f, err := macho.OpenFat(file)
	if err != nil {
		m, err = macho.Open(file)
		if err != nil {
			return nil, err
		}
		defer m.Close()

		return me.addArch(file, m, uint64(st.Size()))
	}

	root := &contracts.MemoryBlock{
		Name: file,
		Size: uint64(st.Size()),
	}

	for _, arch := range f.Arches {
		name := arch.CPU.String() // FIXME: need support for arm64e
		archBlock, err := me.addArch(fmt.Sprintf("Arch %s", name), arch.File, uint64(arch.Size))
		if err != nil {
			return nil, fmt.Errorf("failed to parse arch %s: %w", name, err)
		}
		archBlock.ParentOffset = uint64(arch.Offset)
		me.rebase(archBlock, uint64(arch.Offset))
		root.Content = append(root.Content, archBlock)
	}

	me.allBlocks = make(map[uintptr]*[]*contracts.MemoryBlock)
	fatHeader := me.addStructDetailed(root, f.FatHeader, "FAT Header", 0, 0, []string{"Arches"})
	for i, arch := range f.Arches {
		name := arch.CPU.String() // FIXME: need support for arm64e
		archBlock := me.addStruct(root, arch.FatArchHeader, fmt.Sprintf("FAT Arch %s", name), fatHeader.Size+uint64(i)*uint64(unsafe.Sizeof(macho.FatArchHeader{})))
		err := parsingutils.AddLinkWithAddr(archBlock, "Offset", "points to", uintptr(arch.Offset))
		if err != nil {
			return nil, err
		}
	}
	me.rebalance(root)

	return root, nil
}

func (me *parser) rebase(root *contracts.MemoryBlock, offset uint64) {
	root.Address = root.Address + uintptr(offset)
	for _, child := range root.Content {
		me.rebase(child, offset)
	}
	for _, value := range root.Values {
		for _, link := range value.Links {
			link.TargetAddress += offset
		}
	}
}

func (me *parser) addArch(name string, m *macho.File, sz uint64) (*contracts.MemoryBlock, error) {
	me.allBlocks = make(map[uintptr]*[]*contracts.MemoryBlock)

	root := &contracts.MemoryBlock{
		Name: name,
		Size: sz,
	}

	err := me.addHeader(root, m)
	if err != nil {
		return nil, err
	}

	me.rebalance(root)

	return root, nil
}

func (me *parser) rebalance(root *contracts.MemoryBlock) {
	addresses := maps.Keys(me.allBlocks)
	slices.Sort(addresses)

	for _, address := range addresses {
		sameAddress := me.allBlocks[address]
		for _, block := range *sameAddress {
			newParent := me.addChildDeep(root, block)
			block.ParentOffset = uint64(block.Address - newParent.Address)
		}
	}
}
