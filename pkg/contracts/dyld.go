package contracts

import (
	"fmt"
	"unsafe"
)

const (
	MAGIC_i386           = "dyld_v1    i386"
	MAGIC_x86_64         = "dyld_v1  x86_64"
	MAGIC_x86_64_HASWELL = "dyld_v1 x86_64h"
	MAGIC_armv5          = "dyld_v1   armv5"
	MAGIC_armv6          = "dyld_v1   armv6"
	MAGIC_armv7          = "dyld_v1   armv7"
	MAGIC_armv7x         = "dyld_v1  armv7x"
	MAGIC_arm64          = "dyld_v1   arm64"
	MAGIC_arm64e         = "dyld_v1  arm64e"
	MAGIC_arm64_32       = "dyld_v1arm64_32"
)

// From Apple's `dyld-*/cache-builder/dyld_cache_format.h`

type DYLDCacheMappingInfo struct {
	Address    uint64
	Size       uint64
	FileOffset uint64
	MaxProt    uint32
	InitProt   uint32
}

const (
	DYLD_CACHE_MAPPING_AUTH_DATA   = 1 << 0
	DYLD_CACHE_MAPPING_DIRTY_DATA  = 1 << 1
	DYLD_CACHE_MAPPING_CONST_DATA  = 1 << 2
	DYLD_CACHE_MAPPING_TEXT_STUBS  = 1 << 3
	DYLD_CACHE_DYNAMIC_CONFIG_DATA = 1 << 4
)

type DYLDCacheMappingAndSlideInfo struct {
	Address             uint64
	Size                uint64
	FileOffset          uint64
	SlideInfoFileOffset uint64
	SlideInfoFileSize   uint64
	Flags               uint64
	MaxProt             uint32
	InitProt            uint32
}

type DYLDCacheImageInfo struct {
	Address        uint64
	ModTime        uint64
	Inode          uint64
	PathFileOffset uint32
	Pad            uint32
}

type DYLDCacheImageInfoExtra struct {
	ExportsTrieAddr           uint64 // address of trie in unslid cache
	WeakBindingsAddr          uint64
	ExportsTrieSize           uint32
	WeakBindingsSize          uint32
	DependentsStartArrayIndex uint32
	ReExportsStartArrayIndex  uint32
}

type DYLDCacheAcceleratorInfo struct {
	Version            uint32 // currently 1
	ImageExtrasCount   uint32 // does not include aliases
	ImagesExtrasOffset uint32 // offset into this chunk of first dyld_cache_image_info_extra
	BottomUpListOffset uint32 // offset into this chunk to start of 16-bit array of sorted image indexes
	DylibTrieOffset    uint32 // offset into this chunk to start of trie containing all dylib paths
	DylibTrieSize      uint32 // size of trie containing all dylib paths
	InitializersOffset uint32 // offset into this chunk to start of initializers list
	InitializersCount  uint32 // size of initializers list
	DofSectionsOffset  uint32 // offset into this chunk to start of DOF sections list
	DofSectionsCount   uint32 // size of initializers list
	ReExportListOffset uint32 // offset into this chunk to start of 16-bit array of re-exports
	ReExportCount      uint32 // size of re-exports
	DepListOffset      uint32 // offset into this chunk to start of 16-bit array of dependencies (0x8000 bit set if upward)
	DepListCount       uint32 // size of dependencies
	RangeTableOffset   uint32 // offset into this chunk to start of ss
	RangeTableCount    uint32 // size of dependencies
	DyldSectionAddr    uint64 // address of libdyld's __dyld section in unslid cache
}

type DYLDCacheAcceleratorInitializer struct {
	FunctionOffset uint32 // address offset from start of cache mapping
	ImageIndex     uint32
}

type DYLDCacheRangeEntry struct {
	StartAddress uint64 // unslid address of start of region
	Size         uint32
	ImageIndex   uint32
}

type DYLDCacheAcceleratorDof struct {
	SectionAddress uint64 // unslid address of start of region
	SectionSize    uint32
	ImageIndex     uint32
}

type DYLDCacheImageTextInfo struct {
	UUID            [16]byte
	LoadAddress     uint64
	TextSegmentSize uint32
	PathOffset      uint32
}

