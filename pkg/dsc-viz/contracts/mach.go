package contracts

import "fmt"

// From /Applications/Xcode.app/*/mach-o/loader.h

// The 32-bit mach header appears at the very beginning of the object file for
// 32-bit architectures.
type MachHeader struct {
	Magic      uint32 `struc:"little"` // mach magic number identifier
	CPUType    int32  `struc:"little"` // cpu specifier
	CPUSubType int32  `struc:"little"` // machine specifier
	FileType   uint32 `struc:"little"` // type of file
	NCmds      uint32 `struc:"little"` // number of load commands
	SizeOfCmds uint32 `struc:"little"` // the size of all the load commands
	Flags      uint32 `struc:"little"` // flags
}

// Constant for the magic field of the mach_header (32-bit architectures)
const (
	MH_MAGIC = 0xfeedface // the mach magic number
	MH_CIGAM = 0xcefaedfe // NXSwapInt(MH_MAGIC)
)

// The 64-bit mach header appears at the very beginning of object files for
// 64-bit architectures.
type MachHeader64 struct {
	Magic      uint32 `struc:"little"` // mach magic number identifier
	CPUType    int32  `struc:"little"` // cpu specifier
	CPUSubType int32  `struc:"little"` // machine specifier
	FileType   uint32 `struc:"little"` // type of file
	NCmds      uint32 `struc:"little"` // number of load commands
	SizeOfCmds uint32 `struc:"little"` // the size of all the load commands
	Flags      uint32 `struc:"little"` // flags
	Reserved   uint32 `struc:"little"` // reserved
}

// Constant for the magic field of the mach_header_64 (64-bit architectures)
const (
	MH_MAGIC_64 = 0xfeedfacf // the 64-bit mach magic number
	MH_CIGAM_64 = 0xcffaedfe // NXSwapInt(MH_MAGIC_64)
)

// The layout of the file depends on the filetype.  For all but the MH_OBJECT
// file type the segments are padded out and aligned on a segment alignment
// boundary for efficient demand pageing.  The MH_EXECUTE, MH_FVMLIB, MH_DYLIB,
// MH_DYLINKER and MH_BUNDLE file types also have the headers included as part
// of their first segment.
//
// The file type MH_OBJECT is a compact format intended as output of the
// assembler and input (and possibly output) of the link editor (the .o
// format).  All sections are in one unnamed segment with no segment padding.
// This format is used as an executable format when the file is so small the
// segment padding greatly increases its size.
//
// The file type MH_PRELOAD is an executable format intended for things that
// are not executed under the kernel (proms, stand alones, kernels, etc).  The
// format can be executed under the kernel but may demand paged it and not
// preload it before execution.
//
// A core file is in MH_CORE format and can be any in an arbritray legal
// Mach-O file.
//
// Constants for the filetype field of the mach_header
const (
	MH_OBJECT      = 0x1 // relocatable object file
	MH_EXECUTE     = 0x2 // demand paged executable file
	MH_FVMLIB      = 0x3 // fixed VM shared library file
	MH_CORE        = 0x4 // core file
	MH_PRELOAD     = 0x5 // preloaded executable file
	MH_DYLIB       = 0x6 // dynamically bound shared library
	MH_DYLINKER    = 0x7 // dynamic link editor
	MH_BUNDLE      = 0x8 // dynamically bound bundle file
	MH_DYLIB_STUB  = 0x9 // shared library stub for static linking only, no section contents
	MH_DSYM        = 0xa // companion file with only debug sections
	MH_KEXT_BUNDLE = 0xb // x86_64 kexts
	MH_FILESET     = 0xc // set of mach-o's
)

// Constants for the flags field of the mach_header
const (
	MH_NOUNDEFS                      = 0x1        // the object file has no undefined references
	MH_INCRLINK                      = 0x2        // the object file is the output of an incremental link against a base file and can't be link edited again
	MH_DYLDLINK                      = 0x4        // the object file is input for the dynamic linker and can't be staticly link edited again
	MH_BINDATLOAD                    = 0x8        // the object file's undefined references are bound by the dynamic linker when loaded.
	MH_PREBOUND                      = 0x10       // the file has its dynamic undefined references prebound.
	MH_SPLIT_SEGS                    = 0x20       // the file has its read-only and read-write segments split
	MH_LAZY_INIT                     = 0x40       // the shared library init routine is to be run lazily via catching memory faults to its writeable segments (obsolete)
	MH_TWOLEVEL                      = 0x80       // the image is using two-level name space bindings
	MH_FORCE_FLAT                    = 0x100      // the executable is forcing all images to use flat name space bindings
	MH_NOMULTIDEFS                   = 0x200      // this umbrella guarantees no multiple defintions of symbols in its sub-images so the two-level namespace hints can always be used.
	MH_NOFIXPREBINDING               = 0x400      // do not have dyld notify the prebinding agent about this executable
	MH_PREBINDABLE                   = 0x800      // the binary is not prebound but can have its prebinding redone. only used when MH_PREBOUND is not set.
	MH_ALLMODSBOUND                  = 0x1000     // indicates that this binary binds to all two-level namespace modules of its dependent libraries. only used when MH_PREBINDABLE and MH_TWOLEVEL are both set.
	MH_SUBSECTIONS_VIA_SYMBOLS       = 0x2000     // safe to divide up the sections into sub-sections via symbols for dead code stripping
	MH_CANONICAL                     = 0x4000     // the binary has been canonicalized via the unprebind operation
	MH_WEAK_DEFINES                  = 0x8000     // the final linked image contains external weak symbols
	MH_BINDS_TO_WEAK                 = 0x10000    // the final linked image uses weak symbols
	MH_ALLOW_STACK_EXECUTION         = 0x20000    // When this bit is set, all stacks in the task will be given stack execution privilege.  Only used in MH_EXECUTE filetypes.
	MH_ROOT_SAFE                     = 0x40000    // When this bit is set, the binary declares it is safe for use in processes with uid zero
	MH_SETUID_SAFE                   = 0x80000    // When this bit is set, the binary declares it is safe for use in processes when issetugid() is true
	MH_NO_REEXPORTED_DYLIBS          = 0x100000   // When this bit is set on a dylib, the static linker does not need to examine dependent dylibs to see if any are re-exported
	MH_PIE                           = 0x200000   // When this bit is set, the OS will load the main executable at a random address.  Only used in MH_EXECUTE filetypes.
	MH_DEAD_STRIPPABLE_DYLIB         = 0x400000   // Only for use on dylibs.  When linking against a dylib that has this bit set, the static linker will automatically not create a LC_LOAD_DYLIB load command to the dylib if no symbols are being referenced from the dylib.
	MH_HAS_TLV_DESCRIPTORS           = 0x800000   // Contains a section of type S_THREAD_LOCAL_VARIABLES
	MH_NO_HEAP_EXECUTION             = 0x1000000  // When this bit is set, the OS will run the main executable with a non-executable heap even on platforms (e.g. i386) that don't require it. Only used in MH_EXECUTE filetypes.
	MH_APP_EXTENSION_SAFE            = 0x02000000 // The code was linked for use in an application extension.
	MH_NLIST_OUTOFSYNC_WITH_DYLDINFO = 0x04000000 // The external symbols listed in the nlist symbol table do not include all the symbols listed in the dyld info.
	MH_SIM_SUPPORT                   = 0x08000000 // Allow LC_MIN_VERSION_MACOS and LC_BUILD_VERSION load commands with the platforms macOS, iOSMac, iOSSimulator, tvOSSimulator and watchOSSimulator.
	MH_DYLIB_IN_CACHE                = 0x80000000 // Only for use on dylibs. When this bit is set, the dylib is part of the dyld shared cache, rather than loose in the filesystem.
)

// The load commands directly follow the mach_header.  The total size of all
// of the commands is given by the sizeofcmds field in the mach_header.  All
// load commands must have as their first two fields cmd and cmdsize.  The cmd
// field is filled in with a constant for that command type.  Each command type
// has a structure specifically for it.  The cmdsize field is the size in bytes
// of the particular load command structure plus anything that follows it that
// is a part of the load command (i.e. section structures, strings, etc.).  To
// advance to the next load command the cmdsize can be added to the offset or
// pointer of the current load command.  The cmdsize for 32-bit architectures
// MUST be a multiple of 4 bytes and for 64-bit architectures MUST be a multiple
// of 8 bytes (these are forever the maximum alignment of any load commands).
// The padded bytes must be zero.  All tables in the object file must also
// follow these rules so the file can be memory mapped.  Otherwise the pointers
// to these tables will not work well or at all on some machines.  With all
// padding zeroed like objects will compare byte for byte.
type LoadCommand struct {
	Cmd     uint32 `struc:"little"` // type of load command
	CmdSize uint32 `struc:"little"` // total size of command in bytes
}

// After MacOS X 10.1 when a new load command is added that is required to be
// understood by the dynamic linker for the image to execute properly the
// LC_REQ_DYLD bit will be or'ed into the load command constant.  If the dynamic
// linker sees such a load command it it does not understand will issue a
// "unknown load command required for execution" error and refuse to use the
// image.  Other load commands without this bit that are not understood will
// simply be ignored.
const (
	LC_REQ_DYLD = 0x80000000
)

