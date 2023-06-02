package macho

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
	"github.com/blacktop/go-macho/types"
)

type contextData struct {
	header   *macho.File
	text     *contracts.MemoryBlock
	linkEdit *contracts.MemoryBlock
}

type parseFn func(block, header *contracts.MemoryBlock) error

func (me *parser) addCommand(root, commands *contracts.MemoryBlock, i int, cmd macho.Load, offset uint64, context *contextData) (*contracts.MemoryBlock, error) {
	var data interface{}
	banned := []string{
		"LoadBytes",
	}
	var postParsing parseFn
	size := uint64(len(cmd.Raw()))
	headerSize := uint64(0)

	// FIXME: untestable
	handleObsolete := func(block, header *contracts.MemoryBlock) error {
		return fmt.Errorf("obsolete load command (%s), unsupported", block.Name)
	}

	// FIXME: untestable AND difficult to find
	handlePrivate := func(block, header *contracts.MemoryBlock) error {
		return fmt.Errorf("private load command (%s), unsupported", block.Name)
	}

	// FIXME: unused
	handleUnused := func(block, header *contracts.MemoryBlock) error {
		return fmt.Errorf("unusued load command (%s), currently unsupported", block.Name)
	}

	handleSegment := func(real *macho.Segment, headerSize uint64) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			switch real.Name {
			case "__TEXT":
				context.text = block
			case "__LINKEDIT":
				context.linkEdit = block
			}
			if real.Offset == 0 && real.Filesz == 0 {
				return nil
			}
			segment := me.addChild(root, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("Segment (%s)", real.Name),
				Address:      uintptr(real.Offset),
				Size:         uint64(real.Filesz),
				ParentOffset: real.Offset - uint64(root.Address),
			})
			err := parsingutils.AddLinkWithBlock(header, "Offset", segment, "points to")
			if err != nil {
				return err
			}
			i := uint64(0)
			// FIXME: SectionHeader has the wrong size (8 bytes too much in 64bit)
			sizeOfSection := uint64(unsafe.Sizeof(types.SectionHeader{}) - 8)
			for _, sect := range context.header.Sections {
				if sect.Seg != real.Name {
					continue
				}
				// FIXME: sorta breaks the viz but I want to keep them...
				// if sect.Size == 0 {
				// 	continue
				// }
				sectHeader := me.addStructDetailed(
					block,
					sect.SectionHeader,
					fmt.Sprintf("Section Header (%s)", sect.Name),
					uint64(headerSize)+i*sizeOfSection,
					sizeOfSection,
					[]string{"Type"},
				)
				if sect.Offset != 0 {
					sectData := me.addChild(segment, &contracts.MemoryBlock{
						Name:         fmt.Sprintf("Section (%s)", sect.Name),
						Address:      uintptr(sect.Offset),
						Size:         uint64(sect.Size),
						ParentOffset: uint64(sect.Offset),
					})
					err := parsingutils.AddLinkWithBlock(sectHeader, "Offset", sectData, "points to")
					if err != nil {
						return err
					}
				}
				i += 1
			}
			return nil
		}
	}

	handleThread := func(threads []types.ThreadState) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			offset := header.Size
			for i, thread := range threads {
				block := me.addStruct(block, thread, fmt.Sprintf("Thread State %d", i+1), offset)
				offset += block.Size
			}
			return nil
		}
	}

	handleLEData := func(data types.LinkEditDataCmd) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			if context.linkEdit == nil {
				return fmt.Errorf("no __LINKEDIT segment found")
			}
			segment := me.addChild(root, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("Segment (%s)", data.LoadCmd),
				Address:      uintptr(data.Offset),
				Size:         uint64(data.Size),
				ParentOffset: uint64(data.Offset) - uint64(root.Address),
			})
			err := parsingutils.AddLinkWithBlock(header, "Offset", segment, "points to")
			if err != nil {
				return err
			}
			return nil
		}
	}

	handleDYLDInfo := func(real *macho.DyldInfoOnly) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			links := []struct {
				name string
				prop string
				off  uint64
				size uint64
			}{
				{
					name: "Rebase",
					prop: "RebaseOff",
					off:  uint64(real.RebaseOff),
					size: uint64(real.RebaseSize),
				},
				{
					name: "Bind",
					prop: "BindOff",
					off:  uint64(real.BindOff),
					size: uint64(real.BindSize),
				},
				{
					name: "WeakBind",
					prop: "WeakBindOff",
					off:  uint64(real.WeakBindOff),
					size: uint64(real.WeakBindSize),
				},
				{
					name: "LazyBind",
					prop: "LazyBindOff",
					off:  uint64(real.LazyBindOff),
					size: uint64(real.LazyBindSize),
				},
				{
					name: "Export",
					prop: "ExportOff",
					off:  uint64(real.ExportOff),
					size: uint64(real.ExportSize),
				},
			}
			for _, link := range links {
				if link.off == 0 {
					continue
				}
				segment := me.addChild(root, &contracts.MemoryBlock{
					Name:         fmt.Sprintf("DYLD %s", link.name),
					Address:      uintptr(link.off),
					Size:         uint64(link.size),
					ParentOffset: uint64(link.off) - uint64(root.Address),
				})
				err := parsingutils.AddLinkWithBlock(header, link.prop, segment, "points to")
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	switch cmd.Command() {
	case types.LC_REQ_DYLD:
		return nil, fmt.Errorf("binary contains LC_REQ_DYLD which is not supported")
	case types.LC_SEGMENT, types.LC_SEGMENT_64:
		real := cmd.(*macho.Segment)
		data = real.SegmentHeader
		headerSize = uint64(unsafe.Sizeof(types.Segment32{}))
		if real.Command() == types.LC_SEGMENT_64 {
			headerSize = uint64(unsafe.Sizeof(types.Segment64{}))
		}
		postParsing = handleSegment(real, headerSize)
		banned = append(banned, "Firstsect") // FIXME: header is badly defined
	case types.LC_SYMTAB:
		real := cmd.(*macho.Symtab)
		data = real.SymtabCmd
		// TODO: add symbols
	case types.LC_SYMSEG:
		real := cmd.(*macho.SymSeg)
		data = real.SymsegCmd
		postParsing = handleObsolete
	case types.LC_THREAD:
		real := cmd.(*macho.Thread)
		data = real.ThreadCmd
		postParsing = handleThread(real.Threads)
	case types.LC_UNIXTHREAD:
		real := cmd.(*macho.UnixThread)
		data = real.ThreadCmd
		postParsing = handleThread(real.Threads)
	case types.LC_LOADFVMLIB:
		real := cmd.(*macho.LoadFvmlib)
		data = *real // .LoadFvmLibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleObsolete
	case types.LC_IDFVMLIB:
		real := cmd.(*macho.IDFvmlib)
		data = *real // .LoadFvmLibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleObsolete
	case types.LC_IDENT:
		real := cmd.(*macho.Ident)
		data = real.IdentCmd
		postParsing = handleObsolete
	case types.LC_FVMFILE:
		real := cmd.(*macho.FvmFile)
		data = *real // .FvmFileCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handlePrivate
	case types.LC_PREPAGE:
		real := cmd.(*macho.Prepage)
		data = real.PrePageCmd
		postParsing = handlePrivate
	case types.LC_DYSYMTAB:
		real := cmd.(*macho.Dysymtab)
		data = real.DysymtabCmd
		// TODO: add syms
	case types.LC_LOAD_DYLIB:
		real := cmd.(*macho.LoadDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ID_DYLIB:
		real := cmd.(*macho.IDDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_LOAD_DYLINKER:
		real := cmd.(*macho.LoadDylinker)
		data = *real // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ID_DYLINKER:
		real := cmd.(*macho.DylinkerID)
		data = *real // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_PREBOUND_DYLIB:
		real := cmd.(*macho.PreboundDylib)
		data = *real // .PreboundDylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleUnused
	case types.LC_ROUTINES:
		real := cmd.(*macho.Routines)
		data = real.RoutinesCmd
		postParsing = handleUnused
	case types.LC_SUB_FRAMEWORK:
		real := cmd.(*macho.SubFramework)
		data = *real // .SubFrameworkCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_UMBRELLA:
		real := cmd.(*macho.SubUmbrella)
		data = *real // .SubUmbrellaCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_CLIENT:
		real := cmd.(*macho.SubClient)
		data = *real // .SubClientCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_LIBRARY:
		real := cmd.(*macho.SubLibrary)
		data = *real // .SubLibraryCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_TWOLEVEL_HINTS:
		real := cmd.(*macho.TwolevelHints)
		data = real.TwolevelHintsCmd
		postParsing = handleUnused
	case types.LC_PREBIND_CKSUM:
		real := cmd.(*macho.PrebindCheckSum)
		data = real.PrebindCksumCmd
		postParsing = handleUnused
	case types.LC_LOAD_WEAK_DYLIB:
		real := cmd.(*macho.WeakDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ROUTINES_64:
		real := cmd.(*macho.Routines64)
		data = real.Routines64Cmd
		postParsing = handleUnused
	case types.LC_UUID:
		real := cmd.(*macho.UUID)
		data = real.UUIDCmd
	case types.LC_RPATH:
		real := cmd.(*macho.Rpath)
		data = *real // .RpathCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_CODE_SIGNATURE:
		real := cmd.(*macho.CodeSignature)
		data = real.CodeSignatureCmd
		postParsing = handleLEData(types.LinkEditDataCmd(real.CodeSignatureCmd))
	case types.LC_SEGMENT_SPLIT_INFO:
		real := cmd.(*macho.SplitInfo)
		data = real.SegmentSplitInfoCmd
		postParsing = handleLEData(types.LinkEditDataCmd(real.SegmentSplitInfoCmd))
	case types.LC_REEXPORT_DYLIB:
		real := cmd.(*macho.ReExportDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_LAZY_LOAD_DYLIB:
		real := cmd.(*macho.LazyLoadDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ENCRYPTION_INFO:
		real := cmd.(*macho.EncryptionInfo)
		data = real.EncryptionInfoCmd
		postParsing = handleUnused
	case types.LC_DYLD_INFO:
		real := cmd.(*macho.DyldInfo)
		data = real.DyldInfoCmd
		postParsing = handleUnused
	case types.LC_DYLD_INFO_ONLY:
		real := cmd.(*macho.DyldInfoOnly)
		data = real.DyldInfoCmd
		postParsing = handleDYLDInfo(real)
	case types.LC_LOAD_UPWARD_DYLIB:
		real := cmd.(*macho.UpwardDylib)
		data = *real // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_VERSION_MIN_MACOSX, types.LC_VERSION_MIN_IPHONEOS, types.LC_VERSION_MIN_TVOS, types.LC_VERSION_MIN_WATCHOS:
		real := cmd.(*macho.VersionMin)
		data = real.VersionMinCmd
	case types.LC_FUNCTION_STARTS:
		real := cmd.(*macho.FunctionStarts)
		data = real.LinkEditDataCmd
		postParsing = handleLEData(real.LinkEditDataCmd)
	case types.LC_DYLD_ENVIRONMENT:
		real := cmd.(*macho.DyldEnvironment)
		data = *real // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_MAIN:
		real := cmd.(*macho.EntryPoint)
		data = real.EntryPointCmd
		postParsing = func(block, header *contracts.MemoryBlock) error {
			if context.text == nil {
				return fmt.Errorf("no __TEXT segment found")
			}
			return parsingutils.AddLinkWithAddr(header, "EntryOffset", "points to", context.text.Address+uintptr(real.EntryOffset))
		}
	case types.LC_DATA_IN_CODE:
		real := cmd.(*macho.DataInCode)
		data = real.DataInCodeCmd
		postParsing = handleLEData(types.LinkEditDataCmd(real.DataInCodeCmd))
	case types.LC_SOURCE_VERSION:
		real := cmd.(*macho.SourceVersion)
		data = real.SourceVersionCmd
	case types.LC_DYLIB_CODE_SIGN_DRS:
		real := cmd.(*macho.DylibCodeSignDrs)
		data = real.LinkEditDataCmd
		postParsing = handleLEData(real.LinkEditDataCmd)
	case types.LC_ENCRYPTION_INFO_64:
		real := cmd.(*macho.EncryptionInfo64)
		data = real.EncryptionInfo64Cmd
		postParsing = handleUnused
	case types.LC_LINKER_OPTION:
		real := cmd.(*macho.LinkerOption)
		data = *real // .LinkerOptionCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleUnused
	case types.LC_LINKER_OPTIMIZATION_HINT:
		real := cmd.(*macho.LinkerOptimizationHint)
		data = real.LinkEditDataCmd
		postParsing = handleLEData(real.LinkEditDataCmd)
	case types.LC_NOTE:
		real := cmd.(*macho.Note)
		data = real.NoteCmd
		postParsing = handleUnused
	case types.LC_BUILD_VERSION:
		real := cmd.(*macho.BuildVersion)
		data = *real // .BuildVersionCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_DYLD_EXPORTS_TRIE:
		real := cmd.(*macho.DyldExportsTrie)
		data = real.LinkEditDataCmd
		postParsing = handleLEData(real.LinkEditDataCmd)
	case types.LC_DYLD_CHAINED_FIXUPS:
		real := cmd.(*macho.DyldChainedFixups)
		data = real.LinkEditDataCmd
		postParsing = handleLEData(real.LinkEditDataCmd)
	case types.LC_FILESET_ENTRY:
		real := cmd.(*macho.FilesetEntry)
		data = *real // .FilesetEntryCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleUnused
	default:
		return nil, fmt.Errorf("unknown command %#x", cmd.Command())
	}

	label := fmt.Sprintf("Command %d: %s", i+1, cmd.Command())
	var cmdBlock, headerBlock *contracts.MemoryBlock

	if headerSize == 0 {
		headerSize = uint64(reflect.ValueOf(data).Type().Size())
	}
	if size == headerSize {
		cmdBlock = me.addStructDetailed(commands, data, label, offset, size, banned)
		headerBlock = cmdBlock
	} else {
		cmdBlock = me.addChild(commands, &contracts.MemoryBlock{
			Name:         label,
			Address:      commands.Address + uintptr(offset),
			Size:         size,
			ParentOffset: offset,
		})
		headerBlock = me.addStructDetailed(cmdBlock, data, "Header", 0, headerSize, banned)
	}

	if postParsing != nil {
		err := postParsing(cmdBlock, headerBlock)
		if err != nil {
			return nil, err
		}
	}
	return cmdBlock, nil
}
