package macho

import (
	"fmt"
	"os"
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
	"github.com/blacktop/go-macho/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func addChild(parent *contracts.MemoryBlock, child *contracts.MemoryBlock) *contracts.MemoryBlock {
	parent.Content = append(parent.Content, child)
	return child
}

func addStructComplex(logger *logrus.Logger, parent *contracts.MemoryBlock, data interface{}, name string, offset, size uint64, banned []string) *contracts.MemoryBlock {
	val := parsingutils.GetDataValue(data)
	typ := val.Type()

	if size == 0 {
		size = uint64(typ.Size())
	}
	block := addChild(parent, &contracts.MemoryBlock{
		Name:         name,
		Address:      parent.Address + uintptr(offset),
		ParentOffset: offset,
		Size:         size,
	})

	if typ.Kind() != reflect.Struct {
		parsingutils.AddValue(block, "Value", val.Interface(), 0, uint8(typ.Size()), parsingutils.FormatValue)
		return block
	}

	fieldOffset := uint64(0)
	for _, field := range reflect.VisibleFields(typ) {
		fieldType := field.Type
		if (fieldType.Kind() == reflect.Struct && field.Anonymous) || !field.IsExported() || slices.Contains(banned, field.Name) {
			continue
		}
		size := uint8(fieldType.Size())
		parsingutils.AddValue(block, field.Name, val.FieldByIndex(field.Index).Interface(), fieldOffset, size, parsingutils.FormatValue)
		fieldOffset += uint64(size)
	}

	return block
}

func addStrucSimple(logger *logrus.Logger, parent *contracts.MemoryBlock, data interface{}, name string, offset uint64) *contracts.MemoryBlock {
	return addStructComplex(logger, parent, data, name, offset, 0, nil)
}

func Parse(logger *logrus.Logger, file string) (*contracts.MemoryBlock, error) {
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
	hdr := addStrucSimple(logger, root, m.FileHeader, "Header", 0)
	commands := addChild(root, &contracts.MemoryBlock{
		Name:         fmt.Sprintf("Commands (%d)", m.NCommands),
		Address:      root.Address + uintptr(hdr.Size),
		ParentOffset: uint64(hdr.Size),
		Size:         uint64(m.SizeCommands),
	})
	err = parsingutils.AddLinkWithBlock(hdr, "NCommands", commands, "gives amount")
	if err != nil {
		return nil, err
	}
	err = parsingutils.AddLinkWithBlock(hdr, "SizeCommands", commands, "gives size")
	if err != nil {
		return nil, err
	}

	offset := uint64(0)
	for i, cmd := range m.Loads {
		block, err := addCommand(logger, commands, i, cmd, offset)
		if err != nil {
			return nil, err
		}
		offset += uint64(block.Size)
	}

	return root, nil
}