// Constants for the cmd field of all load commands, the type
const (
	LC_SEGMENT                  = 0x1                  // segment of this file to be mapped
	LC_SYMTAB                   = 0x2                  // link-edit stab symbol table info
	LC_SYMSEG                   = 0x3                  // link-edit gdb symbol table info (obsolete)
	LC_THREAD                   = 0x4                  // thread
	LC_UNIXTHREAD               = 0x5                  // unix thread (includes a stack)
	LC_LOADFVMLIB               = 0x6                  // load a specified fixed VM shared library
	LC_IDFVMLIB                 = 0x7                  // fixed VM shared library identification
	LC_IDENT                    = 0x8                  // object identification info (obsolete)
	LC_FVMFILE                  = 0x9                  // fixed VM file inclusion (internal use)
	LC_PREPAGE                  = 0xa                  // prepage command (internal use)
	LC_DYSYMTAB                 = 0xb                  // dynamic link-edit symbol table info
	LC_LOAD_DYLIB               = 0xc                  // load a dynamically linked shared library
	LC_ID_DYLIB                 = 0xd                  // dynamically linked shared lib ident
	LC_LOAD_DYLINKER            = 0xe                  // load a dynamic linker
	LC_ID_DYLINKER              = 0xf                  // dynamic linker identification
	LC_PREBOUND_DYLIB           = 0x10                 // modules prebound for a dynamically linked shared library
	LC_ROUTINES                 = 0x11                 // image routines
	LC_SUB_FRAMEWORK            = 0x12                 // sub framework
	LC_SUB_UMBRELLA             = 0x13                 // sub umbrella
	LC_SUB_CLIENT               = 0x14                 // sub client
	LC_SUB_LIBRARY              = 0x15                 // sub library
	LC_TWOLEVEL_HINTS           = 0x16                 // two-level namespace lookup hints
	LC_PREBIND_CKSUM            = 0x17                 // prebind checksum
	LC_LOAD_WEAK_DYLIB          = (0x18 | LC_REQ_DYLD) // load a dynamically linked shared library that is allowed to be missing (all symbols are weak imported).
	LC_SEGMENT_64               = 0x19                 // 64-bit segment of this file to be mapped
	LC_ROUTINES_64              = 0x1a                 // 64-bit image routines
	LC_UUID                     = 0x1b                 // the uuid
	LC_RPATH                    = (0x1c | LC_REQ_DYLD) // runpath additions
	LC_CODE_SIGNATURE           = 0x1d                 // local of code signature
	LC_SEGMENT_SPLIT_INFO       = 0x1e                 // local of info to split segments
	LC_REEXPORT_DYLIB           = (0x1f | LC_REQ_DYLD) // load and re-export dylib
	LC_LAZY_LOAD_DYLIB          = 0x20                 // delay load of dylib until first use
	LC_ENCRYPTION_INFO          = 0x21                 // encrypted segment information
	LC_DYLD_INFO                = 0x22                 // compressed dyld information
	LC_DYLD_INFO_ONLY           = (0x22 | LC_REQ_DYLD) // compressed dyld information only
	LC_LOAD_UPWARD_DYLIB        = (0x23 | LC_REQ_DYLD) // load upward dylib
	LC_VERSION_MIN_MACOSX       = 0x24                 // build for MacOSX min OS version
	LC_VERSION_MIN_IPHONEOS     = 0x25                 // build for iPhoneOS min OS version
	LC_FUNCTION_STARTS          = 0x26                 // compressed table of function start addresses
	LC_DYLD_ENVIRONMENT         = 0x27                 // string for dyld to treat like environment variable
	LC_MAIN                     = (0x28 | LC_REQ_DYLD) // replacement for LC_UNIXTHREAD
	LC_DATA_IN_CODE             = 0x29                 // table of non-instructions in __text
	LC_SOURCE_VERSION           = 0x2A                 // source version used to build binary
	LC_DYLIB_CODE_SIGN_DRS      = 0x2B                 // Code signing DRs copied from linked dylibs
	LC_ENCRYPTION_INFO_64       = 0x2C                 // 64-bit encrypted segment information
	LC_LINKER_OPTION            = 0x2D                 // linker options in MH_OBJECT files
	LC_LINKER_OPTIMIZATION_HINT = 0x2E                 // optimization hints in MH_OBJECT files
	LC_VERSION_MIN_TVOS         = 0x2F                 // build for AppleTV min OS version
	LC_VERSION_MIN_WATCHOS      = 0x30                 // build for Watch min OS version
	LC_NOTE                     = 0x31                 // arbitrary data included within a Mach-O file
	LC_BUILD_VERSION            = 0x32                 // build for platform min OS version
	LC_DYLD_EXPORTS_TRIE        = (0x33 | LC_REQ_DYLD) // used with linkedit_data_command, payload is trie
	LC_DYLD_CHAINED_FIXUPS      = (0x34 | LC_REQ_DYLD) // used with linkedit_data_command
	LC_FILESET_ENTRY            = (0x35 | LC_REQ_DYLD) // used with fileset_entry_command
)

func LC2String(lc uint32) string {
	switch lc {
	case LC_SEGMENT:
		return "SEGMENT"
	case LC_SYMTAB:
		return "SYMTAB"
	case LC_SYMSEG:
		return "SYMSEG"
	case LC_THREAD:
		return "THREAD"
	case LC_UNIXTHREAD:
		return "UNIXTHREAD"
	case LC_LOADFVMLIB:
		return "LOADFVMLIB"
	case LC_IDFVMLIB:
		return "IDFVMLIB"
	case LC_IDENT:
		return "IDENT"
	case LC_FVMFILE:
		return "FVMFILE"
	case LC_PREPAGE:
		return "PREPAGE"
	case LC_DYSYMTAB:
		return "DYSYMTAB"
	case LC_LOAD_DYLIB:
		return "LOAD_DYLIB"
	case LC_ID_DYLIB:
		return "ID_DYLIB"
	case LC_LOAD_DYLINKER:
		return "LOAD_DYLINKER"
	case LC_ID_DYLINKER:
		return "ID_DYLINKER"
	case LC_PREBOUND_DYLIB:
		return "PREBOUND_DYLIB"
	case LC_ROUTINES:
		return "ROUTINES"
	case LC_SUB_FRAMEWORK:
		return "SUB_FRAMEWORK"
	case LC_SUB_UMBRELLA:
		return "SUB_UMBRELLA"
	case LC_SUB_CLIENT:
		return "SUB_CLIENT"
	case LC_SUB_LIBRARY:
		return "SUB_LIBRARY"
	case LC_TWOLEVEL_HINTS:
		return "TWOLEVEL_HINTS"
	case LC_PREBIND_CKSUM:
		return "PREBIND_CKSUM"
	case LC_LOAD_WEAK_DYLIB:
		return "LOAD_WEAK_DYLIB"
	case LC_SEGMENT_64:
		return "SEGMENT_64"
	case LC_ROUTINES_64:
		return "ROUTINES_64"
	case LC_UUID:
		return "UUID"
	case LC_RPATH:
		return "RPATH"
	case LC_CODE_SIGNATURE:
		return "CODE_SIGNATURE"
	case LC_SEGMENT_SPLIT_INFO:
		return "SEGMENT_SPLIT_INFO"
	case LC_REEXPORT_DYLIB:
		return "REEXPORT_DYLIB"
	case LC_LAZY_LOAD_DYLIB:
		return "LAZY_LOAD_DYLIB"
	case LC_ENCRYPTION_INFO:
		return "ENCRYPTION_INFO"
	case LC_DYLD_INFO:
		return "DYLD_INFO"
	case LC_DYLD_INFO_ONLY:
		return "DYLD_INFO_ONLY"
	case LC_LOAD_UPWARD_DYLIB:
		return "LOAD_UPWARD_DYLIB"
	case LC_VERSION_MIN_MACOSX:
		return "VERSION_MIN_MACOSX"
	case LC_VERSION_MIN_IPHONEOS:
		return "VERSION_MIN_IPHONEOS"
	case LC_FUNCTION_STARTS:
		return "FUNCTION_STARTS"
	case LC_DYLD_ENVIRONMENT:
		return "DYLD_ENVIRONMENT"
	case LC_MAIN:
		return "MAIN"
	case LC_DATA_IN_CODE:
		return "DATA_IN_CODE"
	case LC_SOURCE_VERSION:
		return "SOURCE_VERSION"
	case LC_DYLIB_CODE_SIGN_DRS:
		return "DYLIB_CODE_SIGN_DRS"
	case LC_ENCRYPTION_INFO_64:
		return "ENCRYPTION_INFO_64"
	case LC_LINKER_OPTION:
		return "LINKER_OPTION"
	case LC_LINKER_OPTIMIZATION_HINT:
		return "LINKER_OPTIMIZATION_HINT"
	case LC_VERSION_MIN_TVOS:
		return "VERSION_MIN_TVOS"
	case LC_VERSION_MIN_WATCHOS:
		return "VERSION_MIN_WATCHOS"
	case LC_NOTE:
		return "NOTE"
	case LC_BUILD_VERSION:
		return "BUILD_VERSION"
	case LC_DYLD_EXPORTS_TRIE:
		return "DYLD_EXPORTS_TRIE"
	case LC_DYLD_CHAINED_FIXUPS:
		return "DYLD_CHAINED_FIXUPS"
	case LC_FILESET_ENTRY:
		return "FILESET_ENTRY"
	default:
		return fmt.Sprintf("Unknown-%#x", lc)
	}
}

