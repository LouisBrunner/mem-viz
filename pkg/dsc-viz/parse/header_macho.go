package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseMachO(frame *blockFrame, parent *contracts.MemoryBlock, path string) (*contracts.MemoryBlock, error) {
	header := subcontracts.MachHeader64{}
	baseAddress := subcontracts.UnslidAddress(parent.Address - uintptr(me.slide))
	headerBlock, err := me.parseAndAdd(baseAddress.GetReader(frame.cache, 0, me.slide), parent, subcontracts.ManualAddress(0), &header, "Mach-O Header")
	if err != nil {
		return nil, err
	}
	cmdsOffset := uint64(unsafe.Sizeof(header))

	if header.Magic != subcontracts.MH_MAGIC_64 {
		return nil, fmt.Errorf("invalid magic number %#x (expected %#x)", header.Magic, subcontracts.MH_MAGIC_64)
	}

	subFrame := frame.pushFrame(parent, headerBlock)
	commandsBlock, err := me.createBlobBlock(subFrame, "", baseAddress+subcontracts.UnslidAddress(cmdsOffset), "SizeOfCmds", uint64(header.SizeOfCmds), fmt.Sprintf("Commands (%d)", header.NCmds))
	if err != nil {
		return nil, err
	}

	offset := uint64(0)
	for i := 0; i < int(header.NCmds); i += 1 {
		address := baseAddress + subcontracts.UnslidAddress(cmdsOffset+offset)

		baseCommand := subcontracts.LoadCommand{}
		err = commons.Unpack(address.GetReader(subFrame.cache, 0, me.slide), &baseCommand)
		if err != nil {
			return nil, err
		}

		loadStruct, postParsing, err := me.getMachOLoadCommandParser(baseCommand)
		if err != nil {
			return nil, err
		}

		lcHeader, err := me.parseAndAdd(address.GetReader(subFrame.cache, 0, me.slide), commandsBlock, address, loadStruct, fmt.Sprintf("Load Command %d (%s)", i+1, subcontracts.LC2String(baseCommand.Cmd)))
		if err != nil {
			return nil, err
		}
		if postParsing != nil {
			err = postParsing(subFrame.siblingFrame(lcHeader))
			if err != nil {
				return nil, err
			}
		}

		offset += uint64(baseCommand.CmdSize)
	}

	// TODO: should we add a struct (MH?)
	return headerBlock, nil
}

type machOLoadCommandParser func(*blockFrame) error

func (me *parser) getMachOLoadCommandParser(baseCommand subcontracts.LoadCommand) (interface{}, machOLoadCommandParser, error) {
	var subCommand interface{}
	var postParsing machOLoadCommandParser

	switch baseCommand.Cmd {
	case subcontracts.LC_SEGMENT:
		realCommand := subcontracts.SegmentCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SYMTAB:
		realCommand := subcontracts.SymtabCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SYMSEG:
		realCommand := subcontracts.SymSegCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_THREAD:
		realCommand := subcontracts.ThreadCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_UNIXTHREAD:
		realCommand := subcontracts.ThreadCommand{} // TODO: FIX!!!
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LOADFVMLIB:
		realCommand := subcontracts.FVMLibCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_IDFVMLIB:
		realCommand := subcontracts.FVMLibCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_IDENT:
		realCommand := subcontracts.IdentCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_FVMFILE:
		realCommand := subcontracts.FVMFileCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_PREPAGE:
		realCommand := subcontracts.LoadCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYSYMTAB:
		realCommand := subcontracts.DYSymTabCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LOAD_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ID_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LOAD_DYLINKER:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ID_DYLINKER:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_PREBOUND_DYLIB:
		realCommand := subcontracts.PreboundDylibCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ROUTINES:
		realCommand := subcontracts.RoutinesCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SUB_FRAMEWORK:
		realCommand := subcontracts.SubFrameworkCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SUB_UMBRELLA:
		realCommand := subcontracts.SubUmbrellaCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SUB_CLIENT:
		realCommand := subcontracts.SubClientCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SUB_LIBRARY:
		realCommand := subcontracts.SubLibraryCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_TWOLEVEL_HINTS:
		realCommand := subcontracts.TwoLevelHintsCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_PREBIND_CKSUM:
		realCommand := subcontracts.PrebindCKSumCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LOAD_WEAK_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SEGMENT_64:
		realCommand := subcontracts.SegmentCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ROUTINES_64:
		realCommand := subcontracts.RoutinesCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_UUID:
		realCommand := subcontracts.UUIDCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_RPATH:
		realCommand := subcontracts.RPathCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_CODE_SIGNATURE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SEGMENT_SPLIT_INFO:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_REEXPORT_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LAZY_LOAD_DYLIB:
		realCommand := subcontracts.LoadCommand{} // TODO: no idea??????
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ENCRYPTION_INFO:
		realCommand := subcontracts.EncryptionInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLD_INFO:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLD_INFO_ONLY:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LOAD_UPWARD_DYLIB:
		realCommand := subcontracts.LoadCommand{} // TODO: no idea??????
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_VERSION_MIN_MACOSX:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_VERSION_MIN_IPHONEOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_FUNCTION_STARTS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLD_ENVIRONMENT:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_MAIN:
		realCommand := subcontracts.EntryPointCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DATA_IN_CODE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_SOURCE_VERSION:
		realCommand := subcontracts.SourceVersionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLIB_CODE_SIGN_DRS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_ENCRYPTION_INFO_64:
		realCommand := subcontracts.EncryptionInfoCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LINKER_OPTION:
		realCommand := subcontracts.LinkerOptionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_LINKER_OPTIMIZATION_HINT:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_VERSION_MIN_TVOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_VERSION_MIN_WATCHOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_NOTE:
		realCommand := subcontracts.NoteCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_BUILD_VERSION:
		realCommand := subcontracts.BuildVersionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLD_EXPORTS_TRIE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_DYLD_CHAINED_FIXUPS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	case subcontracts.LC_FILESET_ENTRY:
		realCommand := subcontracts.FilesetEntryCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame) error {
			return nil
		}
	default:
		return nil, nil, fmt.Errorf("unknown command %#x", baseCommand.Cmd)
	}

	return subCommand, postParsing, nil
}
