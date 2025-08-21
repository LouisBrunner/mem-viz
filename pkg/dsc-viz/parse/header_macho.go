package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
)

type linkEditData struct {
	block   *contracts.MemoryBlock
	command *subcontracts.SegmentCommand64 // TODO: support 32 bits
}

func (me *parser) parseMachO(frame *blockFrame, parent *contracts.MemoryBlock, path string) (*contracts.MemoryBlock, error) {
	header := subcontracts.MachHeader64{} // TODO: support 32 bits
	baseAddress := subcontracts.UnslidAddress(parent.Address - uintptr(me.slide))
	headerBlock, err := me.parseAndAdd(baseAddress.GetReader(frame.cache, 0, me.slide), parent, subcontracts.ManualAddress(0), &header, "Mach-O Header")
	if err != nil {
		return nil, err
	}
	cmdsOffset := uint64(unsafe.Sizeof(header))

	if header.Magic != subcontracts.MH_MAGIC_64 { // TODO: support 32 bits
		return nil, fmt.Errorf("invalid magic number %#x (expected %#x)", header.Magic, subcontracts.MH_MAGIC_64)
	}

	subFrame := frame.pushFrame(parent, headerBlock)
	commandsBlock, err := me.createBlobBlock(subFrame, "", baseAddress+subcontracts.UnslidAddress(cmdsOffset), "SizeOfCmds", uint64(header.SizeOfCmds), fmt.Sprintf("Commands (%d)", header.NCmds))
	if err != nil {
		return nil, err
	}

	var linkEdit linkEditData
	err = me.forEachMachOLoadCommand(subFrame, header, cmdsOffset, baseAddress, func(i int, address subcontracts.UnslidAddress, baseCommand subcontracts.LoadCommand) error {
		loadStruct, postParsing, err := me.getMachOLoadCommandParser(subFrame, baseCommand)
		if err != nil {
			return err
		}

		label := fmt.Sprintf("Load Command %d (%s)", i+1, subcontracts.LC2String(baseCommand.Cmd))
		parent := commandsBlock
		if uintptr(baseCommand.CmdSize) != parsingutils.GetDataValue(loadStruct).Type().Size() {
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
			_, err = postParsing(subFrame.pushFrame(parent, lcHeader), path, address, postDataAddress, &linkEdit)
			if err != nil {
				return err
			}
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
			return fmt.Errorf("failed to unpack load command %d for %+v: %w", i, frame, err)
		}

		err = callback(i, address, baseCommand)
		if err != nil {
			return err
		}

		offset += uint64(baseCommand.CmdSize)
	}

	return nil
}

type machOLoadCommandParser func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error)