func addCommand(logger *logrus.Logger, commands *contracts.MemoryBlock, i int, cmd macho.Load, offset uint64) (*contracts.MemoryBlock, error) {
	var data interface{}
	banned := []string{
		"LoadBytes",
	}

	switch cmd.Command() {
	case types.LC_REQ_DYLD:
		return nil, fmt.Errorf("LC_REQ_DYLD is not supported")
	case types.LC_SEGMENT, types.LC_SEGMENT_64:
		data = cmd.(*macho.Segment)
		banned = append(banned, "Firstsect", "ReaderAt")
	case types.LC_SYMTAB:
		data = cmd.(*macho.Symtab)
		banned = append(banned, "Syms")
	case types.LC_SYMSEG:
		data = cmd.(*macho.SymSeg)
	case types.LC_THREAD:
		data = cmd.(*macho.Thread)
	case types.LC_UNIXTHREAD:
		data = cmd.(*macho.UnixThread)
		banned = append(banned, "Threads")
	case types.LC_LOADFVMLIB:
		data = cmd.(*macho.LoadFvmlib)
	case types.LC_IDFVMLIB:
		data = cmd.(*macho.IDFvmlib)
	case types.LC_IDENT:
		data = cmd.(*macho.Ident)
	case types.LC_FVMFILE:
		data = cmd.(*macho.FvmFile)
	case types.LC_PREPAGE:
		data = cmd.(*macho.Prepage)
	case types.LC_DYSYMTAB:
		data = cmd.(*macho.Dysymtab)
		banned = append(banned, "IndirectSyms")
	case types.LC_LOAD_DYLIB:
		data = cmd.(*macho.LoadDylib)
	case types.LC_ID_DYLIB:
		data = cmd.(*macho.IDDylib)
	case types.LC_LOAD_DYLINKER:
		data = cmd.(*macho.LoadDylinker)
	case types.LC_ID_DYLINKER:
		data = cmd.(*macho.DylinkerID)
	case types.LC_PREBOUND_DYLIB:
		data = cmd.(*macho.PreboundDylib)
	case types.LC_ROUTINES:
		data = cmd.(*macho.Routines)
	case types.LC_SUB_FRAMEWORK:
		data = cmd.(*macho.SubFramework)
	case types.LC_SUB_UMBRELLA:
		data = cmd.(*macho.SubUmbrella)
	case types.LC_SUB_CLIENT:
		data = cmd.(*macho.SubClient)
	case types.LC_SUB_LIBRARY:
		data = cmd.(*macho.SubLibrary)
	case types.LC_TWOLEVEL_HINTS:
		data = cmd.(*macho.TwolevelHints)
	case types.LC_PREBIND_CKSUM:
		data = cmd.(*macho.PrebindCheckSum)
	case types.LC_LOAD_WEAK_DYLIB:
		data = cmd.(*macho.WeakDylib)
	case types.LC_ROUTINES_64:
		data = cmd.(*macho.Routines64)
	case types.LC_UUID:
		data = cmd.(*macho.UUID)
	case types.LC_RPATH:
		data = cmd.(*macho.Rpath)
	case types.LC_CODE_SIGNATURE:
		data = cmd.(*macho.CodeSignature)
		banned = append(banned,
			"CodeDirectories",
			"Requirements",
			"CMSSignature",
			"Entitlements",
			"EntitlementsDER",
			"Errors",
		)
	case types.LC_SEGMENT_SPLIT_INFO:
		data = cmd.(*macho.SplitInfo)
	case types.LC_REEXPORT_DYLIB:
		data = cmd.(*macho.ReExportDylib)
	case types.LC_LAZY_LOAD_DYLIB:
		data = cmd.(*macho.LazyLoadDylib)
	case types.LC_ENCRYPTION_INFO:
		data = cmd.(*macho.EncryptionInfo)
	case types.LC_DYLD_INFO:
		data = cmd.(*macho.DyldInfo)
	case types.LC_DYLD_INFO_ONLY:
		data = cmd.(*macho.DyldInfoOnly)
	case types.LC_LOAD_UPWARD_DYLIB:
		data = cmd.(*macho.UpwardDylib)
	case types.LC_VERSION_MIN_MACOSX, types.LC_VERSION_MIN_IPHONEOS, types.LC_VERSION_MIN_TVOS, types.LC_VERSION_MIN_WATCHOS:
		data = cmd.(*macho.VersionMin)
	case types.LC_FUNCTION_STARTS:
		data = cmd.(*macho.FunctionStarts)
	case types.LC_DYLD_ENVIRONMENT:
		data = cmd.(*macho.DyldEnvironment)
	case types.LC_MAIN:
		data = cmd.(*macho.EntryPoint)
	case types.LC_DATA_IN_CODE:
		data = cmd.(*macho.DataInCode)
	case types.LC_SOURCE_VERSION:
		data = cmd.(*macho.SourceVersion)
	case types.LC_DYLIB_CODE_SIGN_DRS:
		data = cmd.(*macho.DylibCodeSignDrs)
	case types.LC_ENCRYPTION_INFO_64:
		data = cmd.(*macho.EncryptionInfo64)
	case types.LC_LINKER_OPTION:
		data = cmd.(*macho.LinkerOption)
	case types.LC_LINKER_OPTIMIZATION_HINT:
		data = cmd.(*macho.LinkerOptimizationHint)
	case types.LC_NOTE:
		data = cmd.(*macho.Note)
	case types.LC_BUILD_VERSION:
		data = cmd.(*macho.BuildVersion)
		banned = append(banned, "Tools")
	case types.LC_DYLD_EXPORTS_TRIE:
		data = cmd.(*macho.DyldExportsTrie)
	case types.LC_DYLD_CHAINED_FIXUPS:
		data = cmd.(*macho.DyldChainedFixups)
	case types.LC_FILESET_ENTRY:
		data = cmd.(*macho.FilesetEntry)
	default:
		return nil, fmt.Errorf("unknown command %#x", cmd.Command())
	}

	cmdBlock := addStructComplex(logger, commands, data, fmt.Sprintf("Command %d: %s", i+1, cmd.Command()), offset, uint64(cmd.LoadSize()), banned)
	return cmdBlock, nil
}