// The segment load command indicates that a part of this file is to be
// mapped into the task's address space.  The size of this segment in memory,
// vmsize, maybe equal to or larger than the amount to map from this file,
// filesize.  The file is mapped starting at fileoff to the beginning of
// the segment in memory, vmaddr.  The rest of the memory of the segment,
// if any, is allocated zero fill on demand.  The segment's maximum virtual
// memory protection and initial virtual memory protection are specified
// by the maxprot and initprot fields.  If the segment has sections then the
// section structures directly follow the segment command and their size is
// reflected in cmdsize.
type SegmentCommand struct { // for 32-bit architectures
	Cmd      uint32   `struc:"little"` // LC_SEGMENT
	CmdSize  uint32   `struc:"little"` // includes sizeof section structs
	SegName  [16]byte // segment name
	VMAddr   uint32   `struc:"little"` // memory address of this segment
	VMSize   uint32   `struc:"little"` // memory size of this segment
	FileOff  uint32   `struc:"little"` // file offset of this segment
	FileSize uint32   `struc:"little"` // amount to map from the file
	MaxProt  int32    `struc:"little"` // maximum VM protection
	InitProt int32    `struc:"little"` // initial VM protection
	NSects   uint32   `struc:"little"` // number of sections in segment
	Flags    uint32   `struc:"little"` // flags
}

// The 64-bit segment load command indicates that a part of this file is to be
// mapped into a 64-bit task's address space.  If the 64-bit segment has
// sections then section64 structures directly follow the 64-bit segment
// command and their size is reflected in cmdsize.
type SegmentCommand64 struct { // for 64-bit architectures
	Cmd      uint32   `struc:"little"` // LC_SEGMENT_64
	CmdSize  uint32   `struc:"little"` // includes sizeof section64 structs
	SegName  [16]byte // segment name
	VMAddr   uint64   `struc:"little"` // memory address of this segment
	VMSize   uint64   `struc:"little"` // memory size of this segment
	FileOff  uint64   `struc:"little"` // file offset of this segment
	FileSize uint64   `struc:"little"` // amount to map from the file
	MaxProt  int32    `struc:"little"` // maximum VM protection
	InitProt int32    `struc:"little"` // initial VM protection
	NSects   uint32   `struc:"little"` // number of sections in segment
	Flags    uint32   `struc:"little"` // flags
}

// Constants for the flags field of the segment_command
const (
	SG_HIGHVM              = 0x1  // the file contents for this segment is for the high part of the VM space, the low part is zero filled (for stacks in core files)
	SG_FVMLIB              = 0x2  // this segment is the VM that is allocated by a fixed VM library, for overlap checking in the link editor
	SG_NORELOC             = 0x4  // this segment has nothing that was relocated in it and nothing relocated to it, that is it maybe safely replaced without relocation
	SG_PROTECTED_VERSION_1 = 0x8  // This segment is protected.  If the segment starts at file offset 0, the first page of the segment is not protected.  All other pages of the segment are protected.
	SG_READ_ONLY           = 0x10 // This segment is made read-only after fixups
)

// A segment is made up of zero or more sections.  Non-MH_OBJECT files have
// all of their segments with the proper sections in each, and padded to the
// specified segment alignment when produced by the link editor.  The first
// segment of a MH_EXECUTE and MH_FVMLIB format file contains the mach_header
// and load commands of the object file before its first section.  The zero
// fill sections are always last in their segment (in all formats).  This
// allows the zeroed segment padding to be mapped into memory where zero fill
// sections might be. The gigabyte zero fill sections, those with the section
// type S_GB_ZEROFILL, can only be in a segment with sections of this type.
// These segments are then placed after all other segments.
//
// The MH_OBJECT format has all of its sections in one segment for
// compactness.  There is no padding to a specified segment boundary and the
// mach_header and load commands are not part of the segment.
//
// Sections with the same section name, sectname, going into the same segment,
// segname, are combined by the link editor.  The resulting section is aligned
// to the maximum alignment of the combined sections and is the new section's
// alignment.  The combined sections are aligned to their original alignment in
// the combined section.  Any padded bytes to get the specified alignment are
// zeroed.
//
// The format of the relocation entries referenced by the reloff and nreloc
// fields of the section structure for mach object files is described in the
// header file <reloc.h>.
type Section struct { // for 32-bit architectures
	SectName  [16]byte // name of this section
	SegName   [16]byte // segment this section goes in
	Addr      uint32   `struc:"little"` // memory address of this section
	Size      uint32   `struc:"little"` // size in bytes of this section
	Offset    uint32   `struc:"little"` // file offset of this section
	Align     uint32   `struc:"little"` // section alignment (power of 2)
	RelOff    uint32   `struc:"little"` // file offset of relocation entries
	NReloc    uint32   `struc:"little"` // number of relocation entries
	Flags     uint32   `struc:"little"` // flags (section type and attributes)
	Reserved1 uint32   `struc:"little"` // reserved (for offset or index)
	Reserved2 uint32   `struc:"little"` // reserved (for count or sizeof)
}

type Section64 struct { // for 64-bit architectures
	SectName  [16]byte // name of this section
	SegName   [16]byte // segment this section goes in
	Addr      uint64   `struc:"little"` // memory address of this section
	Size      uint64   `struc:"little"` // size in bytes of this section
	Offset    uint32   `struc:"little"` // file offset of this section
	Align     uint32   `struc:"little"` // section alignment (power of 2)
	RelOff    uint32   `struc:"little"` // file offset of relocation entries
	NReloc    uint32   `struc:"little"` // number of relocation entries
	Flags     uint32   `struc:"little"` // flags (section type and attributes)
	Reserved1 uint32   `struc:"little"` // reserved (for offset or index)
	Reserved2 uint32   `struc:"little"` // reserved (for count or sizeof)
	Reserved3 uint32   `struc:"little"` // reserved
}

// The flags field of a section structure is separated into two parts a section
// type and section attributes.  The section types are mutually exclusive (it
// can only have one type) but the section attributes are not (it may have more
// than one attribute).
const (
	SECTION_TYPE       = 0x000000ff // 256 section types
	SECTION_ATTRIBUTES = 0xffffff00 //  24 section attributes
)

// Constants for the type of a section
const (
	S_REGULAR          = 0x0 // regular section
	S_ZEROFILL         = 0x1 // zero fill on demand section
	S_CSTRING_LITERALS = 0x2 // section with only literal C strings
	S_4BYTE_LITERALS   = 0x3 // section with only 4 byte literals
	S_8BYTE_LITERALS   = 0x4 // section with only 8 byte literals
	S_LITERAL_POINTERS = 0x5 // section with only pointers to literals
)

// For the two types of symbol pointers sections and the symbol stubs section
// they have indirect symbol table entries.  For each of the entries in the
// section the indirect symbol table entries, in corresponding order in the
// indirect symbol table, start at the index stored in the reserved1 field
// of the section structure.  Since the indirect symbol table entries
// correspond to the entries in the section the number of indirect symbol table
// entries is inferred from the size of the section divided by the size of the
// entries in the section.  For symbol pointers sections the size of the entries
// in the section is 4 bytes and for symbol stubs sections the byte size of the
// stubs is stored in the reserved2 field of the section structure.
const (
	S_NON_LAZY_SYMBOL_POINTERS   = 0x6  // section with only non-lazy symbol pointers
	S_LAZY_SYMBOL_POINTERS       = 0x7  // section with only lazy symbol pointers
	S_SYMBOL_STUBS               = 0x8  // section with only symbol stubs, byte size of stub in the reserved2 field
	S_MOD_INIT_FUNC_POINTERS     = 0x9  // section with only function pointers for initialization
	S_MOD_TERM_FUNC_POINTERS     = 0xa  // section with only function pointers for termination
	S_COALESCED                  = 0xb  // section contains symbols that are to be coalesced
	S_GB_ZEROFILL                = 0xc  // zero fill on demand section (that can be larger than 4 gigabytes)
	S_INTERPOSING                = 0xd  // section with only pairs of function pointers for interposing
	S_16BYTE_LITERALS            = 0xe  // section with only 16 byte literals
	S_DTRACE_DOF                 = 0xf  // section contains DTrace Object Format
	S_LAZY_DYLIB_SYMBOL_POINTERS = 0x10 // section with only lazy symbol pointers to lazy loaded dylibs
)

// Section types to support thread local variables
const (
	S_THREAD_LOCAL_REGULAR                = 0x11 // template of initial values for TLVs
	S_THREAD_LOCAL_ZEROFILL               = 0x12 // template of initial values for TLVs
	S_THREAD_LOCAL_VARIABLES              = 0x13 // TLV descriptors
	S_THREAD_LOCAL_VARIABLE_POINTERS      = 0x14 // pointers to TLV descriptors
	S_THREAD_LOCAL_INIT_FUNCTION_POINTERS = 0x15 // functions to call to initialize TLV values
	S_INIT_FUNC_OFFSETS                   = 0x16 // 32-bit offsets to  initializers
)