type DYLDCacheSlideInfo struct {
	Version       uint32 // currently 1
	TocOffset     uint32
	TocCount      uint32
	EntriesOffset uint32
	EntriesCount  uint32
	EntriesSize   uint32 // currently 128
	// uint16_t toc[toc_count];
	// entrybitmap entries[entries_count];
}

type DYLDCacheSlideInfoEntry struct {
	Bits [4096 / (8 * 4)]uint8 // 128
}

type DYLDCacheSlideInfo2 struct {
	Version          uint32 // currently 2
	PageSize         uint32 // currently 4096 (may also be 16384)
	PageStartsOffset uint32
	PageStartsCount  uint32
	PageExtrasOffset uint32
	PageExtrasCount  uint32
	DeltaMask        uint64 // which (contiguous) set of bits contains the delta to the next rebase location
	ValueAdd         uint64
	//uint16_t    page_starts[page_starts_count];
	//uint16_t    page_extras[page_extras_count];
}

const (
	DYLD_CACHE_SLIDE_PAGE_ATTRS          = 0xC000 // high bits of uint16_t are flags
	DYLD_CACHE_SLIDE_PAGE_ATTR_EXTRA     = 0x8000 // index is into extras array (not starts array)
	DYLD_CACHE_SLIDE_PAGE_ATTR_NO_REBASE = 0x4000 // page has no rebasing
	DYLD_CACHE_SLIDE_PAGE_ATTR_END       = 0x8000 // last chain entry for page
)

type DYLDCacheSlideInfo3 struct {
	Version         uint32 // currently 3
	PageSize        uint32 // currently 4096 (may also be 16384)
	PageStartsCount uint32 `struc:"sizeof=PageStarts"`
	AuthValueAdd    uint64
	PageStarts      []uint16 /* page_starts_count */
}

const (
	DYLD_CACHE_SLIDE_V3_PAGE_ATTR_NO_REBASE = 0xFFFF // page has no rebasing
)

type DYLDCacheSlidePointer3Raw uint64

type DYLDCacheSlidePointer3Plain uint64

func (me DYLDCacheSlidePointer3Plain) PointerValue() uint64 {
	return uint64(me >> 11)
}

func (me DYLDCacheSlidePointer3Plain) OffsetToNextPointer() int16 {
	return int16(me & 0x7FF)
}

func (me DYLDCacheSlidePointer3Plain) Unused() int {
	return int(me & 0x3)
}

func (me DYLDCacheSlidePointer3Plain) String() string {
	return fmt.Sprintf("{PointerValue: %d, OffsetToNextPointer: %d, Unused: %d}", me.PointerValue(), me.OffsetToNextPointer(), me.Unused())
}

type DYLDCacheSlidePointer3Auth uint64

func (me DYLDCacheSlidePointer3Auth) OffsetFromSharedCacheBase() int32 {
	return int32(me >> 32)
}

func (me DYLDCacheSlidePointer3Auth) DiversityData() int16 {
	return int16(me >> 16)
}

func (me DYLDCacheSlidePointer3Auth) HasAddressDiversity() bool {
	return me&0x8000 != 0
}

func (me DYLDCacheSlidePointer3Auth) Key() int {
	return int(me & 0x6000)
}

func (me DYLDCacheSlidePointer3Auth) OffsetToNextPointer() int {
	return int(me & 0x7FF)
}

func (me DYLDCacheSlidePointer3Auth) Unused() int {
	return int(me & 0x2)
}

func (me DYLDCacheSlidePointer3Auth) Authenticated() bool {
	return me&0x1 == 1
}

func (me DYLDCacheSlidePointer3Auth) String() string {
	return fmt.Sprintf("{OffsetFromSharedCacheBase: %d, DiversityData: %d, HasAddressDiversity: %t, Key: %d, OffsetToNextPointer: %d, Unused: %d, Authenticated: %t}", me.OffsetFromSharedCacheBase(), me.DiversityData(), me.HasAddressDiversity(), me.Key(), me.OffsetToNextPointer(), me.Unused(), me.Authenticated())
}

type DYLDCacheSlideInfo4 struct {
	Version          uint32 // currently 4
	PageSize         uint32 // currently 4096 (may also be 16384)
	PageStartsOffset uint32
	PageStartsCount  uint32
	PageExtrasOffset uint32
	PageExtrasCount  uint32
	DeltaMask        uint64 // which (contiguous) set of bits contains the delta to the next rebase location (0xC0000000)
	ValueAdd         uint64 // base address of cache
	//uint16_t    page_starts[page_starts_count];
	//uint16_t    page_extras[page_extras_count];
}