func (me *parser) getMachOLoadCommandParser(frame *blockFrame, baseCommand subcontracts.LoadCommand) (interface{}, machOLoadCommandParser, error) {
	var subCommand interface{}
	var postParsing machOLoadCommandParser

	calculateLEAddress := func(linkEdit *linkEditData, offset subcontracts.LinkEditOffset) subcontracts.Address {
		return subcontracts.UnslidAddress(linkEdit.command.VMAddr + subcontracts.UnslidAddress(uint64(offset)-uint64(linkEdit.command.FileOff)))
	}

	addString := func(label string, offset *subcontracts.RelativeAddress32) machOLoadCommandParser {
		return func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			str := parsingutils.ReadCString(base.GetReader(frame.cache, uint64(*offset), me.slide))
			return me.createBlobBlock(frame, label, subcontracts.ManualAddress(*offset), "", uint64(len(str)+1), fmt.Sprintf("%s: %s", label, str))
		}
	}

	usePostLinkEdit := func(parent *contracts.MemoryBlock, apply func(ple *contracts.MemoryBlock) (*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, error) {
		// TODO: would be nice to group all of those loose LinkEdit blocks into a few
		// but I can't find a way to do it in a way that works for both the x86_86h split cache and the arm64e more consolidated one
		// ple, err := me.findOrCreateUniqueBlock(categoryPostLinkEdit, func(i int, block *contracts.MemoryBlock) bool {
		// 	return true
		// }, func() (*contracts.MemoryBlock, error) {
		// 	block, err := me.createCommonBlock(parent, "Post Link Edit", subcontracts.ManualAddress(0), 0)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	block.Address = math.MaxUint64
		// 	return block, nil
		// })
		// if err != nil {
		// 	return nil, err
		// }
		block, err := apply(nil)
		if err != nil {
			return nil, err
		}
		if block == nil {
			return nil, nil
		}
		// if block.Address < ple.Address {
		// 	ple.Address = block.Address
		// 	ple.Size += block.GetSize()
		// } else if !isInsideOf(block, ple) {
		// 	ple.Size = uint64(block.Address) - uint64(ple.Address) + block.GetSize()
		// }
		return block, nil
	}

	addSegment := func(segment *subcontracts.SegmentCommand64) machOLoadCommandParser {
		addSegmentHelper := func(frame *blockFrame, base, after subcontracts.Address, linkEdit *linkEditData, prefix string) (*contracts.MemoryBlock, error) {
			blob, err := me.createBlobBlock(frame, "VMAddr", segment.VMAddr, "VMSize", segment.VMSize, fmt.Sprintf("%sSegment %s", prefix, commons.FromCString(segment.SegName[:])))
			if err != nil {
				return nil, err
			}
			// TODO: don't know how to interpret this correctly, but it's basically duplicate information anyway
			// err = me.addLinkWithOffset(frame, "FileOff", segment.FileOff, "points to")
			// if err != nil {
			// 	return nil, err
			// }
			// if me.addSizeLink {
			// 	err = addLink(frame.parentStruct, "FileSize", blob, "gives size to")
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// }
			// TODO: should add the sections too
			return blob, nil
		}

		return func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			if commons.FromCString(segment.SegName[:]) == "__LINKEDIT" {
				absAddr := segment.VMAddr.AddBase(frame.parent.Address).Calculate(me.slide)

				block, err := me.findOrCreateUniqueBlock(categoryLinkEdit, func(i int, block *contracts.MemoryBlock) bool {
					return block.Address == absAddr
				}, func() (*contracts.MemoryBlock, error) {
					return addSegmentHelper(frame, base, after, linkEdit, "")
				})
				if err != nil {
					return nil, err
				}
				linkEdit.block = block
				linkEdit.command = segment
				return block, nil
			}

			return addSegmentHelper(frame, base, after, linkEdit, fmt.Sprintf("%s > ", path))
		}
	}

	type fieldLookup struct {
		labelOffset string
		offset      subcontracts.LinkEditOffset
		labelSize   string
		size        uint32
		data        interface{}
	}

	addLEOffsetFields := func(label string, fields map[string]fieldLookup, extra func(block *contracts.MemoryBlock) error) machOLoadCommandParser {
		return func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			for label, field := range fields {
				_, err := usePostLinkEdit(frame.parent, func(ple *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
					if field.offset == 0 {
						return nil, nil
					}
					offset := calculateLEAddress(linkEdit, field.offset)
					var block *contracts.MemoryBlock
					var err error
					if field.data != nil {
						if me.deepSearch {
							block, _, err = me.parseAndAddArray(frame, field.labelOffset, offset, field.labelSize, uint64(field.size), field.data, fmt.Sprintf("%s > %s > %s", path, label, label))
						} else {
							structSize := uint64(parsingutils.GetDataValue(field.data).Type().Size())
							block, err = me.createBlobBlock(frame, field.labelOffset, offset, field.labelSize, uint64(field.size)*structSize, fmt.Sprintf("%s > %s > %s", path, label, label))
						}
					} else {
						block, err = me.createBlobBlock(frame, field.labelOffset, offset, field.labelSize, uint64(field.size), fmt.Sprintf("%s > %s > %s", path, label, label))
					}
					if err != nil {
						return nil, err
					}
					if block == nil {
						return nil, nil
					}
					if extra != nil {
						err = extra(block)
					}
					return block, err
				})
				if err != nil {
					return nil, err
				}
			}
			return nil, nil
		}
	}

	addLESection := func(le *subcontracts.LinkEditDataCommand, extra func(block *contracts.MemoryBlock) error) machOLoadCommandParser {
		return addLEOffsetFields(subcontracts.LC2String(le.Cmd), map[string]fieldLookup{
			"Data": {"DataOff", le.DataOff, "DataSize", le.DataSize, nil},
		}, extra)
	}

	addDYLDInfo := func(di *subcontracts.DYLDInfoCommand, extra func(block *contracts.MemoryBlock) error) machOLoadCommandParser {
		return addLEOffsetFields(subcontracts.LC2String(di.Cmd), map[string]fieldLookup{
			"Rebase":   {"RebaseOff", di.RebaseOff, "RebaseSize", di.RebaseSize, nil},
			"Bind":     {"BindOff", di.BindOff, "BindSize", di.BindSize, nil},
			"WeakBind": {"WeakBindOff", di.WeakBindOff, "WeakBindSize", di.WeakBindSize, nil},
			"LazyBind": {"LazyBindOff", di.LazyBindOff, "LazyBindSize", di.LazyBindSize, nil},
			"Export":   {"ExportOff", di.ExportOff, "ExportSize", di.ExportSize, nil},
		}, extra)
	}

	// FIXME: untestable
	handleObsolete := func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
		return nil, fmt.Errorf("obsolete load command, unsupported")
	}

	// FIXME: untestable AND difficult to find
	handlePrivate := func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
		return nil, fmt.Errorf("private load command, unsupported")
	}

	// FIXME: unused
	handleUnused := func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
		return nil, fmt.Errorf("unusued load command (%s), currently unsupported", subcontracts.LC2String(baseCommand.Cmd))
	}

	// FIXME: unused but have a common struct
	handleUnusedExtra := func(block *contracts.MemoryBlock) error {
		return fmt.Errorf("unusued load command (%s), currently unsupported", block.Name)
	}

	switch baseCommand.Cmd {
	case subcontracts.LC_SEGMENT:
		realCommand := subcontracts.SegmentCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			return addSegment(&subcontracts.SegmentCommand64{
				Cmd:      subcontracts.LC_SEGMENT_64,
				CmdSize:  realCommand.CmdSize,
				SegName:  realCommand.SegName,
				VMAddr:   subcontracts.UnslidAddress(realCommand.VMAddr),
				VMSize:   uint64(realCommand.VMSize),
				FileOff:  subcontracts.RelativeAddress64(realCommand.FileOff),
				FileSize: uint64(realCommand.FileSize),
				MaxProt:  realCommand.MaxProt,
				InitProt: realCommand.InitProt,
				NSects:   realCommand.NSects,
				Flags:    realCommand.Flags,
			})(frame, path, base, after, linkEdit)
		}
	case subcontracts.LC_SYMTAB:
		realCommand := subcontracts.SymtabCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			block, err := addLEOffsetFields("SYMTAB", map[string]fieldLookup{
				"Symbols": {"SymOff", realCommand.SymOff, "NSyms", realCommand.NSyms, &subcontracts.NList64{}}, // TODO: wrong size on 32bit
			}, nil)(frame, path, base, after, linkEdit)
			if err != nil {
				return nil, err
			}
			strAddr := calculateLEAddress(linkEdit, realCommand.StrOff)
			_, err = usePostLinkEdit(frame.parent, func(ple *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
				return me.findOrCreateUniqueBlock(categoryStrings, func(i int, block *contracts.MemoryBlock) bool {
					return block.Address == strAddr.Calculate(me.slide)
				}, func() (*contracts.MemoryBlock, error) {
					// FIXME: would be nice to parse all those strings but probably _very_ noisy
					return me.createBlobBlock(frame, "StrOff", strAddr, "StrSize", uint64(realCommand.StrSize), "Strings")
				})
			})
			return block, err
		}
	case subcontracts.LC_SYMSEG:
		realCommand := subcontracts.SymSegCommand{}
		subCommand = &realCommand
		postParsing = handleObsolete
	case subcontracts.LC_THREAD:
		fallthrough
	case subcontracts.LC_UNIXTHREAD:
		header := frame.cache.Header()
		magic := commons.FromCString(header.Magic[:])
		switch magic {
		case subcontracts.MAGIC_arm64e:
			fallthrough
		case subcontracts.MAGIC_arm64:
			subCommand = &subcontracts.ThreadCommandARM64{}
		case subcontracts.MAGIC_x86_64:
			fallthrough
		case subcontracts.MAGIC_x86_64_HASWELL:
			subCommand = &subcontracts.ThreadCommandX8664{}
		default:
			return nil, nil, fmt.Errorf("unsupported architecture: %s", magic) // TODO: support more
		}
	case subcontracts.LC_LOADFVMLIB:
		fallthrough
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
		postParsing = func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			entry := uint32(0)
			return addLEOffsetFields("DYSYMTAB", map[string]fieldLookup{
				"TOC":           {"TOCOff", realCommand.TOCOff, "TOCSize", realCommand.NTOC, &entry},
				"Module Table":  {"ModTabOff", realCommand.ModTabOff, "ModTabSize", realCommand.NModTab, &entry},
				"External Refs": {"ExtRefOff", realCommand.ExtRefSymOff, "ExtRefSize", realCommand.NExtRefSyms, &entry},
				"Indirect Syms": {"IndirectSymOff", realCommand.IndirectSymOff, "IndirectSymSize", realCommand.NIndirectSyms, &entry},
				"External Rels": {"ExtRelOff", realCommand.ExtRelOff, "ExtRelSize", realCommand.NExtRel, &entry},
				"Local Rels":    {"LocRelOff", realCommand.LocRelOff, "LocRelSize", realCommand.NLocRel, &entry},
			}, nil)(frame, path, base, after, linkEdit)
		}
	case subcontracts.LC_LOAD_DYLIB:
		fallthrough
	case subcontracts.LC_ID_DYLIB:
		fallthrough
	case subcontracts.LC_LOAD_WEAK_DYLIB:
		fallthrough
	case subcontracts.LC_REEXPORT_DYLIB:
		fallthrough
	case subcontracts.LC_LAZY_LOAD_DYLIB:
		fallthrough
	case subcontracts.LC_LOAD_UPWARD_DYLIB:
		realCommand := subcontracts.DYLIBCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_LOAD_DYLINKER:
		fallthrough
	case subcontracts.LC_ID_DYLINKER:
		fallthrough
	case subcontracts.LC_DYLD_ENVIRONMENT:
		realCommand := subcontracts.DylinkerCommand{}
		subCommand = &realCommand
		postParsing = addString("Name", &realCommand.Name)
	case subcontracts.LC_PREBOUND_DYLIB:
		realCommand := subcontracts.PreboundDylibCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_ROUTINES:
		realCommand := subcontracts.RoutinesCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
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
		postParsing = handleUnused
	case subcontracts.LC_PREBIND_CKSUM:
		realCommand := subcontracts.PrebindCKSumCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_SEGMENT_64:
		realCommand := subcontracts.SegmentCommand64{}
		subCommand = &realCommand
		postParsing = addSegment(&realCommand)
	case subcontracts.LC_ROUTINES_64:
		realCommand := subcontracts.RoutinesCommand64{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_UUID:
		realCommand := subcontracts.UUIDCommand{}
		subCommand = &realCommand
	case subcontracts.LC_RPATH:
		realCommand := subcontracts.RPathCommand{}
		subCommand = &realCommand
		postParsing = addString("Path", &realCommand.Path)
	case subcontracts.LC_CODE_SIGNATURE:
		fallthrough
	case subcontracts.LC_SEGMENT_SPLIT_INFO:
		fallthrough
	case subcontracts.LC_DYLD_CHAINED_FIXUPS:
		fallthrough
	case subcontracts.LC_LINKER_OPTIMIZATION_HINT:
		fallthrough
	case subcontracts.LC_DYLIB_CODE_SIGN_DRS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = addLESection(&realCommand, handleUnusedExtra)
	case subcontracts.LC_ENCRYPTION_INFO:
		realCommand := subcontracts.EncryptionInfoCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_DYLD_INFO:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = addDYLDInfo(&realCommand, nil) // TODO: would be nice to have deeper parsing of the content
	case subcontracts.LC_DYLD_INFO_ONLY:
		realCommand := subcontracts.DYLDInfoCommand{}
		subCommand = &realCommand
		postParsing = addDYLDInfo(&realCommand, nil) // TODO: would be nice to have deeper parsing of the content
	case subcontracts.LC_VERSION_MIN_MACOSX:
		fallthrough
	case subcontracts.LC_VERSION_MIN_IPHONEOS:
		fallthrough
	case subcontracts.LC_VERSION_MIN_TVOS:
		fallthrough
	case subcontracts.LC_VERSION_MIN_WATCHOS:
		realCommand := subcontracts.VersionMinCommand{}
		subCommand = &realCommand
	case subcontracts.LC_FUNCTION_STARTS:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = addLESection(&realCommand, nil) // TODO: would be nice to have deeper parsing of the content
	case subcontracts.LC_MAIN:
		realCommand := subcontracts.EntryPointCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_DATA_IN_CODE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = addLESection(&realCommand, nil) // FIXME: always empty, but would be nice to have deeper parsing of the content
	case subcontracts.LC_SOURCE_VERSION:
		realCommand := subcontracts.SourceVersionCommand{}
		subCommand = &realCommand
	case subcontracts.LC_ENCRYPTION_INFO_64:
		realCommand := subcontracts.EncryptionInfoCommand64{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_LINKER_OPTION:
		realCommand := subcontracts.LinkerOptionCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_NOTE:
		realCommand := subcontracts.NoteCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	case subcontracts.LC_BUILD_VERSION:
		realCommand := subcontracts.BuildVersionCommand{}
		subCommand = &realCommand
		postParsing = func(frame *blockFrame, path string, base, after subcontracts.Address, linkEdit *linkEditData) (*contracts.MemoryBlock, error) {
			_, _, err := me.parseAndAddArray(frame, "", after, "NTools", uint64(realCommand.NTools), &subcontracts.BuildToolVersion{}, "Tools")
			return nil, err
		}
	case subcontracts.LC_DYLD_EXPORTS_TRIE:
		realCommand := subcontracts.LinkEditDataCommand{}
		subCommand = &realCommand
		postParsing = addLESection(&realCommand, nil) // TODO: would be nice to have deeper parsing of the content
	case subcontracts.LC_FILESET_ENTRY:
		realCommand := subcontracts.FilesetEntryCommand{}
		subCommand = &realCommand
		postParsing = handleUnused
	default:
		return nil, nil, fmt.Errorf("unknown command %#x", baseCommand.Cmd)
	}

	return subCommand, postParsing, nil
}