// Constants for the section attributes part of the flags field of a section
// structure.
const (
	SECTION_ATTRIBUTES_USR     = 0xff000000 // User setable attributes
	S_ATTR_PURE_INSTRUCTIONS   = 0x80000000 // section contains only true machine instructions
	S_ATTR_NO_TOC              = 0x40000000 // section contains coalesced symbols that are not to be in a ranlib table of contents
	S_ATTR_STRIP_STATIC_SYMS   = 0x20000000 // ok to strip static symbols in this section in files with the MH_DYLDLINK flag
	S_ATTR_NO_DEAD_STRIP       = 0x10000000 // no dead stripping
	S_ATTR_LIVE_SUPPORT        = 0x08000000 // blocks are live if they reference live blocks
	S_ATTR_SELF_MODIFYING_CODE = 0x04000000 // Used with i386 code stubs written on by dyld
	// If a segment contains any sections marked with S_ATTR_DEBUG then all
	// sections in that segment must have this attribute.  No section other than
	// a section marked with this attribute may reference the contents of this
	// section.  A section with this attribute may contain no symbols and must have
	// a section type S_REGULAR.  The static linker will not copy section contents
	// from sections with this attribute into its output file.  These sections
	// generally contain DWARF debugging info.
	S_ATTR_DEBUG             = 0x02000000 // a debug section
	SECTION_ATTRIBUTES_SYS   = 0x00ffff00 // system setable attributes
	S_ATTR_SOME_INSTRUCTIONS = 0x00000400 // section contains some machine instructions
	S_ATTR_EXT_RELOC         = 0x00000200 // section has external relocation entries
	S_ATTR_LOC_RELOC         = 0x00000100 // section has local relocation entries
)

// The names of segments and sections in them are mostly meaningless to the
// link-editor.  But there are few things to support traditional UNIX
// executables that require the link-editor and assembler to use some names
// agreed upon by convention.
//
// The initial protection of the "__TEXT" segment has write protection turned
// off (not writeable).
//
// The link-editor will allocate common symbols at the end of the "__common"
// section in the "__DATA" segment.  It will create the section and segment
// if needed.
//
// The currently known segment names and the section names in those segments
const (
	SEG_PAGEZERO      = "__PAGEZERO"      // the pagezero segment which has no protections and catches NULL references for MH_EXECUTE files
	SEG_TEXT          = "__TEXT"          // the tradition UNIX text segment
	SECT_TEXT         = "__text"          // the real text part of the text section no headers, and no padding
	SECT_FVMLIB_INIT0 = "__fvmlib_init0"  // the fvmlib initialization section
	SECT_FVMLIB_INIT1 = "__fvmlib_init1"  // the section following the fvmlib initialization section
	SEG_DATA          = "__DATA"          // the tradition UNIX data segment
	SECT_DATA         = "__data"          // the real initialized data section no padding, no bss overlap
	SECT_BSS          = "__bss"           // the real uninitialized data section no padding
	SECT_COMMON       = "__common"        // the section common symbols are allocated in by the link editor
	SEG_OBJC          = "__OBJC"          // objective-C runtime segment
	SECT_OBJC_SYMBOLS = "__symbol_table"  // symbol table
	SECT_OBJC_MODULES = "__module_info"   // module information
	SECT_OBJC_STRINGS = "__selector_strs" // string table
	SECT_OBJC_REFS    = "__selector_refs" // string table
	SEG_ICON          = "__ICON"          // the icon segment
	SECT_ICON_HEADER  = "__header"        // the icon headers
	SECT_ICON_TIFF    = "__tiff"          // the icons in tiff format
	SEG_LINKEDIT      = "__LINKEDIT"      // the segment containing all structs created and maintained by the link editor.  Created with -seglinkedit option to ld(1) for MH_EXECUTE and FVMLIB file types only
	SEG_LINKINFO      = "__LINKINFO"      // the segment overlapping with linkedit containing linking information
	SEG_UNIXSTACK     = "__UNIXSTACK"     // the unix stack segment
	SEG_IMPORT        = "__IMPORT"        // the segment for the self (dyld) modifing code stubs that has read, write and execute permissions
)

// Fixed virtual memory shared libraries are identified by two things.  The
// target pathname (the name of the library as found for execution), and the
// minor version number.  The address of where the headers are loaded is in
// header_addr. (THIS IS OBSOLETE and no longer supported).
type FVMLib struct {
	Name         uint32 `struc:"little"` // library's target pathname
	MinorVersion uint32 `struc:"little"` // library's minor version number
	HeaderAddr   uint32 `struc:"little"` // library's header address
}

// A fixed virtual shared library (filetype == MH_FVMLIB in the mach header)
// contains a fvmlib_command (cmd == LC_IDFVMLIB) to identify the library.
// An object that uses a fixed virtual shared library also contains a
// fvmlib_command (cmd == LC_LOADFVMLIB) for each library it uses.
// (THIS IS OBSOLETE and no longer supported).
type FVMLibCommand struct {
	Cmd     uint32 `struc:"little"` // LC_IDFVMLIB or LC_LOADFVMLIB
	CmdSize uint32 `struc:"little"` // includes pathname string
	FVMLib  FVMLib // the library identification
}

// Dynamicly linked shared libraries are identified by two things.  The
// pathname (the name of the library as found for execution), and the
// compatibility version number.  The pathname must match and the compatibility
// number in the user of the library must be greater than or equal to the
// library being used.  The time stamp is used to record the time a library was
// built and copied into user so it can be use to determined if the library used
// at runtime is exactly the same as used to built the program.
type DYLIB struct {
	Name                 uint32 `struc:"little"` // library's path name
	Timestamp            uint32 `struc:"little"` // library's build time stamp
	CurrentVersion       uint32 `struc:"little"` // library's current version number
	CompatibilityVersion uint32 `struc:"little"` // library's compatibility vers number
}

// A dynamically linked shared library (filetype == MH_DYLIB in the mach header)
// contains a dylib_command (cmd == LC_ID_DYLIB) to identify the library.
// An object that uses a dynamically linked shared library also contains a
// dylib_command (cmd == LC_LOAD_DYLIB, LC_LOAD_WEAK_DYLIB, or
// LC_REEXPORT_DYLIB) for each library it uses.
type DYLIBCommand struct {
	Cmd     uint32 `struc:"little"` // LC_ID_DYLIB, LC_LOAD_{,WEAK_}DYLIB, LC_REEXPORT_DYLIB
	CmdSize uint32 `struc:"little"` // includes pathname string
	DYLIB   DYLIB  // the library identification
}

// A dynamically linked shared library may be a subframework of an umbrella
// framework.  If so it will be linked with "-umbrella umbrella_name" where
// Where "umbrella_name" is the name of the umbrella framework. A subframework
// can only be linked against by its umbrella framework or other subframeworks
// that are part of the same umbrella framework.  Otherwise the static link
// editor produces an error and states to link against the umbrella framework.
// The name of the umbrella framework for subframeworks is recorded in the
// following structure.
type SubFrameworkCommand struct {
	Cmd      uint32 `struc:"little"` // LC_SUB_FRAMEWORK
	CmdSize  uint32 `struc:"little"` // includes umbrella string
	Umbrella uint32 `struc:"little"` // the umbrella framework name
}

// For dynamically linked shared libraries that are subframework of an umbrella
// framework they can allow clients other than the umbrella framework or other
// subframeworks in the same umbrella framework.  To do this the subframework
// is built with "-allowable_client client_name" and an LC_SUB_CLIENT load
// command is created for each -allowable_client flag.  The client_name is
// usually a framework name.  It can also be a name used for bundles clients
// where the bundle is built with "-client_name client_name".
type SubClientCommand struct {
	Cmd     uint32 `struc:"little"` // LC_SUB_CLIENT
	CmdSize uint32 `struc:"little"` // includes client string
	Client  uint32 `struc:"little"` // the client name
}

// A dynamically linked shared library may be a sub_umbrella of an umbrella
// framework.  If so it will be linked with "-sub_umbrella umbrella_name" where
// Where "umbrella_name" is the name of the sub_umbrella framework.  When
// staticly linking when -twolevel_namespace is in effect a twolevel namespace
// umbrella framework will only cause its subframeworks and those frameworks
// listed as sub_umbrella frameworks to be implicited linked in.  Any other
// dependent dynamic libraries will not be linked it when -twolevel_namespace
// is in effect.  The primary library recorded by the static linker when
// resolving a symbol in these libraries will be the umbrella framework.
// Zero or more sub_umbrella frameworks may be use by an umbrella framework.
// The name of a sub_umbrella framework is recorded in the following structure.
type SubUmbrellaCommand struct {
	Cmd         uint32 `struc:"little"` // LC_SUB_UMBRELLA
	CmdSize     uint32 `struc:"little"` // includes sub_umbrella string
	SubUmbrella uint32 `struc:"little"` // the sub_umbrella framework name
}

// A dynamically linked shared library may be a sub_library of another shared
// library.  If so it will be linked with "-sub_library library_name" where
// Where "library_name" is the name of the sub_library shared library.  When
// staticly linking when -twolevel_namespace is in effect a twolevel namespace
// shared library will only cause its subframeworks and those frameworks
// listed as sub_umbrella frameworks and libraries listed as sub_libraries to
// be implicited linked in.  Any other dependent dynamic libraries will not be
// linked it when -twolevel_namespace is in effect.  The primary library
// recorded by the static linker when resolving a symbol in these libraries
// will be the umbrella framework (or dynamic library). Zero or more sub_library
// shared libraries may be use by an umbrella framework or (or dynamic library).
// The name of a sub_library framework is recorded in the following structure.
// For example /usr/lib/libobjc_profile.A.dylib would be recorded as "libobjc".
type SubLibraryCommand struct {
	Cmd        uint32 `struc:"little"` // LC_SUB_LIBRARY
	CmdSize    uint32 `struc:"little"` // includes sub_library string
	SubLibrary uint32 `struc:"little"` // the sub_library name
}