const (
	DYLD_CACHE_SLIDE4_PAGE_NO_REBASE = 0xFFFF // page has no rebasing
	DYLD_CACHE_SLIDE4_PAGE_INDEX     = 0x7FFF // mask of page_starts[] values
	DYLD_CACHE_SLIDE4_PAGE_USE_EXTRA = 0x8000 // index is into extras array (not a chain start offset)
	DYLD_CACHE_SLIDE4_PAGE_EXTRA_END = 0x8000 // last chain entry for page
)

type DYLDCacheLocalSymbolsInfo struct {
	NlistOffset   uint32 // offset into this chunk of nlist entries
	NlistCount    uint32 // count of nlist entries
	StringsOffset uint32 // offset into this chunk of string pool
	StringsSize   uint32 // byte count of string pool
	EntriesOffset uint32 // offset into this chunk of array of dyld_cache_local_symbols_entry
	EntriesCount  uint32 // number of elements in dyld_cache_local_symbols_entry array
}

type DYLDCacheLocalSymbolsEntry struct {
	DylibOffset     uint32 // offset in cache file of start of dylib
	NlistStartIndex uint32 // start index of locals for this dylib
	NlistCount      uint32 // number of local symbols for this dylib
}

type DYLDCacheLocalSymbolsEntry64 struct {
	DylibOffset     uint64 // offset in cache file of start of dylib
	NlistStartIndex uint32 // start index of locals for this dylib
	NlistCount      uint32 // number of local symbols for this dylib
}

type DYLDSubcacheEntryV1 struct {
	UUID          [16]byte // The UUID of the subCache file
	CacheVmOffset uint64   // The offset of this subcache from the main cache base address
}

type DYLDSubcacheEntryV2 struct {
	DYLDSubcacheEntryV1
	FileSuffix [32]byte // The file name suffix of the subCache file e.g. ".25.data", ".03.development"
}

// This struct is a small piece of dynamic data that can be included in the shared region, and contains configuration
// data about the shared cache in use by the process. It is located
type DYLDCacheDynamicDataHeader struct {
	Magic   [16]byte // e.g. "dyld_data    v0"
	FsId    uint64   // The fsid_t of the shared cache being used by a process
	FsObjId uint64   // The fs_obj_id_t of the shared cache being used by a process
}

const (
	MACOSX_DYLD_SHARED_CACHE_DIR         = "/System/Library/dyld/"
	IPHONE_DYLD_SHARED_CACHE_DIR         = "/System/Library/Caches/com.apple.dyld/"
	DRIVERKIT_DYLD_SHARED_CACHE_DIR      = "/System/DriverKit/System/Library/dyld/"
	DYLD_SHARED_CACHE_BASE_NAME          = "dyld_shared_cache_"
	DYLD_SIM_SHARED_CACHE_BASE_NAME      = "dyld_sim_shared_cache_"
	DYLD_SHARED_CACHE_DEVELOPMENT_EXT    = ".development"
	DYLD_SHARED_CACHE_DYNAMIC_DATA_MAGIC = "dyld_data    v0"
)

var CryptexPrefixes = []string{
	"/System/Volumes/Preboot/Cryptexes/OS/",
	"/private/preboot/Cryptexes/OS/",
	"/System/Cryptexes/OS",
}

const (
	DYLD_SHARED_CACHE_TYPE_DEVELOPMENT = 0
	DYLD_SHARED_CACHE_TYPE_PRODUCTION  = 1
	DYLD_SHARED_CACHE_TYPE_UNIVERSAL   = 2
)

