package macho

import (
	"os"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
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

	m, err := macho.Open(file)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	root := &contracts.MemoryBlock{
		Name: file,
		Size: uint64(st.Size()),
	}

	err = me.addHeader(root, m)
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