// A program (filetype == MH_EXECUTE) that is
// prebound to its dynamic libraries has one of these for each library that
// the static linker used in prebinding.  It contains a bit vector for the
// modules in the library.  The bits indicate which modules are bound (1) and
// which are not (0) from the library.  The bit for module 0 is the low bit
// of the first byte.  So the bit for the Nth module is:
// (linked_modules[N/8] >> N%8) & 1
type PreboundDylibCommand struct {
	Cmd           uint32 `struc:"little"` // LC_PREBOUND_DYLIB
	CmdSize       uint32 `struc:"little"` // includes strings
	Name          uint32 `struc:"little"` // library's path name
	NModules      uint32 `struc:"little"` // number of modules in library
	LinkedModules uint32 `struc:"little"` // bit vector of linked modules
}

// A program that uses a dynamic linker contains a dylinker_command to identify
// the name of the dynamic linker (LC_LOAD_DYLINKER).  And a dynamic linker
// contains a dylinker_command to identify the dynamic linker (LC_ID_DYLINKER).
// A file can have at most one of these.
// This struct is also used for the LC_DYLD_ENVIRONMENT load command and
// contains string for dyld to treat like environment variable.
type DylinkerCommand struct {
	Cmd     uint32 `struc:"little"` // LC_ID_DYLINKER, LC_LOAD_DYLINKER or LC_DYLD_ENVIRONMENT
	CmdSize uint32 `struc:"little"` // includes pathname string
	Name    uint32 `struc:"little"` // dynamic linker's path name
}

// Thread commands contain machine-specific data structures suitable for
// use in the thread state primitives.  The machine specific data structures
// follow the struct thread_command as follows.
// Each flavor of machine specific data structure is preceded by an uint32
// constant for the flavor of that data structure, an uint32 that is the
// count of uint32's of the size of the state data structure and then
// the state data structure follows.  This triple may be repeated for many
// flavors.  The constants for the flavors, counts and state data structure
// definitions are expected to be in the header file <machine/thread_status.h>.
// These machine specific data structures sizes must be multiples of
// 4 bytes.  The cmdsize reflects the total size of the thread_command
// and all of the sizes of the constants for the flavors, counts and state
// data structures.
//
// For executable objects that are unix processes there will be one
// thread_command (cmd == LC_UNIXTHREAD) created for it by the link-editor.
// This is the same as a LC_THREAD, except that a stack is automatically
// created (based on the shell's limit for the stack size).  Command arguments
// and environment variables are copied onto that stack.
type ThreadCommand struct {
	Cmd     uint32 `struc:"little"` // LC_THREAD or  LC_UNIXTHREAD
	CmdSize uint32 `struc:"little"` // total size of this command
	// uint32 flavor		   flavor of thread state
	// uint32 count		   count of uint32's in thread state
	// struct XXX_thread_state state   thread state for this flavor
	// ...
}

// The routines command contains the address of the dynamic shared library
// initialization routine and an index into the module table for the module
// that defines the routine.  Before any modules are used from the library the
// dynamic linker fully binds the module that defines the initialization routine
// and then calls it.  This gets called before any module initialization
// routines (used for C++ static constructors) in the library.
type RoutinesCommand struct { // for 32-bit architectures
	Cmd         uint32 `struc:"little"` // LC_ROUTINES
	CmdSize     uint32 `struc:"little"` // total size of this command
	InitAddress uint32 `struc:"little"` // address of initialization routine
	InitModule  uint32 `struc:"little"` // index into the module table that the init routine is defined in
	Reserved1   uint32 `struc:"little"`
	Reserved2   uint32 `struc:"little"`
	Reserved3   uint32 `struc:"little"`
	Reserved4   uint32 `struc:"little"`
	Reserved5   uint32 `struc:"little"`
	Reserved6   uint32 `struc:"little"`
}

// The 64-bit routines command.  Same use as above.
type RoutinesCommand64 struct { // for 64-bit architectures
	Cmd         uint32 `struc:"little"` // LC_ROUTINES_64
	CmdSize     uint32 `struc:"little"` // total size of this command
	InitAddress uint64 `struc:"little"` // address of initialization routine
	InitModule  uint64 `struc:"little"` // index into the module table that the init routine is defined in
	Reserved1   uint64 `struc:"little"`
	Reserved2   uint64 `struc:"little"`
	Reserved3   uint64 `struc:"little"`
	Reserved4   uint64 `struc:"little"`
	Reserved5   uint64 `struc:"little"`
	Reserved6   uint64 `struc:"little"`
}

// The symtab_command contains the offsets and sizes of the link-edit 4.3BSD
// "stab" style symbol table information as described in the header files
// <nlist.h> and <stab.h>.
type SymtabCommand struct {
	Cmd     uint32 `struc:"little"` // LC_SYMTAB
	CmdSize uint32 `struc:"little"` // sizeof(struct symtab_command)
	SymOff  uint32 `struc:"little"` // symbol table offset
	NSyms   uint32 `struc:"little"` // number of symbol table entries
	StrOff  uint32 `struc:"little"` // string table offset
	StrSize uint32 `struc:"little"` // string table size in bytes
}

// This is the second set of the symbolic information which is used to support
// the data structures for the dynamically link editor.
//
// The original set of symbolic information in the symtab_command which contains
// the symbol and string tables must also be present when this load command is
// present.  When this load command is present the symbol table is organized
// into three groups of symbols:
// local symbols (static and debugging symbols) - grouped by module
// defined external symbols - grouped by module (sorted by name if not lib)
// undefined external symbols (sorted by name if MH_BINDATLOAD is not set,
//
//	     			    and in order the were seen by the static
//				    linker if MH_BINDATLOAD is set)
//
// In this load command there are offsets and counts to each of the three groups
// of symbols.
//
// This load command contains a the offsets and sizes of the following new
// symbolic information tables:
//
//	table of contents
//	module table
//	reference symbol table
//	indirect symbol table
//
// The first three tables above (the table of contents, module table and
// reference symbol table) are only present if the file is a dynamically linked
// shared library.  For executable and object modules, which are files
// containing only one module, the information that would be in these three
// tables is determined as follows:
//
//		table of contents - the defined external symbols are sorted by name
//	 module table - the file contains only one module so everything in the
//		       file is part of the module.
//	 reference symbol table - is the defined and undefined external symbols
//
// For dynamically linked shared library files this load command also contains
// offsets and sizes to the pool of relocation entries for all sections
// separated into two groups:
//
//	external relocation entries
//	local relocation entries
//
// For executable and object modules the relocation entries continue to hang
// off the section structures.
type DYSymTabCommand struct {
	Cmd     uint32 `struc:"little"` // LC_DYSYMTAB
	CmdSize uint32 `struc:"little"` // sizeof(struct dysymtab_command)

	// The symbols indicated by symoff and nsyms of the LC_SYMTAB load command
	// are grouped into the following three groups:
	//    local symbols (further grouped by the module they are from)
	//    defined external symbols (further grouped by the module they are from)
	//    undefined symbols
	//
	// The local symbols are used only for debugging.  The dynamic binding
	// process may have to use them to indicate to the debugger the local
	// symbols for a module that is being bound.
	//
	// The last two groups are used by the dynamic binding process to do the
	// binding (indirectly through the module table and the reference symbol
	// table when this is a dynamically linked shared library file).

	ILocalSym uint32 `struc:"little"` // index to local symbols
	NLocalSym uint32 `struc:"little"` // number of local symbols

	IExtdefSym uint32 `struc:"little"` // index to externally defined symbols
	NExtdefSym uint32 `struc:"little"` // number of externally defined symbols

	IUndefSym uint32 `struc:"little"` // index to undefined symbols
	NUndefSym uint32 `struc:"little"` // number of undefined symbols

	// For the for the dynamic binding process to find which module a symbol
	// is defined in the table of contents is used (analogous to the ranlib
	// structure in an archive) which maps defined external symbols to modules
	// they are defined in.  This exists only in a dynamically linked shared
	// library file.  For executable and object modules the defined external
	// symbols are sorted by name and is use as the table of contents.

	TOCOff uint32 `struc:"little"` // file offset to table of contents
	NTOC   uint32 `struc:"little"` // number of entries in table of contents

	// To support dynamic binding of "modules" (whole object files) the symbol
	// table must reflect the modules that the file was created from.  This is
	// done by having a module table that has indexes and counts into the merged
	// tables for each module.  The module structure that these two entries
	// refer to is described below.  This exists only in a dynamically linked
	// shared library file.  For executable and object modules the file only
	// contains one module so everything in the file belongs to the module.

	ModTabOff uint32 `struc:"little"` // file offset to module table
	NModTab   uint32 `struc:"little"` // number of module table entries

	// To support dynamic module binding the module structure for each module
	// indicates the external references (defined and undefined) each module
	// makes.  For each module there is an offset and a count into the
	// reference symbol table for the symbols that the module references.
	// This exists only in a dynamically linked shared library file.  For
	// executable and object modules the defined external symbols and the
	// undefined external symbols indicates the external references.

	ExtRefSymOff uint32 `struc:"little"` // offset to referenced symbol table
	NExtRefSyms  uint32 `struc:"little"` // number of referenced symbol table entries

	// The sections that contain "symbol pointers" and "routine stubs" have
	// indexes and (implied counts based on the size of the section and fixed
	// size of the entry) into the "indirect symbol" table for each pointer
	// and stub.  For every section of these two types the index into the
	// indirect symbol table is stored in the section header in the field
	// reserved1.  An indirect symbol table entry is simply a 32bit index into
	// the symbol table to the symbol that the pointer or stub is referring to.
	// The indirect symbol table is ordered to match the entries in the section.

	IndirectSymOff uint32 `struc:"little"` // file offset to the indirect symbol table
	NIndirectSyms  uint32 `struc:"little"` // number of indirect symbol table entries

	// To support relocating an individual module in a library file quickly the
	// external relocation entries for each module in the library need to be
	// accessed efficiently.  Since the relocation entries can't be accessed
	// through the section headers for a library file they are separated into
	// groups of local and external entries further grouped by module.  In this
	// case the presents of this load command who's extreloff, nextrel,
	// locreloff and nlocrel fields are non-zero indicates that the relocation
	// entries of non-merged sections are not referenced through the section
	// structures (and the reloff and nreloc fields in the section headers are
	// set to zero).
	//
	// Since the relocation entries are not accessed through the section headers
	// this requires the r_address field to be something other than a section
	// offset to identify the item to be relocated.  In this case r_address is
	// set to the offset from the vmaddr of the first LC_SEGMENT command.
	// For MH_SPLIT_SEGS images r_address is set to the the offset from the
	// vmaddr of the first read-write LC_SEGMENT command.
	//
	// The relocation entries are grouped by module and the module table
	// entries have indexes and counts into them for the group of external
	// relocation entries for that the module.
	//
	// For sections that are merged across modules there must not be any
	// remaining external relocation entries for them (for merged sections
	// remaining relocation entries must be local).

	ExtRelOff uint32 `struc:"little"` // offset to external relocation entries
	NExtRel   uint32 `struc:"little"` // number of external relocation entries

	// All the local relocation entries are grouped together (they are not
	// grouped by their module since they are only used if the object is moved
	// from it staticly link edited address).

	LocRelOff uint32 `struc:"little"` // offset to local relocation entries
	NLocRel   uint32 `struc:"little"` // number of local relocation entries
}