type DYLDCacheHeaderV1 struct {
	Magic                [16]byte  // e.g. "dyld_v0    i386"
	MappingOffset        uint32    // file offset to first dyld_cache_mapping_info
	MappingCount         uint32    // number of dyld_cache_mapping_info entries
	ImagesOffset         uint32    // file offset to first dyld_cache_image_info
	ImagesCount          uint32    // number of dyld_cache_image_info entries
	DyldBaseAddress      uint64    // base address of dyld when cache was built
	CodeSignatureOffset  uint64    // file offset of code signature blob
	CodeSignatureSize    uint64    // size of code signature blob (zero means to end of file)
	SlideInfoOffset      uint64    // file offset of kernel slid info
	SlideInfoSize        uint64    // size of kernel slid info
	LocalSymbolsOffset   uint64    // file offset of where local symbols are stored
	LocalSymbolsSize     uint64    // size of local symbols information
	UUID                 [16]uint8 // unique value for each shared cache file
	CacheType            uint64    // 0 for development, 1 for production
	BranchPoolsOffset    uint32    // file offset to table of uint64_t pool addresses
	BranchPoolsCount     uint32    // number of uint64_t entries
	AccelerateInfoAddr   uint64    // (unslid) address of optimization info
	AccelerateInfoSize   uint64    // size of optimization info
	ImagesTextOffset     uint64    // file offset to first dyld_cache_image_text_info
	ImagesTextCount      uint64    // number of dyld_cache_image_text_info entries
	DylibsImageGroupAddr uint64    // (unslid) address of ImageGroup for dylibs in this cache
	DylibsImageGroupSize uint64    // size of ImageGroup for dylibs in this cache
	OtherImageGroupAddr  uint64    // (unslid) address of ImageGroup for other OS dylibs
	OtherImageGroupSize  uint64    // size of oImageGroup for other OS dylibs
	ProgClosuresAddr     uint64    // (unslid) address of list of program launch closures
	ProgClosuresSize     uint64    // size of list of program launch closures
	ProgClosuresTrieAddr uint64    // (unslid) address of trie of indexes into program launch closures
	ProgClosuresTrieSize uint64    // size of trie of indexes into program launch closures
	Platform             uint32    // platform number (macOS=1, etc)
	BitField             DYLDCacheHeaderV1BitField
	SharedRegionStart    uint64 // base load address of cache if not slid
	SharedRegionSize     uint64 // overall size of region cache can be mapped into
	MaxSlide             uint64 // runtime slide of cache can be between zero and this value
	DylibsImageArrayAddr uint64 // (unslid) address of ImageArray for dylibs in this cache
	DylibsImageArraySize uint64 // size of ImageArray for dylibs in this cache
	DylibsTrieAddr       uint64 // (unslid) address of trie of indexes of all cached dylibs
	DylibsTrieSize       uint64 // size of trie of cached dylib paths
	OtherImageArrayAddr  uint64 // (unslid) address of ImageArray for dylibs and bundles with dlopen closures
	OtherImageArraySize  uint64 // size of ImageArray for dylibs and bundles with dlopen closures
	OtherTrieAddr        uint64 // (unslid) address of trie of indexes of all dylibs and bundles with dlopen closures
	OtherTrieSize        uint64 // size of trie of dylibs and bundles with dlopen closures
}

type DYLDCacheHeaderV1BitField uint32

func (me DYLDCacheHeaderV1BitField) FormatVersion() int {
	return int(me >> 24 & 0xFF) // dyld3::closure::kFormatVersion
}

func (me DYLDCacheHeaderV1BitField) DylibsExpectedOnDisk() bool {
	return me>>23&0x1 == 1 // dyld should expect the dylib exists on disk and to compare inode/mtime to see if cache is valid
}

func (me DYLDCacheHeaderV1BitField) Simulator() bool {
	return me>>22&0x1 == 1 // for simulator of specified platform
}

func (me DYLDCacheHeaderV1BitField) LocallyBuiltCache() bool {
	return me>>21&0x1 == 1 // 0 for B&I built cache, 1 for locally built cache
}

func (me DYLDCacheHeaderV1BitField) Padding() int {
	return int(me & 0x1FFFFF) // TBD
}

func (me DYLDCacheHeaderV1BitField) String() string {
	return fmt.Sprintf("{FormatVersion: %d, DylibsExpectedOnDisk: %t, Simulator: %t, LocallyBuiltCache: %t, Padding: %d}", me.FormatVersion(), me.DylibsExpectedOnDisk(), me.Simulator(), me.LocallyBuiltCache(), me.Padding())
}

