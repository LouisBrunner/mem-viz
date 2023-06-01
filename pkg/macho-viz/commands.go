package macho

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
	"github.com/blacktop/go-macho/types"
)

type contextData struct {
	text     *contracts.MemoryBlock
	linkEdit *contracts.MemoryBlock
}

func (me *parser) addCommand(root, commands *contracts.MemoryBlock, i int, cmd macho.Load, offset uint64, context *contextData) (*contracts.MemoryBlock, error) {
	var data interface{}
	banned := []string{
		"LoadBytes",
	}
	var postParsing func(command *contracts.MemoryBlock) error

	// FIXME: untestable
	handleObsolete := func(command *contracts.MemoryBlock) error {
		return fmt.Errorf("obsolete load command (%s), unsupported", command.Name)
	}

	// FIXME: untestable AND difficult to find
	handlePrivate := func(command *contracts.MemoryBlock) error {
		return fmt.Errorf("private load command (%s), unsupported", command.Name)
	}

	// FIXME: unused
	handleUnused := func(command *contracts.MemoryBlock) error {
		return fmt.Errorf("unusued load command (%s), currently unsupported", command.Name)
	}

	switch cmd.Command() {
	case types.LC_REQ_DYLD:
		return nil, fmt.Errorf("binary contains LC_REQ_DYLD which is not supported")
	case types.LC_SEGMENT, types.LC_SEGMENT_64:
		real := cmd.(*macho.Segment)
		postParsing = func(command *contracts.MemoryBlock) error {
			switch real.Name {
			case "__TEXT":
				context.text = command
			case "__LINKEDIT":
				context.linkEdit = command
			}
			if real.Offset == 0 && real.Filesz == 0 {
				return nil
			}
			me.addChild(root, &contracts.MemoryBlock{
				Name:    fmt.Sprintf("Segment (%s)", real.Name),
				Address: uintptr(real.Offset),
				Size:    uint64(real.Filesz),
			})
			// TODO: add sections
			// TODO: add detail of sections
			return nil
		}
		data = real
		banned = append(banned, "Firstsect", "ReaderAt")
	case types.LC_SYMTAB:
		real := cmd.(*macho.Symtab)
		data = real
		// TODO: add symbols
		banned = append(banned, "Syms")
	case types.LC_SYMSEG:
		real := cmd.(*macho.SymSeg)
		data = real
		postParsing = handleObsolete
	case types.LC_THREAD:
		real := cmd.(*macho.Thread)
		data = real
		// TODO: add registers
		banned = append(banned, "Threads")
	case types.LC_UNIXTHREAD:
		real := cmd.(*macho.UnixThread)
		data = real
		// TODO: add registers
		banned = append(banned, "Threads")
	case types.LC_LOADFVMLIB:
		real := cmd.(*macho.LoadFvmlib)
		data = real
		postParsing = handleObsolete
	case types.LC_IDFVMLIB:
		real := cmd.(*macho.IDFvmlib)
		data = real
		postParsing = handleObsolete
	case types.LC_IDENT:
		real := cmd.(*macho.Ident)
		data = real
		postParsing = handleObsolete
	case types.LC_FVMFILE:
		real := cmd.(*macho.FvmFile)
		data = real
		postParsing = handlePrivate
	case types.LC_PREPAGE:
		real := cmd.(*macho.Prepage)
		data = real
		postParsing = handlePrivate
	case types.LC_DYSYMTAB:
		real := cmd.(*macho.Dysymtab)
		data = real
		// TODO: add syms
		banned = append(banned, "IndirectSyms")
	case types.LC_LOAD_DYLIB:
		real := cmd.(*macho.LoadDylib)
		data = real
	case types.LC_ID_DYLIB:
		real := cmd.(*macho.IDDylib)
		data = real
	case types.LC_LOAD_DYLINKER:
		real := cmd.(*macho.LoadDylinker)
		data = real
	case types.LC_ID_DYLINKER:
		real := cmd.(*macho.DylinkerID)
		data = real
	case types.LC_PREBOUND_DYLIB:
		real := cmd.(*macho.PreboundDylib)
		data = real
		postParsing = handleUnused
	case types.LC_ROUTINES:
		real := cmd.(*macho.Routines)
		data = real
		postParsing = handleUnused
	case types.LC_SUB_FRAMEWORK:
		real := cmd.(*macho.SubFramework)
		data = real
	case types.LC_SUB_UMBRELLA:
		real := cmd.(*macho.SubUmbrella)
		data = real
	case types.LC_SUB_CLIENT:
		real := cmd.(*macho.SubClient)
		data = real
	case types.LC_SUB_LIBRARY:
		real := cmd.(*macho.SubLibrary)
		data = real
	case types.LC_TWOLEVEL_HINTS:
		real := cmd.(*macho.TwolevelHints)
		data = real
		postParsing = handleUnused
	case types.LC_PREBIND_CKSUM:
		real := cmd.(*macho.PrebindCheckSum)
		data = real
		postParsing = handleUnused
	case types.LC_LOAD_WEAK_DYLIB:
		real := cmd.(*macho.WeakDylib)
		data = real
	case types.LC_ROUTINES_64:
		real := cmd.(*macho.Routines64)
		data = real
		postParsing = handleUnused
	case types.LC_UUID:
		real := cmd.(*macho.UUID)
		data = real
	case types.LC_RPATH:
		real := cmd.(*macho.Rpath)
		data = real
	case types.LC_CODE_SIGNATURE:
		real := cmd.(*macho.CodeSignature)
		data = real
		// TODO: add detail
		banned = append(banned,
			"CodeDirectories",
			"Requirements",
			"CMSSignature",
			"Entitlements",
			"EntitlementsDER",
			"Errors",
		)
	case types.LC_SEGMENT_SPLIT_INFO:
		real := cmd.(*macho.SplitInfo)
		data = real
		// TODO: add links to rest
	case types.LC_REEXPORT_DYLIB:
		real := cmd.(*macho.ReExportDylib)
		data = real
	case types.LC_LAZY_LOAD_DYLIB:
		real := cmd.(*macho.LazyLoadDylib)
		data = real
	case types.LC_ENCRYPTION_INFO:
		real := cmd.(*macho.EncryptionInfo)
		data = real
		postParsing = handleUnused
	case types.LC_DYLD_INFO:
		real := cmd.(*macho.DyldInfo)
		data = real
		postParsing = handleUnused
	case types.LC_DYLD_INFO_ONLY:
		real := cmd.(*macho.DyldInfoOnly)
		data = real
		// TODO: add links
	case types.LC_LOAD_UPWARD_DYLIB:
		real := cmd.(*macho.UpwardDylib)
		data = real
	case types.LC_VERSION_MIN_MACOSX, types.LC_VERSION_MIN_IPHONEOS, types.LC_VERSION_MIN_TVOS, types.LC_VERSION_MIN_WATCHOS:
		real := cmd.(*macho.VersionMin)
		data = real
	case types.LC_FUNCTION_STARTS:
		real := cmd.(*macho.FunctionStarts)
		data = real
		// TODO: add links to rest
	case types.LC_DYLD_ENVIRONMENT:
		real := cmd.(*macho.DyldEnvironment)
		data = real
	case types.LC_MAIN:
		real := cmd.(*macho.EntryPoint)
		data = real
		postParsing = func(command *contracts.MemoryBlock) error {
			return parsingutils.AddLinkWithAddr(command, "EntryOffset", "points to", context.text.Address+uintptr(real.EntryOffset))
		}
	case types.LC_DATA_IN_CODE:
		real := cmd.(*macho.DataInCode)
		data = real
		// TODO: add links to rest
	case types.LC_SOURCE_VERSION:
		real := cmd.(*macho.SourceVersion)
		data = real
	case types.LC_DYLIB_CODE_SIGN_DRS:
		real := cmd.(*macho.DylibCodeSignDrs)
		data = real
		// TODO: add links to rest
	case types.LC_ENCRYPTION_INFO_64:
		real := cmd.(*macho.EncryptionInfo64)
		data = real
		postParsing = handleUnused
	case types.LC_LINKER_OPTION:
		real := cmd.(*macho.LinkerOption)
		data = real
		postParsing = handleUnused
	case types.LC_LINKER_OPTIMIZATION_HINT:
		real := cmd.(*macho.LinkerOptimizationHint)
		data = real
		// TODO: add links to rest
	case types.LC_NOTE:
		real := cmd.(*macho.Note)
		data = real
		postParsing = handleUnused
	case types.LC_BUILD_VERSION:
		real := cmd.(*macho.BuildVersion)
		data = real
	case types.LC_DYLD_EXPORTS_TRIE:
		real := cmd.(*macho.DyldExportsTrie)
		data = real
		// TODO: add links to rest
	case types.LC_DYLD_CHAINED_FIXUPS:
		real := cmd.(*macho.DyldChainedFixups)
		data = real
		// TODO: add links to rest
	case types.LC_FILESET_ENTRY:
		real := cmd.(*macho.FilesetEntry)
		data = real
		postParsing = handleUnused
	default:
		return nil, fmt.Errorf("unknown command %#x", cmd.Command())
	}

	size := uint64(cmd.LoadSize())
	cmdBlock := me.addStructDetailed(commands, data, fmt.Sprintf("Command %d: %s", i+1, cmd.Command()), offset, size, banned)
	if postParsing != nil {
		err := postParsing(cmdBlock)
		if err != nil {
			return nil, err
		}
	}
	return cmdBlock, nil
}
