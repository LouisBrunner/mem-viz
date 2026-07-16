package macho

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
	"github.com/blacktop/go-macho/types"
)

type contextData struct {
	header              *macho.File
	text                *contracts.MemoryBlock
	linkEdit            *contracts.MemoryBlock
	symbols             *contracts.MemoryBlock
	symtab              *macho.Symtab
	stubSections        []*types.Section
	lazyPointerSections []*types.Section
}

type parseFn func(block, header *contracts.MemoryBlock) error

func (me *parser) addCommand(root, commands *contracts.MemoryBlock, i int, cmd macho.Load, offset uint64, context *contextData) (*contracts.MemoryBlock, error) { //nolint:gocyclo
	var data any
	banned := []string{
		"LoadBytes",
	}
	var postParsing parseFn
	size := uint64(len(cmd.Raw()))
	headerSize := uint64(0)

	// FIXME: untestable
	handleObsolete := func(block, _header *contracts.MemoryBlock) error {
		return fmt.Errorf("obsolete load command (%s), unsupported", block.Name)
	}

	// FIXME: untestable AND difficult to find
	handlePrivate := func(block, _header *contracts.MemoryBlock) error {
		return fmt.Errorf("private load command (%s), unsupported", block.Name)
	}

	// FIXME: unused
	handleUnused := func(block, _header *contracts.MemoryBlock) error {
		return fmt.Errorf("unused load command (%s), currently unsupported", block.Name)
	}

	handleSegment := func(realSeg *macho.Segment, headerSize uint64) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			if realSeg.Offset == 0 && realSeg.Filesz == 0 {
				return nil
			}
			segment := me.addChild(root, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("Segment (%s)", realSeg.Name),
				Address:      uintptr(realSeg.Offset),
				Size:         uint64(realSeg.Filesz),
				ParentOffset: realSeg.Offset - uint64(root.Address),
			})
			switch realSeg.Name {
			case "__TEXT":
				context.text = segment
			case "__LINKEDIT":
				context.linkEdit = segment
			}
			err := parsingutils.AddLinkWithBlock(header, "Offset", segment, "points to")
			if err != nil {
				return err
			}
			i := uint64(0)
			// FIXME: SectionHeader has the wrong size (8 bytes too much in 64bit)
			sizeOfSection := uint64(unsafe.Sizeof(types.SectionHeader{}) - 8)
			for _, sect := range context.header.Sections {
				if sect.Seg != realSeg.Name {
					continue
				}
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
					switch {
					case sect.Flags.IsSymbolStubs():
						context.stubSections = append(context.stubSections, sect)
					case sect.Flags.IsLazySymbolPointers():
						context.lazyPointerSections = append(context.lazyPointerSections, sect)
					case sect.Flags.IsNonLazySymbolPointers():
						// TODO: 8 byte per entry, don't know how to interpret this
					case sect.Flags.IsInterposing():
						// TODO: 2 pointers per entry, don't know how to interpret this
					case sect.Flags.IsCstringLiterals():
						for i := 0; i < int(sect.Size); {
							addr := sectData.Address + uintptr(i)
							str := parsingutils.ReadCString(io.NewSectionReader(context.header, int64(addr), int64(sect.Size)-int64(i)))
							strBlock := me.addChild(sectData, &contracts.MemoryBlock{
								Name:         fmt.Sprintf("%q", str),
								Address:      addr,
								Size:         uint64(len(str) + 1),
								ParentOffset: uint64(addr) - uint64(sectData.Address),
							})
							i += int(strBlock.Size)
						}
					default:
						switch sect.Name {
						case "__info_plist":
							plist := parsingutils.ReadCString(io.NewSectionReader(context.header, int64(sectData.Address), int64(sect.Size)))
							me.addChildDeep(sectData, &contracts.MemoryBlock{
								Name:         plist,
								Address:      sectData.Address,
								Size:         uint64(len(plist)),
								ParentOffset: 0,
							})
						}
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
				// FIXME: would be nice to have a link from PC/RIP
			}
			return nil
		}
	}

	handleLEData := func(data types.LinkEditDataCmd) parseFn {
		return func(_block, header *contracts.MemoryBlock) error {
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

	handleDYLDInfo := func(realDIO *macho.DyldInfoOnly) parseFn {
		return func(_block, header *contracts.MemoryBlock) error {
			links := []struct {
				name string
				prop string
				off  uint64
				size uint64
			}{
				{
					name: "Rebase",
					prop: "RebaseOff",
					off:  uint64(realDIO.RebaseOff),
					size: uint64(realDIO.RebaseSize),
				},
				{
					name: "Bind",
					prop: "BindOff",
					off:  uint64(realDIO.BindOff),
					size: uint64(realDIO.BindSize),
				},
				{
					name: "WeakBind",
					prop: "WeakBindOff",
					off:  uint64(realDIO.WeakBindOff),
					size: uint64(realDIO.WeakBindSize),
				},
				{
					name: "LazyBind",
					prop: "LazyBindOff",
					off:  uint64(realDIO.LazyBindOff),
					size: uint64(realDIO.LazyBindSize),
				},
				{
					name: "Export",
					prop: "ExportOff",
					off:  uint64(realDIO.ExportOff),
					size: uint64(realDIO.ExportSize),
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

	handleSymtab := func(realST *macho.Symtab) parseFn {
		return func(_block, header *contracts.MemoryBlock) error {
			// FIXME: assume 64bit
			symbols := me.addChild(root, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("Symbols (%d)", realST.Nsyms),
				Address:      uintptr(realST.Symoff),
				Size:         uint64(realST.Nsyms) * uint64(unsafe.Sizeof(subcontracts.NList64{})),
				ParentOffset: uint64(realST.Symoff) - uint64(root.Address),
			})
			context.symbols = symbols
			context.symtab = realST
			err := parsingutils.AddLinkWithBlock(header, "Symoff", symbols, "points to")
			if err != nil {
				return err
			}
			offset := uint64(0)
			for i, sym := range realST.Syms {
				// FIXME: assume 64bit
				nlist := subcontracts.NList64{
					NStrx:  0, // FIXME: unsupported because of the way they parse symbols
					NType:  uint8(sym.Type),
					NSect:  sym.Sect,
					NDesc:  uint16(sym.Desc),
					NValue: sym.Value,
				}
				symBlock := me.addStruct(symbols, nlist, fmt.Sprintf("Symbol (%d)", i+1), offset)
				addValue(symBlock, "Name", sym.Name, 0, 0)
				// FIXME: impossible due to the way they parse symbols
				// err = parsingutils.AddLinkWithAddr(header, "Stroff", "points to", uintptr(nlist.NStrx))
				// if err != nil {
				// 	return err
				// }
				offset += symBlock.Size
			}
			strings := me.addChild(root, &contracts.MemoryBlock{
				Name:         "Symbol Strings",
				Address:      uintptr(realST.Stroff),
				Size:         uint64(realST.Strsize),
				ParentOffset: uint64(realST.Stroff) - uint64(root.Address),
			})
			err = parsingutils.AddLinkWithBlock(header, "Stroff", strings, "points to")
			if err != nil {
				return err
			}
			for i := 0; i < int(realST.Strsize); {
				addr := strings.Address + uintptr(i)
				str := parsingutils.ReadCString(io.NewSectionReader(context.header, int64(addr), int64(realST.Strsize)-int64(i)))
				strBlock := me.addChild(strings, &contracts.MemoryBlock{
					Name:         fmt.Sprintf("%q", str),
					Address:      addr,
					Size:         uint64(len(str) + 1),
					ParentOffset: uint64(addr) - uint64(strings.Address),
				})
				i += int(strBlock.Size)
			}
			return nil
		}
	}

	handleDSymtab := func(realDST *macho.Dysymtab) parseFn {
		return func(block, header *contracts.MemoryBlock) error {
			if context.symbols == nil {
				return fmt.Errorf("no symbols found")
			}
			isyms := []struct {
				name string
				prop string
				off  uint64
				len  uint64
			}{
				{
					name: "Local Symbols",
					prop: "Ilocalsym",
					off:  uint64(realDST.Ilocalsym),
					len:  uint64(realDST.Nlocalsym),
				},
				{
					name: "External Symbols",
					prop: "Iextdefsym",
					off:  uint64(realDST.Iextdefsym),
					len:  uint64(realDST.Nextdefsym),
				},
				{
					name: "Undefined Symbols",
					prop: "Iundefsym",
					off:  uint64(realDST.Iundefsym),
					len:  uint64(realDST.Nundefsym),
				},
			}
			for _, isym := range isyms {
				if isym.off == 0 && isym.len == 0 {
					continue
				}
				// FIXME: assume 64bit
				entrySize := uintptr(unsafe.Sizeof(subcontracts.NList64{}))
				segment := me.addChild(context.symbols, &contracts.MemoryBlock{
					Name:         fmt.Sprintf("DSYM %s (%d)", isym.name, isym.len),
					Address:      context.symbols.Address + uintptr(isym.off)*entrySize,
					Size:         isym.len * uint64(entrySize),
					ParentOffset: isym.off * uint64(entrySize),
				})
				err := parsingutils.AddLinkWithBlock(header, isym.prop, segment, "points to")
				if err != nil {
					return err
				}
			}
			eentries := []struct {
				name   string
				prop   string
				off    uint64
				len    uint64
				sizeOf uint64
			}{
				{
					name:   "TOC",
					prop:   "Tocoffset",
					off:    uint64(realDST.Tocoffset),
					len:    uint64(realDST.Ntoc),
					sizeOf: uint64(unsafe.Sizeof(types.DylibTableOfContents{})),
				},
				{
					name: "Module Table",
					prop: "Modtaboff",
					off:  uint64(realDST.Modtaboff),
					len:  uint64(realDST.Nmodtab),
					// FIXME: assume 64bit
					sizeOf: uint64(unsafe.Sizeof(types.DylibModule64{})),
				},
				{
					name:   "External Relocations Table",
					prop:   "Extrefsymoff",
					off:    uint64(realDST.Extrefsymoff),
					len:    uint64(realDST.Nextrefsyms),
					sizeOf: uint64(unsafe.Sizeof(types.DylibReference(0))),
				},
				{
					name:   "Indirect Symbols Table",
					prop:   "Indirectsymoff",
					off:    uint64(realDST.Indirectsymoff),
					len:    uint64(realDST.Nindirectsyms),
					sizeOf: uint64(unsafe.Sizeof(types.DylibReference(0))),
				},
				{
					name:   "External Symbols Table",
					prop:   "Extreloff",
					off:    uint64(realDST.Extreloff),
					len:    uint64(realDST.Nextrel),
					sizeOf: uint64(unsafe.Sizeof(types.Reloc{})),
				},
				{
					name:   "Local Symbols Table",
					prop:   "Locreloff",
					off:    uint64(realDST.Locreloff),
					len:    uint64(realDST.Nlocrel),
					sizeOf: uint64(unsafe.Sizeof(types.Reloc{})),
				},
			}
			for _, entries := range eentries {
				if entries.off == 0 && entries.len == 0 {
					continue
				}
				segment := me.addChild(root, &contracts.MemoryBlock{
					Name:         fmt.Sprintf("DSYM %s (%d)", entries.name, entries.len),
					Address:      uintptr(entries.off),
					Size:         uint64(entries.len * entries.sizeOf),
					ParentOffset: uint64(entries.off) - uint64(root.Address),
				})
				err := parsingutils.AddLinkWithBlock(header, entries.prop, segment, "points to")
				if err != nil {
					return err
				}
				switch entries.prop {
				case "Indirectsymoff":
					names := make([]string, realDST.Nindirectsyms)
					for i, sym := range realDST.IndirectSyms {
						entry := me.addChild(segment, &contracts.MemoryBlock{
							Name:         fmt.Sprintf("Indirect Symbol (%d)", i+1),
							Address:      uintptr(entries.off) + uintptr(i)*uintptr(entries.sizeOf),
							Size:         entries.sizeOf,
							ParentOffset: uint64(i) * entries.sizeOf,
						})
						addValue(entry, "Index", sym, 0, 0)
						name := "not found"
						if sym < uint32(len(context.symtab.Syms)) {
							name = context.symtab.Syms[sym].Name
							err := parsingutils.AddLinkWithAddr(entry, "Index", "refers to", context.symbols.Address+uintptr(sym)*unsafe.Sizeof(subcontracts.NList64{}))
							if err != nil {
								return err
							}
						}
						addValue(entry, "Name", name, 0, 0)
						names[i] = name
					}
					for _, sect := range context.stubSections {
						indirectIndex := sect.Reserved1
						sizeOfStub := uint64(sect.Reserved2)
						numberOfStubs := sect.Size / sizeOfStub
						if uint64(indirectIndex)+numberOfStubs > uint64(realDST.Nindirectsyms) {
							return fmt.Errorf("invalid indirect symbol index in %s (%d vs %d)", sect.Name, uint64(indirectIndex)+numberOfStubs, realDST.Nindirectsyms)
						}
						for i := range numberOfStubs {
							index := indirectIndex + uint32(i)
							me.addChild(segment, &contracts.MemoryBlock{
								Name:         fmt.Sprintf("Stub of %s", names[index]),
								Address:      uintptr(sect.Offset) + uintptr(i*sizeOfStub),
								Size:         sizeOfStub,
								ParentOffset: i * sizeOfStub,
							})
							// TODO: how to add pointer to the lazy pointer section?
						}
					}
					for _, sect := range context.lazyPointerSections {
						indirectIndex := sect.Reserved1
						pointerSize := uint64(8) // TODO: assume 64bit
						numberOfPointers := sect.Size / uint64(pointerSize)
						if uint64(indirectIndex)+numberOfPointers > uint64(realDST.Nindirectsyms) {
							return fmt.Errorf("invalid indirect symbol index in %s (%d vs %d)", sect.Name, uint64(indirectIndex)+numberOfPointers, realDST.Nindirectsyms)
						}
						for i := range numberOfPointers {
							index := indirectIndex + uint32(i)
							me.addChild(segment, &contracts.MemoryBlock{
								Name:         fmt.Sprintf("Lazy Pointer for %s", names[index]),
								Address:      uintptr(sect.Offset) + uintptr(i*pointerSize),
								Size:         uint64(pointerSize),
								ParentOffset: i * uint64(pointerSize),
							})
						}
					}
				default:
					// TODO: add details for each entry
				}
			}
			return nil
		}
	}

	switch cmd.Command() {
	case types.LC_REQ_DYLD:
		return nil, fmt.Errorf("binary contains LC_REQ_DYLD which is not supported")
	case types.LC_SEGMENT, types.LC_SEGMENT_64:
		realSeg := cmd.(*macho.Segment)
		data = realSeg.SegmentHeader
		headerSize = uint64(unsafe.Sizeof(types.Segment32{}))
		if realSeg.Command() == types.LC_SEGMENT_64 {
			headerSize = uint64(unsafe.Sizeof(types.Segment64{}))
		}
		postParsing = handleSegment(realSeg, headerSize)
		banned = append(banned, "Firstsect") // FIXME: header is badly defined
	case types.LC_SYMTAB:
		realSeg := cmd.(*macho.Symtab)
		data = realSeg.SymtabCmd
		postParsing = handleSymtab(realSeg)
	case types.LC_SYMSEG:
		realSeg := cmd.(*macho.SymSeg)
		data = realSeg.SymsegCmd
		postParsing = handleObsolete
	case types.LC_THREAD:
		realSeg := cmd.(*macho.Thread)
		data = realSeg.ThreadCmd
		postParsing = handleThread(realSeg.Threads)
	case types.LC_UNIXTHREAD:
		realSeg := cmd.(*macho.UnixThread)
		data = realSeg.ThreadCmd
		postParsing = handleThread(realSeg.Threads)
	case types.LC_LOADFVMLIB:
		realSeg := cmd.(*macho.LoadFvmlib)
		data = *realSeg // .LoadFvmLibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleObsolete
	case types.LC_IDFVMLIB:
		realSeg := cmd.(*macho.IDFvmlib)
		data = *realSeg // .LoadFvmLibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleObsolete
	case types.LC_IDENT:
		realSeg := cmd.(*macho.Ident)
		data = realSeg.IdentCmd
		postParsing = handleObsolete
	case types.LC_FVMFILE:
		realSeg := cmd.(*macho.FvmFile)
		data = *realSeg // .FvmFileCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handlePrivate
	case types.LC_PREPAGE:
		realSeg := cmd.(*macho.Prepage)
		data = realSeg.PrePageCmd
		postParsing = handlePrivate
	case types.LC_DYSYMTAB:
		realSeg := cmd.(*macho.Dysymtab)
		data = realSeg.DysymtabCmd
		postParsing = handleDSymtab(realSeg)
	case types.LC_LOAD_DYLIB:
		realSeg := cmd.(*macho.LoadDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ID_DYLIB:
		realSeg := cmd.(*macho.IDDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_LOAD_DYLINKER:
		realSeg := cmd.(*macho.LoadDylinker)
		data = *realSeg // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ID_DYLINKER:
		realSeg := cmd.(*macho.DylinkerID)
		data = *realSeg // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_PREBOUND_DYLIB:
		realSeg := cmd.(*macho.PreboundDylib)
		data = *realSeg // .PreboundDylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleUnused
	case types.LC_ROUTINES:
		realSeg := cmd.(*macho.Routines)
		data = realSeg.RoutinesCmd
		postParsing = handleUnused
	case types.LC_SUB_FRAMEWORK:
		realSeg := cmd.(*macho.SubFramework)
		data = *realSeg // .SubFrameworkCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_UMBRELLA:
		realSeg := cmd.(*macho.SubUmbrella)
		data = *realSeg // .SubUmbrellaCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_CLIENT:
		realSeg := cmd.(*macho.SubClient)
		data = *realSeg // .SubClientCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_SUB_LIBRARY:
		realSeg := cmd.(*macho.SubLibrary)
		data = *realSeg // .SubLibraryCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_TWOLEVEL_HINTS:
		realSeg := cmd.(*macho.TwolevelHints)
		data = realSeg.TwolevelHintsCmd
		postParsing = handleUnused
	case types.LC_PREBIND_CKSUM:
		realSeg := cmd.(*macho.PrebindCheckSum)
		data = realSeg.PrebindCksumCmd
		postParsing = handleUnused
	case types.LC_LOAD_WEAK_DYLIB:
		realSeg := cmd.(*macho.WeakDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ROUTINES_64:
		realSeg := cmd.(*macho.Routines64)
		data = realSeg.Routines64Cmd
		postParsing = handleUnused
	case types.LC_UUID:
		realSeg := cmd.(*macho.UUID)
		data = realSeg.UUIDCmd
	case types.LC_RPATH:
		realSeg := cmd.(*macho.Rpath)
		data = *realSeg // .RpathCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_CODE_SIGNATURE:
		realSeg := cmd.(*macho.CodeSignature)
		data = realSeg.CodeSignatureCmd
		postParsing = handleLEData(types.LinkEditDataCmd(realSeg.CodeSignatureCmd))
	case types.LC_SEGMENT_SPLIT_INFO:
		realSeg := cmd.(*macho.SplitInfo)
		data = realSeg.SegmentSplitInfoCmd
		postParsing = handleLEData(types.LinkEditDataCmd(realSeg.SegmentSplitInfoCmd))
	case types.LC_REEXPORT_DYLIB:
		realSeg := cmd.(*macho.ReExportDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_LAZY_LOAD_DYLIB:
		realSeg := cmd.(*macho.LazyLoadDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_ENCRYPTION_INFO:
		realSeg := cmd.(*macho.EncryptionInfo)
		data = realSeg.EncryptionInfoCmd
		postParsing = handleUnused
	case types.LC_DYLD_INFO:
		realSeg := cmd.(*macho.DyldInfo)
		data = realSeg.DyldInfoCmd
		postParsing = handleUnused
	case types.LC_DYLD_INFO_ONLY:
		realSeg := cmd.(*macho.DyldInfoOnly)
		data = realSeg.DyldInfoCmd
		postParsing = handleDYLDInfo(realSeg)
	case types.LC_LOAD_UPWARD_DYLIB:
		realSeg := cmd.(*macho.UpwardDylib)
		data = *realSeg // .DylibCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_VERSION_MIN_MACOSX, types.LC_VERSION_MIN_IPHONEOS, types.LC_VERSION_MIN_TVOS, types.LC_VERSION_MIN_WATCHOS:
		realSeg := cmd.(*macho.VersionMin)
		data = realSeg.VersionMinCmd
	case types.LC_FUNCTION_STARTS:
		realSeg := cmd.(*macho.FunctionStarts)
		data = realSeg.LinkEditDataCmd
		postParsing = handleLEData(realSeg.LinkEditDataCmd)
	case types.LC_DYLD_ENVIRONMENT:
		realSeg := cmd.(*macho.DyldEnvironment)
		data = *realSeg // .DylinkerCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_MAIN:
		realSeg := cmd.(*macho.EntryPoint)
		data = realSeg.EntryPointCmd
		postParsing = func(block, header *contracts.MemoryBlock) error {
			if context.text == nil {
				return fmt.Errorf("no __TEXT segment found")
			}
			return parsingutils.AddLinkWithAddr(header, "EntryOffset", "points to", context.text.Address+uintptr(realSeg.EntryOffset))
		}
	case types.LC_DATA_IN_CODE:
		realSeg := cmd.(*macho.DataInCode)
		data = realSeg.DataInCodeCmd
		postParsing = handleLEData(types.LinkEditDataCmd(realSeg.DataInCodeCmd))
	case types.LC_SOURCE_VERSION:
		realSeg := cmd.(*macho.SourceVersion)
		data = realSeg.SourceVersionCmd
	case types.LC_DYLIB_CODE_SIGN_DRS:
		realSeg := cmd.(*macho.DylibCodeSignDrs)
		data = realSeg.LinkEditDataCmd
		postParsing = handleLEData(realSeg.LinkEditDataCmd)
	case types.LC_ENCRYPTION_INFO_64:
		realSeg := cmd.(*macho.EncryptionInfo64)
		data = realSeg.EncryptionInfo64Cmd
		postParsing = handleUnused
	case types.LC_LINKER_OPTION:
		realSeg := cmd.(*macho.LinkerOption)
		data = *realSeg // .LinkerOptionCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
		postParsing = handleUnused
	case types.LC_LINKER_OPTIMIZATION_HINT:
		realSeg := cmd.(*macho.LinkerOptimizationHint)
		data = realSeg.LinkEditDataCmd
		postParsing = handleLEData(realSeg.LinkEditDataCmd)
	case types.LC_NOTE:
		realSeg := cmd.(*macho.Note)
		data = realSeg.NoteCmd
		postParsing = handleUnused
	case types.LC_BUILD_VERSION:
		realSeg := cmd.(*macho.BuildVersion)
		data = *realSeg // .BuildVersionCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
		headerSize = size
	case types.LC_DYLD_EXPORTS_TRIE:
		realSeg := cmd.(*macho.DyldExportsTrie)
		data = realSeg.LinkEditDataCmd
		postParsing = handleLEData(realSeg.LinkEditDataCmd)
	case types.LC_DYLD_CHAINED_FIXUPS:
		realSeg := cmd.(*macho.DyldChainedFixups)
		data = realSeg.LinkEditDataCmd
		postParsing = handleLEData(realSeg.LinkEditDataCmd)
	case types.LC_FILESET_ENTRY:
		realSeg := cmd.(*macho.FilesetEntry)
		data = *realSeg // .FilesetEntryCmd // FIXME: technically should use the sub struct but it's nice to get the Name for free
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