type DYLDCacheHeaderV2 struct {
	Magic                     [16]byte  // e.g. "dyld_v0    i386"
	MappingOffset             uint32    // file offset to first dyld_cache_mapping_info
	MappingCount              uint32    // number of dyld_cache_mapping_info entries
	ImagesOffsetOld           uint32    // UNUSED: moved to imagesOffset to prevent older dsc_extarctors from crashing
	ImagesCountOld            uint32    // UNUSED: moved to imagesCount to prevent older dsc_extarctors from crashing
	DyldBaseAddress           uint64    // base address of dyld when cache was built
	CodeSignatureOffset       uint64    // file offset of code signature blob
	CodeSignatureSize         uint64    // size of code signature blob (zero means to end of file)
	SlideInfoOffsetUnused     uint64    // unused.  Used to be file offset of kernel slid info
	SlideInfoSizeUnused       uint64    // unused.  Used to be size of kernel slid info
	LocalSymbolsOffset        uint64    // file offset of where local symbols are stored
	LocalSymbolsSize          uint64    // size of local symbols information
	UUID                      [16]uint8 // unique value for each shared cache file
	CacheType                 uint64    // 0 for development, 1 for production, 2 for multi-cache
	BranchPoolsOffset         uint32    // file offset to table of uint64_t pool addresses
	BranchPoolsCount          uint32    // number of uint64_t entries
	DyldInCacheMh             uint64    // (unslid) address of mach_header of dyld in cache
	DyldInCacheEntry          uint64    // (unslid) address of entry point (_dyld_start) of dyld in cache
	ImagesTextOffset          uint64    // file offset to first dyld_cache_image_text_info
	ImagesTextCount           uint64    // number of dyld_cache_image_text_info entries
	PatchInfoAddr             uint64    // (unslid) address of dyld_cache_patch_info
	PatchInfoSize             uint64    // Size of all of the patch information pointed to via the dyld_cache_patch_info
	OtherImageGroupAddrUnused uint64    // unused
	OtherImageGroupSizeUnused uint64    // unused
	ProgClosuresAddr          uint64    // (unslid) address of list of program launch closures
	ProgClosuresSize          uint64    // size of list of program launch closures
	ProgClosuresTrieAddr      uint64    // (unslid) address of trie of indexes into program launch closures
	ProgClosuresTrieSize      uint64    // size of trie of indexes into program launch closures
	Platform                  uint32    // platform number (macOS=1, etc)
	BitField                  DYLDCacheHeaderV2BitField
	SharedRegionStart         uint64 // base load address of cache if not slid
	SharedRegionSize          uint64 // overall size required to map the cache and all subCaches, if any
	MaxSlide                  uint64 // runtime slide of cache can be between zero and this value
	DylibsImageArrayAddr      uint64 // (unslid) address of ImageArray for dylibs in this cache
	DylibsImageArraySize      uint64 // size of ImageArray for dylibs in this cache
	DylibsTrieAddr            uint64 // (unslid) address of trie of indexes of all cached dylibs
	DylibsTrieSize            uint64 // size of trie of cached dylib paths
	OtherImageArrayAddr       uint64 // (unslid) address of ImageArray for dylibs and bundles with dlopen closures
	OtherImageArraySize       uint64 // size of ImageArray for dylibs and bundles with dlopen closures
	OtherTrieAddr             uint64 // (unslid) address of trie of indexes of all dylibs and bundles with dlopen closures
	OtherTrieSize             uint64 // size of trie of dylibs and bundles with dlopen closures
	// End of V1 (not using an embedded struct to avoid losing the name and to use the right bitfield)
	MappingWithSlideOffset        uint32 // file offset to first dyld_cache_mapping_and_slide_info
	MappingWithSlideCount         uint32 // number of dyld_cache_mapping_and_slide_info entries
	DylibsPblStateArrayAddrUnused uint64 // unused
	DylibsPblSetAddr              uint64 // (unslid) address of PrebuiltLoaderSet of all cached dylibs
	ProgramsPblSetPoolAddr        uint64 // (unslid) address of pool of PrebuiltLoaderSet for each program
	ProgramsPblSetPoolSize        uint64 // size of pool of PrebuiltLoaderSet for each program
	ProgramTrieAddr               uint64 // (unslid) address of trie mapping program path to PrebuiltLoaderSet
	ProgramTrieSize               uint32
	OsVersion                     uint32    // OS Version of dylibs in this cache for the main platform
	AltPlatform                   uint32    // e.g. iOSMac on macOS
	AltOsVersion                  uint32    // e.g. 14.0 for iOSMac
	SwiftOptsOffset               uint64    // VM offset from cache_header* to Swift optimizations header
	SwiftOptsSize                 uint64    // size of Swift optimizations header
	SubCacheArrayOffset           uint32    // file offset to first dyld_subcache_entry
	SubCacheArrayCount            uint32    // number of subCache entries
	SymbolFileUuid                [16]uint8 // unique value for the shared cache file containing unmapped local symbols
	RosettaReadOnlyAddr           uint64    // (unslid) address of the start of where Rosetta can add read-only/executable data
	RosettaReadOnlySize           uint64    // maximum size of the Rosetta read-only/executable region
	RosettaReadWriteAddr          uint64    // (unslid) address of the start of where Rosetta can add read-write data
	RosettaReadWriteSize          uint64    // maximum size of the Rosetta read-write region
	ImagesOffset                  uint32    // file offset to first dyld_cache_image_info
	ImagesCount                   uint32    // number of dyld_cache_image_info entries
}

