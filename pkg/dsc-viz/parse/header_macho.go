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

	linkEditBase := uintptr(0)
	err = me.forEachMachOLoadCommand(subFrame, header, cmdsOffset, baseAddress, func(i int, address subcontracts.UnslidAddress, baseCommand subcontracts.LoadCommand) error {
		switch baseCommand.Cmd {
		case subcontracts.LC_SEGMENT:
			return fmt.Errorf("32-bit segments are not supported at the moment")
		case subcontracts.LC_SEGMENT_64:
			segment := subcontracts.SegmentCommand64{}
			err := commons.Unpack(address.GetReader(subFrame.cache, 0, me.slide), &segment)
			if err != nil {
				return err
			}
			if commons.FromCString(segment.SegName[:]) == "__LINKEDIT" {
				linkEditBase = uintptr(segment.VMAddr) + uintptr(me.slide)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = me.forEachMachOLoadCommand(subFrame, header, cmdsOffset, baseAddress, func(i int, address subcontracts.UnslidAddress, baseCommand subcontracts.LoadCommand) error {
		loadStruct, postParsing, err := me.getMachOLoadCommandParser(baseCommand)
		if err != nil {
			return err
		}

		label := fmt.Sprintf("Load Command %d (%s)", i+1, subcontracts.LC2String(baseCommand.Cmd))
		parent := commandsBlock
		if uintptr(baseCommand.CmdSize) != getDataValue(loadStruct).Type().Size() {
			metaBlock, err := me.createCommonBlock(parent, label, address, uint64(baseCommand.CmdSize))
			if err != nil {
				return err
			}
			label = fmt.Sprintf("%s Header", label)
			parent = metaBlock
		}

		lcHeader, err := me.parseAndAdd(address.GetReader(subFrame.cache, 0, me.slide), parent, address, loadStruct, label)
		if err != nil {
			return err
		}
		if postParsing != nil {
			postDataAddress := address + subcontracts.UnslidAddress(lcHeader.Size)
			err = postParsing(subFrame.pushFrame(parent, lcHeader), address, postDataAddress, linkEditBase)
			if err != nil {
				return err
			}
		} else if parent != commandsBlock {
			return fmt.Errorf("incomplete parsing for %s > %s", path, label)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return headerBlock, nil
}

func (me *parser) forEachMachOLoadCommand(frame *blockFrame, header subcontracts.MachHeader64, cmdsOffset uint64, baseAddress subcontracts.UnslidAddress, callback func(i int, address subcontracts.UnslidAddress, baseCommand subcontracts.LoadCommand) error) error {
	offset := uint64(0)
	for i := 0; i < int(header.NCmds); i += 1 {
		address := baseAddress + subcontracts.UnslidAddress(cmdsOffset+offset)

		baseCommand := subcontracts.LoadCommand{}
		err := commons.Unpack(address.GetReader(frame.cache, 0, me.slide), &baseCommand)
		if err != nil {
			return err
		}

		err = callback(i, address, baseCommand)
		if err != nil {
			return err
		}

		offset += uint64(baseCommand.CmdSize)
	}

	return nil
}

type machOLoadCommandParser func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error

func (me *parser) getMachOLoadCommandParser(baseCommand subcontracts.LoadCommand) (interface{}, machOLoadCommandParser, error) {
	var subCommand interface{}
	var postParsing machOLoadCommandParser

	addString := func(label string, offset *subcontracts.RelativeAddress32) machOLoadCommandParser {
		return func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			str := readCString(base.GetReader(frame.cache, uint64(*offset), me.slide))
			_, err := me.createBlobBlock(frame, label, subcontracts.ManualAddress(*offset), "", uint64(len(str)+1), fmt.Sprintf("%s: %s", label, str))
			return err
		}
	}

	handleObsolete := func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
		return fmt.Errorf("obsolete load command, unsupported")
	}

	handlePrivate := func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
		return fmt.Errorf("private load command, unsupported")
	}

	switch baseCommand.Cmd {
	case subcontracts.LC_SEGMENT:
		realCommand := subcontracts.SegmentCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_SYMTAB:
		realCommand := subcontracts.SymtabCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_SYMSEG:
		realCommand := subcontracts.SymSegCommand{}
		subCommand = &realCommand
		postParsing = handleObsolete
	case subcontracts.LC_THREAD:
		realCommand := subcontracts.ThreadCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_UNIXTHREAD:
		realCommand := subcontracts.ThreadCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LOADFVMLIB:
		realCommand := subcontracts.FVMLibCommand{}
		subCommand = &realCommand
		postParsing = handleObsolete
	case subcontracts.LC_IDFVMLIB:
		realCommand := subcontracts.FVMLibCommand{}
		subCommand = &realCommand
		postParsing = handleObsolete
	case subcontracts.LC_IDENT:
		realCommand := subcontracts.IdentCommand{}
		subCommand = &realCommand
		postParsing = handleObsolete
	case subcontracts.LC_FVMFILE:
		realCommand := subcontracts.FVMFileCommand{}
		subCommand = &realCommand
		postParsing = handlePrivate
	case subcontracts.LC_PREPAGE:
		realCommand := subcontracts.LoadCommand{}
		subCommand = &realCommand
		postParsing = handlePrivate
	case subcontracts.LC_DYSYMTAB:
		realCommand := subcontracts.DYSymTabCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LOAD_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_ID_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_LOAD_DYLINKER:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_ID_DYLINKER:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_PREBOUND_DYLIB:
		realCommand := subcontracts.PreboundDylibCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_ROUTINES:
		realCommand := subcontracts.RoutinesCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_SUB_FRAMEWORK:
		realCommand := subcontracts.SubFrameworkCommand{}
		subCommand = &realCommand
		postParsing = addString("Umbrella", &realCommand.Umbrella)
	case subcontracts.LC_SUB_UMBRELLA:
		realCommand := subcontracts.SubUmbrellaCommand{}
		subCommand = &realCommand
		postParsing = addString("SubUmbrella", &realCommand.SubUmbrella)
	case subcontracts.LC_SUB_CLIENT:
		realCommand := subcontracts.SubClientCommand{}
		subCommand = &realCommand
		postParsing = addString("Client", &realCommand.Client)
	case subcontracts.LC_SUB_LIBRARY:
		realCommand := subcontracts.SubLibraryCommand{}
		subCommand = &realCommand
		postParsing = addString("SubLibrary", &realCommand.SubLibrary)
	case subcontracts.LC_TWOLEVEL_HINTS:
		realCommand := subcontracts.TwoLevelHintsCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_PREBIND_CKSUM:
		realCommand := subcontracts.PrebindCKSumCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LOAD_WEAK_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_SEGMENT_64:
		realCommand := subcontracts.SegmentCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_ROUTINES_64:
		realCommand := subcontracts.RoutinesCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_UUID:
		realCommand := subcontracts.UUIDCommand{}
		subCommand = &realCommand
	case subcontracts.LC_RPATH:
		realCommand := subcontracts.RPathCommand{}
		subCommand = &realCommand
		postParsing = addString("Path", &realCommand.Path)
	case subcontracts.LC_CODE_SIGNATURE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_SEGMENT_SPLIT_INFO:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_REEXPORT_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_LAZY_LOAD_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_ENCRYPTION_INFO:
		realCommand := subcontracts.EncryptionInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_DYLD_INFO:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_DYLD_INFO_ONLY:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LOAD_UPWARD_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_VERSION_MIN_MACOSX:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
	case subcontracts.LC_VERSION_MIN_IPHONEOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
	case subcontracts.LC_FUNCTION_STARTS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_DYLD_ENVIRONMENT:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_MAIN:
		realCommand := subcontracts.EntryPointCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_DATA_IN_CODE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_SOURCE_VERSION:
		realCommand := subcontracts.SourceVersionCommand{}
		subCommand = &realCommand
	case subcontracts.LC_DYLIB_CODE_SIGN_DRS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_ENCRYPTION_INFO_64:
		realCommand := subcontracts.EncryptionInfoCommand64{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LINKER_OPTION:
		realCommand := subcontracts.LinkerOptionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_LINKER_OPTIMIZATION_HINT:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_VERSION_MIN_TVOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
	case subcontracts.LC_VERSION_MIN_WATCHOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
	case subcontracts.LC_NOTE:
		realCommand := subcontracts.NoteCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_BUILD_VERSION:
		realCommand := subcontracts.BuildVersionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			_, _, err := me.parseAndAddArray(frame, "", after, "NTools", uint64(realCommand.NTools), &subcontracts.BuildToolVersion{}, "Tools")
			return err
		}
	case subcontracts.LC_DYLD_EXPORTS_TRIE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_DYLD_CHAINED_FIXUPS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	case subcontracts.LC_FILESET_ENTRY:
		realCommand := subcontracts.FilesetEntryCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, base, after subcontracts.Address, linkEditBase uintptr) error {
			return nil // TODO: finish
		}
	default:
		return nil, nil, fmt.Errorf("unknown command %#x", baseCommand.Cmd)
	}

	return subCommand, postParsing, nil
}