// An indirect symbol table entry is simply a 32bit index into the symbol table
// to the symbol that the pointer or stub is refering to.  Unless it is for a
// non-lazy symbol pointer section for a defined symbol which strip(1) as
// removed.  In which case it has the value INDIRECT_SYMBOL_LOCAL.  If the
// symbol was also absolute INDIRECT_SYMBOL_ABS is or'ed with that.
const (
	INDIRECT_SYMBOL_LOCAL = 0x80000000
	INDIRECT_SYMBOL_ABS   = 0x40000000
)

// a table of contents entry
type DYLIBTableOfContents struct {
	SymbolIndex uint32 `struc:"little"` // the defined external symbol (index into the symbol table)
	ModuleIndex uint32 `struc:"little"` // index into the module table this symbol is defined in
}

// a module table entry
type DYLIBModule struct {
	ModuleName uint32 `struc:"little"` // the module name (index into string table)

	IExtDefSym uint32 `struc:"little"` // index into externally defined symbols
	NExtDefSym uint32 `struc:"little"` // number of externally defined symbols
	IRefSym    uint32 `struc:"little"` // index into reference symbol table
	NRefSym    uint32 `struc:"little"` // number of reference symbol table entries
	ILocalSym  uint32 `struc:"little"` // index into symbols for local symbols
	NLocalSym  uint32 `struc:"little"` // number of local symbols

	IExtRel uint32 `struc:"little"` // index into external relocation entries
	NExtRel uint32 `struc:"little"` // number of external relocation entries

	IInitITerm uint32 `struc:"little"` // low 16 bits are the index into the init section, high 16 bits are the index into the term section
	NInitNTerm uint32 `struc:"little"` // low 16 bits are the number of init section entries, high 16 bits are the number of term section entries

	ObjcModuleInfoAddr uint32 `struc:"little"` // for this module address of the start of, the (__OBJC,__module_info) section
	ObjcModuleInfoSize uint32 `struc:"little"` // for this module size of, the (__OBJC,__module_info) section
}

// a 64-bit module table entry
type DYLIBModule64 struct {
	ModuleName uint32 `struc:"little"` // the module name (index into string table)

	IExtDefSym uint32 `struc:"little"` // index into externally defined symbols
	NExtDefSym uint32 `struc:"little"` // number of externally defined symbols
	IRefSym    uint32 `struc:"little"` // index into reference symbol table
	NRefSym    uint32 `struc:"little"` // number of reference symbol table entries
	ILocalSym  uint32 `struc:"little"` // index into symbols for local symbols
	NLocalSym  uint32 `struc:"little"` // number of local symbols

	IExtRel uint32 `struc:"little"` // index into external relocation entries
	NExtRel uint32 `struc:"little"` // number of external relocation entries

	IInitITerm uint32 `struc:"little"` // low 16 bits are the index into the init section, high 16 bits are the index into the term section
	NInitNTerm uint32 `struc:"little"` // low 16 bits are the number of init section entries, high 16 bits are the number of term section entries

	ObjcModuleInfoSize uint32 `struc:"little"` // for this module size of, the (__OBJC,__module_info) section
	ObjcModuleInfoAddr uint64 `struc:"little"` // for this module address of the start of, the (__OBJC,__module_info) section
}

// The entries in the reference symbol table are used when loading the module
// (both by the static and dynamic link editors) and if the module is unloaded
// or replaced.  Therefore all external symbols (defined and undefined) are
// listed in the module's reference table.  The flags describe the type of
// reference that is being made.  The constants for the flags are defined in
// <mach-o/nlist.h> as they are also used for symbol table entries.
type DYLIBReference struct {
	BitField DYLIBReferenceBitField `struc:"little"`
}

type DYLIBReferenceBitField uint32

func (me DYLIBReferenceBitField) ISym() int {
	return int(me & 0xffffff) // index into the symbol table
}

func (me DYLIBReferenceBitField) Flags() uint8 {
	return uint8(me >> 24) // flags to indicate the type of reference
}

func (me DYLIBReferenceBitField) String() string {
	return fmt.Sprintf("{ISym: %d, Flags: %d}", me.ISym(), me.Flags())
}

// The twolevel_hints_command contains the offset and number of hints in the
// two-level namespace lookup hints table.
type TwoLevelHintsCommand struct {
	Cmd     uint32 `struc:"little"` // LC_TWOLEVEL_HINTS
	CmdSize uint32 `struc:"little"` // sizeof(struct twolevel_hints_command)
	Offset  uint32 `struc:"little"` // offset to the hint table
	NHints  uint32 `struc:"little"` // number of hints in the hint table
}

// The entries in the two-level namespace lookup hints table are twolevel_hint
// structs.  These provide hints to the dynamic link editor where to start
// looking for an undefined symbol in a two-level namespace image.  The
// isub_image field is an index into the sub-images (sub-frameworks and
// sub-umbrellas list) that made up the two-level image that the undefined
// symbol was found in when it was built by the static link editor.  If
// isub-image is 0 the the symbol is expected to be defined in library and not
// in the sub-images.  If isub-image is non-zero it is an index into the array
// of sub-images for the umbrella with the first index in the sub-images being
// 1. The array of sub-images is the ordered list of sub-images of the umbrella
// that would be searched for a symbol that has the umbrella recorded as its
// primary library.  The table of contents index is an index into the
// library's table of contents.  This is used as the starting point of the
// binary search or a directed linear search.
type TwoLevelHint struct {
	BitField TwoLevelHintBitField `struc:"little"`
}

type TwoLevelHintBitField uint32

func (me TwoLevelHintBitField) ISubImage() uint8 {
	return uint8(me & 0xff)
}

func (me TwoLevelHintBitField) ITOC() uint32 {
	return uint32(me >> 8)
}

func (me TwoLevelHintBitField) String() string {
	return fmt.Sprintf("{ISubImage: %d, ITOC: %d}", me.ISubImage(), me.ITOC())
}

// The prebind_cksum_command contains the value of the original check sum for
// prebound files or zero.  When a prebound file is first created or modified
// for other than updating its prebinding information the value of the check sum
// is set to zero.  When the file has it prebinding re-done and if the value of
// the check sum is zero the original check sum is calculated and stored in
// cksum field of this load command in the output file.  If when the prebinding
// is re-done and the cksum field is non-zero it is left unchanged from the
// input file.
type PrebindCKSumCommand struct {
	Cmd     uint32 `struc:"little"` // LC_PREBIND_CKSUM
	CmdSize uint32 `struc:"little"` // sizeof(struct prebind_cksum_command)
	CKSum   uint32 `struc:"little"` // the check sum or zero
}

// The uuid load command contains a single 128-bit unique random number that
// identifies an object produced by the static link editor.
type UUIDCommand struct {
	Cmd     uint32    `struc:"little"` // LC_UUID
	CmdSize uint32    `struc:"little"` // sizeof(struct uuid_command)
	UUID    [16]uint8 // the 128-bit uuid
}