// Apple considers both those structs to be the same, as V2 simply extends V1
// They just tell the difference by checking if the extra fields collide with the mappings
// as they are supposed to be right after the header
func (me DYLDCacheHeaderV2) V1() (*DYLDCacheHeaderV1, bool) {
	isV1 := me.MappingOffset <= uint32(DYLDCacheHeaderV2SubCacheArrayOffsetOffset)
	if !isV1 {
		return nil, false
	}
	return (*DYLDCacheHeaderV1)(unsafe.Pointer(&me)), true
}

var DYLDCacheHeaderV2SubCacheArrayOffsetOffset = unsafe.Offsetof(DYLDCacheHeaderV2{}.SubCacheArrayOffset)

type DYLDCacheHeaderV2BitField uint32

func (me DYLDCacheHeaderV2BitField) FormatVersion() int {
	return int(me >> 24 & 0xFF) // dyld3::closure::kFormatVersion
}

func (me DYLDCacheHeaderV2BitField) DylibsExpectedOnDisk() bool {
	return me>>23&0x1 == 1 // dyld should expect the dylib exists on disk and to compare inode/mtime to see if cache is valid
}

func (me DYLDCacheHeaderV2BitField) Simulator() bool {
	return me>>22&0x1 == 1 // for simulator of specified platform
}

func (me DYLDCacheHeaderV2BitField) LocallyBuiltCache() bool {
	return me>>21&0x1 == 1 // 0 for B&I built cache, 1 for locally built cache
}

func (me DYLDCacheHeaderV2BitField) BuiltFromChainedFixups() bool {
	return me>>20&0x1 == 1 // 0 for B&I built cache, 1 for locally built cache
}

func (me DYLDCacheHeaderV2BitField) Padding() int {
	return int(me & 0xFFFFF) // padding
}

func (me DYLDCacheHeaderV2BitField) String() string {
	return fmt.Sprintf("{FormatVersion: %d, DylibsExpectedOnDisk: %t, Simulator: %t, LocallyBuiltCache: %t, BuiltFromChainedFixups: %t, Padding: %d}", me.FormatVersion(), me.DylibsExpectedOnDisk(), me.Simulator(), me.LocallyBuiltCache(), me.BuiltFromChainedFixups(), me.Padding())
}

type DYLDCacheHeaderV3 struct {
	DYLDCacheHeaderV2
	CacheSubType       uint32 // 0 for development, 1 for production, when cacheType is multi-cache(2)
	ObjcOptsOffset     uint64 // VM offset from cache_header* to ObjC optimizations header
	ObjcOptsSize       uint64 // size of ObjC optimizations header
	CacheAtlasOffset   uint64 // VM offset from cache_header* to embedded cache atlas for process introspection
	CacheAtlasSize     uint64 // size of embedded cache atlas
	DynamicDataOffset  uint64 // VM offset from cache_header* to the location of dyld_cache_dynamic_data_header
	DynamicDataMaxSize uint64 // maximum size of space reserved from dynamic data
}

// Same deal as for V1 vs V2, V3 just superseeds V2 without having a dedicated struct
func (me DYLDCacheHeaderV3) V2() (*DYLDCacheHeaderV2, bool) {
	isV2 := me.MappingOffset <= uint32(DYLDCacheHeaderV3CacheSubTypeOffset)
	if !isV2 {
		return nil, false
	}
	return (*DYLDCacheHeaderV2)(unsafe.Pointer(&me)), true
}

var DYLDCacheHeaderV3CacheSubTypeOffset = unsafe.Offsetof(DYLDCacheHeaderV3{}.CacheSubType)