// The rpath_command contains a path which at runtime should be added to
// the current run path used to find @rpath prefixed dylibs.
type RPathCommand struct {
	Cmd     uint32 `struc:"little"` // LC_RPATH
	CmdSize uint32 `struc:"little"` // includes string
	Path    uint32 `struc:"little"` // path to add to run path
}

// The linkedit_data_command contains the offsets and sizes of a blob
// of data in the __LINKEDIT segment.
type LinkEditDataCommand struct {
	Cmd      uint32 `struc:"little"` // LC_CODE_SIGNATURE, LC_SEGMENT_SPLIT_INFO, LC_FUNCTION_STARTS, LC_DATA_IN_CODE, LC_DYLIB_CODE_SIGN_DRS, LC_LINKER_OPTIMIZATION_HINT, LC_DYLD_EXPORTS_TRIE, or LC_DYLD_CHAINED_FIXUPS.
	CmdSize  uint32 `struc:"little"` // sizeof(struct linkedit_data_command)
	DataOff  uint32 `struc:"little"` // file offset of data in __LINKEDIT segment
	DataSize uint32 `struc:"little"` // file size of data in __LINKEDIT segment
}

type FilesetEntryCommand struct {
	Cmd      uint32 `struc:"little"` // LC_FILESET_ENTRY
	CmdSize  uint32 `struc:"little"` // includes id string
	VMAddr   uint64 `struc:"little"` // memory address of the dylib
	FileOff  uint64 `struc:"little"` // file offset of the dylib
	EntryID  uint32 `struc:"little"` // contained entry id
	Reserved uint32 `struc:"little"` // entry_id is 32-bits long, so this is the reserved padding
}

// The encryption_info_command contains the file offset and size of an
// of an encrypted segment.
type EncryptionInfoCommand struct {
	Cmd       uint32 `struc:"little"` // LC_ENCRYPTION_INFO
	CmdSize   uint32 `struc:"little"` // sizeof(struct encryption_info_command)
	CryptOff  uint32 `struc:"little"` // file offset of encrypted range
	CryptSize uint32 `struc:"little"` // file size of encrypted range
	CryptID   uint32 `struc:"little"` // which enryption system, 0 means not-encrypted yet
}

// The encryption_info_command_64 contains the file offset and size of an
// of an encrypted segment (for use in x86_64 targets).
type EncryptionInfoCommand64 struct {
	Cmd       uint32 `struc:"little"` // LC_ENCRYPTION_INFO_64
	CmdSize   uint32 `struc:"little"` // sizeof(struct encryption_info_command_64)
	CryptOff  uint32 `struc:"little"` // file offset of encrypted range
	CryptSize uint32 `struc:"little"` // file size of encrypted range
	CryptID   uint32 `struc:"little"` // which enryption system, 0 means not-encrypted yet
	Pad       uint32 `struc:"little"` // padding to make this struct's size a multiple of 8 bytes
}

// The version_min_command contains the min OS version on which this
// binary was built to run.
type VersionMinCommand struct {
	Cmd     uint32 `struc:"little"` // LC_VERSION_MIN_MACOSX or LC_VERSION_MIN_IPHONEOS or LC_VERSION_MIN_WATCHOS or LC_VERSION_MIN_TVOS
	CmdSize uint32 `struc:"little"` // sizeof(struct min_version_command)
	Version uint32 `struc:"little"` // X.Y.Z is encoded in nibbles xxxx.yy.zz
	SDK     uint32 `struc:"little"` // X.Y.Z is encoded in nibbles xxxx.yy.zz
}

// The build_version_command contains the min OS version on which this
// binary was built to run for its platform.  The list of known platforms and
// tool values following it.
type BuildVersionCommand struct {
	Cmd      uint32 `struc:"little"` // LC_BUILD_VERSION
	CmdSize  uint32 `struc:"little"` // sizeof(struct build_version_command) plus ntools// sizeof(struct build_tool_version)
	Platform uint32 `struc:"little"` // platform
	MinOS    uint32 `struc:"little"` // X.Y.Z is encoded in nibbles xxxx.yy.zz
	SDK      uint32 `struc:"little"` // X.Y.Z is encoded in nibbles xxxx.yy.zz
	NTools   uint32 `struc:"little"` // number of tool entries following this
}

type BuildToolVersion struct {
	Tool    uint32 `struc:"little"` // enum for the tool
	Version uint32 `struc:"little"` // version number of the tool
}

// Known values for the platform field above.
const (
	PLATFORM_MACOS            = 1
	PLATFORM_IOS              = 2
	PLATFORM_TVOS             = 3
	PLATFORM_WATCHOS          = 4
	PLATFORM_BRIDGEOS         = 5
	PLATFORM_MACCATALYST      = 6
	PLATFORM_IOSSIMULATOR     = 7
	PLATFORM_TVOSSIMULATOR    = 8
	PLATFORM_WATCHOSSIMULATOR = 9
	PLATFORM_DRIVERKIT        = 10
	PLATFORM_MAX              = PLATFORM_DRIVERKIT
	// Addition of simulated platfrom also needs to update proc_is_simulated()
)

// Known values for the tool field above.
const (
	TOOL_CLANG = 1
	TOOL_SWIFT = 2
	TOOL_LD    = 3
)

// The dyld_info_command contains the file offsets and sizes of
// the new compressed form of the information dyld needs to
// load the image.  This information is used by dyld on Mac OS X
// 10.6 and later.  All information pointed to by this command
// is encoded using byte streams, so no endian swapping is needed
// to interpret it.
type DYLDInfoCommand struct {
	Cmd     uint32 `struc:"little"` // LC_DYLD_INFO or LC_DYLD_INFO_ONLY
	CmdSize uint32 `struc:"little"` // sizeof(struct dyld_info_command)

	// Dyld rebases an image whenever dyld loads it at an address different
	// from its preferred address.  The rebase information is a stream
	// of byte sized opcodes whose symbolic names start with REBASE_OPCODE_.
	// Conceptually the rebase information is a table of tuples:
	//    <seg-index, seg-offset, type>
	// The opcodes are a compressed way to encode the table by only
	// encoding when a column changes.  In addition simple patterns
	// like "every n'th offset for m times" can be encoded in a few
	// bytes.

	RebaseOff  uint32 `struc:"little"` // file offset to rebase info
	RebaseSize uint32 `struc:"little"` // size of rebase info

	// Dyld binds an image during the loading process, if the image
	// requires any pointers to be initialized to symbols in other images.
	// The bind information is a stream of byte sized
	// opcodes whose symbolic names start with BIND_OPCODE_.
	// Conceptually the bind information is a table of tuples:
	//    <seg-index, seg-offset, type, symbol-library-ordinal, symbol-name, addend>
	// The opcodes are a compressed way to encode the table by only
	// encoding when a column changes.  In addition simple patterns
	// like for runs of pointers initialzed to the same value can be
	// encoded in a few bytes.

	BindOff  uint32 `struc:"little"` // file offset to binding info
	BindSize uint32 `struc:"little"` // size of binding info

	// Some C++ programs require dyld to unique symbols so that all
	// images in the process use the same copy of some code/data.
	// This step is done after binding. The content of the weak_bind
	// info is an opcode stream like the bind_info.  But it is sorted
	// alphabetically by symbol name.  This enable dyld to walk
	// all images with weak binding information in order and look
	// for collisions.  If there are no collisions, dyld does
	// no updating.  That means that some fixups are also encoded
	// in the bind_info.  For instance, all calls to "operator new"
	// are first bound to libstdc++.dylib using the information
	// in bind_info.  Then if some image overrides operator new
	// that is detected when the weak_bind information is processed
	// and the call to operator new is then rebound.

	WeakBindOff  uint32 `struc:"little"` // file offset to weak binding info
	WeakBindSize uint32 `struc:"little"` // size of weak binding info

	// Some uses of external symbols do not need to be bound immediately.
	// Instead they can be lazily bound on first use.  The lazy_bind
	// are contains a stream of BIND opcodes to bind all lazy symbols.
	// Normal use is that dyld ignores the lazy_bind section when
	// loading an image.  Instead the static linker arranged for the
	// lazy pointer to initially point to a helper function which
	// pushes the offset into the lazy_bind area for the symbol
	// needing to be bound, then jumps to dyld which simply adds
	// the offset to lazy_bind_off to get the information on what
	// to bind.

	LazyBindOff  uint32 `struc:"little"` // file offset to lazy binding info
	LazyBindSize uint32 `struc:"little"` // size of lazy binding infs

	// The symbols exported by a dylib are encoded in a trie.  This
	// is a compact representation that factors out common prefixes.
	// It also reduces LINKEDIT pages in RAM because it encodes all
	// information (name, address, flags) in one small, contiguous range.
	// The export area is a stream of nodes.  The first node sequentially
	// is the start node for the trie.
	//
	// Nodes for a symbol start with a uleb128 that is the length of
	// the exported symbol information for the string so far.
	// If there is no exported symbol, the node starts with a zero byte.
	// If there is exported info, it follows the length.
	//
	// First is a uleb128 containing flags. Normally, it is followed by
	//  a uleb128 encoded offset which is location of the content named
	//  by the symbol from the mach_header for the image.  If the flags
	//  is EXPORT_SYMBOL_FLAGS_REEXPORT, then following the flags is
	//  a uleb128 encoded library ordinal, then a zero terminated
	//  UTF8 string.  If the string is zero length, then the symbol
	//  is re-export from the specified dylib with the same name.
	// If the flags is EXPORT_SYMBOL_FLAGS_STUB_AND_RESOLVER, then following
	// the flags is two uleb128s: the stub offset and the resolver offset.
	// The stub is used by non-lazy pointers.  The resolver is used
	// by lazy pointers and must be called to get the actual address to use.
	//  After the optional exported symbol information is a byte of
	//  how many edges (0-255) that this node has leaving it,
	//  followed by each edge.
	//  Each edge is a zero terminated UTF8 of the addition chars
	//  in the symbol, followed by a uleb128 offset for the node that
	//  edge points to.

	ExportOff  uint32 `struc:"little"` // file offset to lazy binding info
	ExportSize uint32 `struc:"little"` // size of lazy binding infs
}

// The following are used to encode rebasing information
const (
	REBASE_TYPE_POINTER         = 1
	REBASE_TYPE_TEXT_ABSOLUTE32 = 2
	REBASE_TYPE_TEXT_PCREL32    = 3

	REBASE_OPCODE_MASK                               = 0xF0
	REBASE_IMMEDIATE_MASK                            = 0x0F
	REBASE_OPCODE_DONE                               = 0x00
	REBASE_OPCODE_SET_TYPE_IMM                       = 0x10
	REBASE_OPCODE_SET_SEGMENT_AND_OFFSET_ULEB        = 0x20
	REBASE_OPCODE_ADD_ADDR_ULEB                      = 0x30
	REBASE_OPCODE_ADD_ADDR_IMM_SCALED                = 0x40
	REBASE_OPCODE_DO_REBASE_IMM_TIMES                = 0x50
	REBASE_OPCODE_DO_REBASE_ULEB_TIMES               = 0x60
	REBASE_OPCODE_DO_REBASE_ADD_ADDR_ULEB            = 0x70
	REBASE_OPCODE_DO_REBASE_ULEB_TIMES_SKIPPING_ULEB = 0x80
)

// The following are used to encode binding information
const (
	BIND_TYPE_POINTER         = 1
	BIND_TYPE_TEXT_ABSOLUTE32 = 2
	BIND_TYPE_TEXT_PCREL32    = 3

	BIND_SPECIAL_DYLIB_SELF            = 0
	BIND_SPECIAL_DYLIB_MAIN_EXECUTABLE = -1
	BIND_SPECIAL_DYLIB_FLAT_LOOKUP     = -2
	BIND_SPECIAL_DYLIB_WEAK_LOOKUP     = -3

	BIND_SYMBOL_FLAGS_WEAK_IMPORT         = 0x1
	BIND_SYMBOL_FLAGS_NON_WEAK_DEFINITION = 0x8

	BIND_OPCODE_MASK                                         = 0xF0
	BIND_IMMEDIATE_MASK                                      = 0x0F
	BIND_OPCODE_DONE                                         = 0x00
	BIND_OPCODE_SET_DYLIB_ORDINAL_IMM                        = 0x10
	BIND_OPCODE_SET_DYLIB_ORDINAL_ULEB                       = 0x20
	BIND_OPCODE_SET_DYLIB_SPECIAL_IMM                        = 0x30
	BIND_OPCODE_SET_SYMBOL_TRAILING_FLAGS_IMM                = 0x40
	BIND_OPCODE_SET_TYPE_IMM                                 = 0x50
	BIND_OPCODE_SET_ADDEND_SLEB                              = 0x60
	BIND_OPCODE_SET_SEGMENT_AND_OFFSET_ULEB                  = 0x70
	BIND_OPCODE_ADD_ADDR_ULEB                                = 0x80
	BIND_OPCODE_DO_BIND                                      = 0x90
	BIND_OPCODE_DO_BIND_ADD_ADDR_ULEB                        = 0xA0
	BIND_OPCODE_DO_BIND_ADD_ADDR_IMM_SCALED                  = 0xB0
	BIND_OPCODE_DO_BIND_ULEB_TIMES_SKIPPING_ULEB             = 0xC0
	BIND_OPCODE_THREADED                                     = 0xD0
	BIND_SUBOPCODE_THREADED_SET_BIND_ORDINAL_TABLE_SIZE_ULEB = 0x00
	BIND_SUBOPCODE_THREADED_APPLY                            = 0x01
)

// The following are used on the flags byte of a terminal node
// in the export information.
const (
	EXPORT_SYMBOL_FLAGS_KIND_MASK         = 0x03
	EXPORT_SYMBOL_FLAGS_KIND_REGULAR      = 0x00
	EXPORT_SYMBOL_FLAGS_KIND_THREAD_LOCAL = 0x01
	EXPORT_SYMBOL_FLAGS_KIND_ABSOLUTE     = 0x02
	EXPORT_SYMBOL_FLAGS_WEAK_DEFINITION   = 0x04
	EXPORT_SYMBOL_FLAGS_REEXPORT          = 0x08
	EXPORT_SYMBOL_FLAGS_STUB_AND_RESOLVER = 0x10
)

// The linker_option_command contains linker options embedded in object files.
type LinkerOptionCommand struct {
	Cmd     uint32 `struc:"little"` // LC_LINKER_OPTION only used in MH_OBJECT filetypes
	CmdSize uint32 `struc:"little"`
	Count   uint32 `struc:"little"` // number of strings concatenation of zero terminated UTF8 strings. Zero filled at end to align
}

// The symseg_command contains the offset and size of the GNU style
// symbol table information as described in the header file <symseg.h>.
// The symbol roots of the symbol segments must also be aligned properly
// in the file.  So the requirement of keeping the offsets aligned to a
// multiple of a 4 bytes translates to the length field of the symbol
// roots also being a multiple of a long.  Also the padding must again be
// zeroed. (THIS IS OBSOLETE and no longer supported).
type SymSegCommand struct {
	Cmd     uint32 `struc:"little"` // LC_SYMSEG
	CmdSize uint32 `struc:"little"` // sizeof(struct symseg_command)
	Offset  uint32 `struc:"little"` // symbol segment offset
	Size    uint32 `struc:"little"` // symbol segment size in bytes
}

// The ident_command contains a free format string table following the
// ident_command structure.  The strings are null terminated and the size of
// the command is padded out with zero bytes to a multiple of 4 bytes/
// (THIS IS OBSOLETE and no longer supported).
type IdentCommand struct {
	Cmd     uint32 `struc:"little"` // LC_IDENT
	CmdSize uint32 `struc:"little"` // strings that follow this command
}

// The fvmfile_command contains a reference to a file to be loaded at the
// specified virtual address.  (Presently, this command is reserved for
// internal use.  The kernel ignores this command when loading a program into
// memory).
type FVMFileCommand struct {
	Cmd        uint32 `struc:"little"` // LC_FVMFILE
	CmdSize    uint32 `struc:"little"` // includes pathname string
	Name       uint32 `struc:"little"` // files pathname
	HeaderAddr uint32 `struc:"little"` // files virtual address
}

// The entry_point_command is a replacement for thread_command.
// It is used for main executables to specify the location (file offset)
// of main().  If -stack_size was used at link time, the stacksize
// field will contain the stack size need for the main thread.
type EntryPointCommand struct {
	Cmd       uint32 `struc:"little"` // LC_MAIN only used in MH_EXECUTE filetypes
	CmdSize   uint32 `struc:"little"` // 24
	EntryOff  uint64 `struc:"little"` // file (__TEXT) offset of main()
	StackSize uint64 `struc:"little"` // if not zero, initial stack size
}

// The source_version_command is an optional load command containing
// the version of the sources used to build the binary.
type SourceVersionCommand struct {
	Cmd     uint32 `struc:"little"` // LC_SOURCE_VERSION
	CmdSize uint32 `struc:"little"` // 16
	Version uint64 `struc:"little"` // A.B.C.D.E packed as a24.b10.c10.d10.e10
}

// The LC_DATA_IN_CODE load commands uses a linkedit_data_command
// to point to an array of data_in_code_entry entries. Each entry
// describes a range of data in a code section.
type DataInCodeEntry struct {
	Offset uint32 `struc:"little"` // from mach_header to start of data range
	Length uint16 `struc:"little"` // number of bytes in data range
	Kind   uint16 `struc:"little"` // a DICE_KIND_* value
}

const (
	DICE_KIND_DATA             = 0x0001
	DICE_KIND_JUMP_TABLE8      = 0x0002
	DICE_KIND_JUMP_TABLE16     = 0x0003
	DICE_KIND_JUMP_TABLE32     = 0x0004
	DICE_KIND_ABS_JUMP_TABLE32 = 0x0005
)

// Sections of type S_THREAD_LOCAL_VARIABLES contain an array
// of tlv_descriptor structures.
type TLCDescriptor struct {
	Thunk  uint64 `struc:"little"` // void* (*thunk)(struct tlv_descriptor*)
	Key    uint64 `struc:"little"`
	Offset uint64 `struc:"little"`
}

// LC_NOTE commands describe a region of arbitrary data included in a Mach-O
// file.  Its initial use is to record extra data in MH_CORE files.
type NoteCommand struct {
	Cmd       uint32   `struc:"little"` // LC_NOTE
	CmdSize   uint32   `struc:"little"` // sizeof(struct note_command)
	DataOwner [16]byte // owner name for this LC_NOTE
	Offset    uint64   `struc:"little"` // file offset of this data
	Size      uint64   `struc:"little"` // length of data region
}
