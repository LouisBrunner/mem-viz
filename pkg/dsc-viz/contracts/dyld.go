package contracts

import (
	"fmt"
	"io"
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

type Address interface {
	AddBase(base uintptr) Address
	Calculate(slide uint64) uintptr
	GetReader(cache Cache, offset, slide uint64) io.Reader
	Invalid() bool
}

type RelativeAddress32 uint32

func (me RelativeAddress32) AddBase(base uintptr) Address {
	return RelativeAddress64(uint64(base) + uint64(me))
}

func (me RelativeAddress32) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me RelativeAddress32) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me RelativeAddress32) Invalid() bool {
	return me == 0
}

type RelativeAddress64 uint64

func (me RelativeAddress64) AddBase(base uintptr) Address {
	return RelativeAddress64(uint64(base) + uint64(me))
}

func (me RelativeAddress64) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me RelativeAddress64) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me RelativeAddress64) Invalid() bool {
	return me == 0
}

type ManualAddress uint64

func (me ManualAddress) AddBase(base uintptr) Address {
	return ManualAddress(uint64(base) + uint64(me))
}

func (me ManualAddress) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me ManualAddress) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me ManualAddress) Invalid() bool {
	return false
}

type UnslidAddress uint64

func (me UnslidAddress) AddBase(base uintptr) Address {
	return me
}

func (me UnslidAddress) Calculate(slide uint64) uintptr {
	return uintptr(me) + uintptr(slide)
}

func (me UnslidAddress) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAbsolute(uint64(me) + offset + slide)
}

func (me UnslidAddress) Invalid() bool {
	return me == 0
}

// From Apple's `dyld-*/cache-builder/dyld_cache_format.h`

type DYLDCacheMappingInfo struct {
	Address    uint64 `struc:"little"`
	Size       uint64 `struc:"little"`
	FileOffset uint64 `struc:"little"`
	MaxProt    uint32 `struc:"little"`
	InitProt   uint32 `struc:"little"`
}

const (
	DYLD_CACHE_MAPPING_AUTH_DATA   = 1 << 0
	DYLD_CACHE_MAPPING_DIRTY_DATA  = 1 << 1
	DYLD_CACHE_MAPPING_CONST_DATA  = 1 << 2
	DYLD_CACHE_MAPPING_TEXT_STUBS  = 1 << 3
	DYLD_CACHE_DYNAMIC_CONFIG_DATA = 1 << 4
)

type DYLDCacheMappingAndSlideInfo struct {
	Address             uint64 `struc:"little"`
	Size                uint64 `struc:"little"`
	FileOffset          uint64 `struc:"little"`
	SlideInfoFileOffset uint64 `struc:"little"`
	SlideInfoFileSize   uint64 `struc:"little"`
	Flags               uint64 `struc:"little"`
	MaxProt             uint32 `struc:"little"`
	InitProt            uint32 `struc:"little"`
}

type DYLDCacheImageInfo struct {
	Address        UnslidAddress     `struc:"little"`
	ModTime        uint64            `struc:"little"`
	Inode          uint64            `struc:"little"`
	PathFileOffset RelativeAddress32 `struc:"little"`
	Pad            uint32            `struc:"little"`
}

type DYLDCacheImageInfoExtra struct {
	ExportsTrieAddr           uint64 `struc:"little"` // address of trie in unslid cache
	WeakBindingsAddr          uint64 `struc:"little"`
	ExportsTrieSize           uint32 `struc:"little"`
	WeakBindingsSize          uint32 `struc:"little"`
	DependentsStartArrayIndex uint32 `struc:"little"`
	ReExportsStartArrayIndex  uint32 `struc:"little"`
}

type DYLDCacheAcceleratorInfo struct {
	Version            uint32 `struc:"little"` // currently 1
	ImageExtrasCount   uint32 `struc:"little"` // does not include aliases
	ImagesExtrasOffset uint32 `struc:"little"` // offset into this chunk of first DYLDCacheImageInfoExtra
	BottomUpListOffset uint32 `struc:"little"` // offset into this chunk to start of 16-bit array of sorted image indexes
	DylibTrieOffset    uint32 `struc:"little"` // offset into this chunk to start of trie containing all dylib paths
	DylibTrieSize      uint32 `struc:"little"` // size of trie containing all dylib paths
	InitializersOffset uint32 `struc:"little"` // offset into this chunk to start of initializers list
	InitializersCount  uint32 `struc:"little"` // size of initializers list
	DofSectionsOffset  uint32 `struc:"little"` // offset into this chunk to start of DOF sections list
	DofSectionsCount   uint32 `struc:"little"` // size of initializers list
	ReExportListOffset uint32 `struc:"little"` // offset into this chunk to start of 16-bit array of re-exports
	ReExportCount      uint32 `struc:"little"` // size of re-exports
	DepListOffset      uint32 `struc:"little"` // offset into this chunk to start of 16-bit array of dependencies (0x8000 bit set if upward)
	DepListCount       uint32 `struc:"little"` // size of dependencies
	RangeTableOffset   uint32 `struc:"little"` // offset into this chunk to start of ss
	RangeTableCount    uint32 `struc:"little"` // size of dependencies
	DyldSectionAddr    uint64 `struc:"little"` // address of libdyld's __dyld section in unslid cache
}

type DYLDCacheAcceleratorInitializer struct {
	FunctionOffset uint32 `struc:"little"` // address offset from start of cache mapping
	ImageIndex     uint32 `struc:"little"`
}

type DYLDCacheRangeEntry struct {
	StartAddress uint64 `struc:"little"` // unslid address of start of region
	Size         uint32 `struc:"little"`
	ImageIndex   uint32 `struc:"little"`
}

type DYLDCacheAcceleratorDof struct {
	SectionAddress uint64 `struc:"little"` // unslid address of start of region
	SectionSize    uint32 `struc:"little"`
	ImageIndex     uint32 `struc:"little"`
}

type DYLDCacheImageTextInfo struct {
	UUID            [16]uint8
	LoadAddress     uint64 `struc:"little"`
	TextSegmentSize uint32 `struc:"little"`
	PathOffset      uint32 `struc:"little"`
}

type DYLDCacheSlideInfo struct {
	Version       uint32 `struc:"little"` // currently 1
	TocOffset     uint32 `struc:"little"`
	TocCount      uint32 `struc:"little"`
	EntriesOffset uint32 `struc:"little"`
	EntriesCount  uint32 `struc:"little"`
	EntriesSize   uint32 `struc:"little"` // currently 128
	// uint16_t toc[toc_count];
	// entrybitmap entries[entries_count];
}

type DYLDCacheSlideInfoEntry struct {
	Bits [4096 / (8 * 4)]uint8 // 128
}

type DYLDCacheSlideInfo2 struct {
	Version          uint32 `struc:"little"` // currently 2
	PageSize         uint32 `struc:"little"` // currently 4096 (may also be 16384)
	PageStartsOffset uint32
	PageStartsCount  uint32
	PageExtrasOffset uint32
	PageExtrasCount  uint32
	DeltaMask        uint64 `struc:"little"` // which (contiguous) set of bits contains the delta to the next rebase location
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
	Version         uint32 `struc:"little"` // currently 3
	PageSize        uint32 `struc:"little"` // currently 4096 (may also be 16384)
	PageStartsCount uint32 `struc:"little,sizeof=PageStarts"`
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
	Version          uint32 `struc:"little"` // currently 4
	PageSize         uint32 `struc:"little"` // currently 4096 (may also be 16384)
	PageStartsOffset uint32 `struc:"little"`
	PageStartsCount  uint32 `struc:"little"`
	PageExtrasOffset uint32 `struc:"little"`
	PageExtrasCount  uint32 `struc:"little"`
	DeltaMask        uint64 `struc:"little"` // which (contiguous) set of bits contains the delta to the next rebase location (0xC0000000)
	ValueAdd         uint64 `struc:"little"` // base address of cache
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
	NlistOffset   uint32 `struc:"little"` // offset into this chunk of nlist entries
	NlistCount    uint32 `struc:"little"` // count of nlist entries
	StringsOffset uint32 `struc:"little"` // offset into this chunk of string pool
	StringsSize   uint32 `struc:"little"` // byte count of string pool
	EntriesOffset uint32 `struc:"little"` // offset into this chunk of array of DYLDCacheLocalSymbolsEntry
	EntriesCount  uint32 `struc:"little"` // number of elements in DYLDCacheLocalSymbolsEntry array
}

type DYLDCacheLocalSymbolsEntry struct {
	DylibOffset     uint32 `struc:"little"` // offset in cache file of start of dylib
	NlistStartIndex uint32 `struc:"little"` // start index of locals for this dylib
	NlistCount      uint32 `struc:"little"` // number of local symbols for this dylib
}

type DYLDCacheLocalSymbolsEntry64 struct {
	DylibOffset     uint64 `struc:"little"` // offset in cache file of start of dylib
	NlistStartIndex uint32 `struc:"little"` // start index of locals for this dylib
	NlistCount      uint32 `struc:"little"` // number of local symbols for this dylib
}

type DYLDSubcacheEntryV1 struct {
	UUID          [16]uint8 // The UUID of the subCache file
	CacheVmOffset uint64    `struc:"little"` // The offset of this subcache from the main cache base address
}

type DYLDSubcacheEntryV2 struct {
	DYLDSubcacheEntryV1
	FileSuffix [32]byte // The file name suffix of the subCache file e.g. ".25.data", ".03.development"
}

// This struct is a small piece of dynamic data that can be included in the shared region, and contains configuration
// data about the shared cache in use by the process. It is located
type DYLDCacheDynamicDataHeader struct {
	Magic   [16]byte // e.g. "dyld_data    v0"
	FsId    uint64   `struc:"little"` // The fsid_t of the shared cache being used by a process
	FsObjId uint64   `struc:"little"` // The fs_obj_id_t of the shared cache being used by a process
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
	Magic                [16]byte                  // e.g. "dyld_v0    i386"
	MappingOffset        RelativeAddress32         `struc:"little"` // file offset to first DYLDCacheMappingInfo
	MappingCount         uint32                    `struc:"little"` // number of DYLDCacheMappingInfo entries
	ImagesOffset         RelativeAddress32         `struc:"little"` // file offset to first DYLDCacheImageInfo
	ImagesCount          uint32                    `struc:"little"` // number of DYLDCacheImageInfo entries
	DyldBaseAddress      uint64                    `struc:"little"` // base address of dyld when cache was built
	CodeSignatureOffset  uint64                    `struc:"little"` // file offset of code signature blob
	CodeSignatureSize    uint64                    `struc:"little"` // size of code signature blob (zero means to end of file)
	SlideInfoOffset      RelativeAddress64         `struc:"little"` // file offset of kernel slid info
	SlideInfoSize        uint64                    `struc:"little"` // size of kernel slid info
	LocalSymbolsOffset   RelativeAddress64         `struc:"little"` // file offset of where local symbols are stored
	LocalSymbolsSize     uint64                    `struc:"little"` // size of local symbols information
	UUID                 [16]uint8                 // unique value for each shared cache file
	CacheType            uint64                    `struc:"little"` // 0 for development, 1 for production
	BranchPoolsOffset    RelativeAddress32         `struc:"little"` // file offset to table of uint64_t pool addresses
	BranchPoolsCount     uint32                    `struc:"little"` // number of uint64_t entries
	AccelerateInfoAddr   UnslidAddress             `struc:"little"` // (unslid) address of optimization info
	AccelerateInfoSize   uint64                    `struc:"little"` // size of optimization info
	ImagesTextOffset     RelativeAddress64         `struc:"little"` // file offset to first DYLDCacheImageTextInfo
	ImagesTextCount      uint64                    `struc:"little"` // number of DYLDCacheImageTextInfo entries
	DylibsImageGroupAddr UnslidAddress             `struc:"little"` // (unslid) address of ImageGroup for dylibs in this cache
	DylibsImageGroupSize uint64                    `struc:"little"` // size of ImageGroup for dylibs in this cache
	OtherImageGroupAddr  UnslidAddress             `struc:"little"` // (unslid) address of ImageGroup for other OS dylibs
	OtherImageGroupSize  uint64                    `struc:"little"` // size of oImageGroup for other OS dylibs
	ProgClosuresAddr     UnslidAddress             `struc:"little"` // (unslid) address of list of program launch closures
	ProgClosuresSize     uint64                    `struc:"little"` // size of list of program launch closures
	ProgClosuresTrieAddr UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes into program launch closures
	ProgClosuresTrieSize uint64                    `struc:"little"` // size of trie of indexes into program launch closures
	Platform             uint32                    `struc:"little"` // platform number (macOS=1, etc)
	BitField             DYLDCacheHeaderV1BitField `struc:"little"`
	SharedRegionStart    uint64                    `struc:"little"` // base load address of cache if not slid
	SharedRegionSize     uint64                    `struc:"little"` // overall size of region cache can be mapped into
	MaxSlide             uint64                    `struc:"little"` // runtime slide of cache can be between zero and this value
	DylibsImageArrayAddr UnslidAddress             `struc:"little"` // (unslid) address of ImageArray for dylibs in this cache
	DylibsImageArraySize uint64                    `struc:"little"` // size of ImageArray for dylibs in this cache
	DylibsTrieAddr       UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes of all cached dylibs
	DylibsTrieSize       uint64                    `struc:"little"` // size of trie of cached dylib paths
	OtherImageArrayAddr  UnslidAddress             `struc:"little"` // (unslid) address of ImageArray for dylibs and bundles with dlopen closures
	OtherImageArraySize  uint64                    `struc:"little"` // size of ImageArray for dylibs and bundles with dlopen closures
	OtherTrieAddr        UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes of all dylibs and bundles with dlopen closures
	OtherTrieSize        uint64                    `struc:"little"` // size of trie of dylibs and bundles with dlopen closures
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
	Magic                     [16]byte                  // e.g. "dyld_v0    i386"
	MappingOffset             RelativeAddress32         `struc:"little"` // file offset to first DYLDCacheMappingInfo
	MappingCount              uint32                    `struc:"little"` // number of DYLDCacheMappingInfo entries
	ImagesOffsetOld           uint32                    `struc:"little"` // UNUSED: moved to imagesOffset to prevent older dsc_extarctors from crashing
	ImagesCountOld            uint32                    `struc:"little"` // UNUSED: moved to imagesCount to prevent older dsc_extarctors from crashing
	DyldBaseAddress           uint64                    `struc:"little"` // base address of dyld when cache was built
	CodeSignatureOffset       RelativeAddress64         `struc:"little"` // file offset of code signature blob
	CodeSignatureSize         uint64                    `struc:"little"` // size of code signature blob (zero means to end of file)
	SlideInfoOffsetUnused     uint64                    `struc:"little"` // unused.  Used to be file offset of kernel slid info
	SlideInfoSizeUnused       uint64                    `struc:"little"` // unused.  Used to be size of kernel slid info
	LocalSymbolsOffset        RelativeAddress64         `struc:"little"` // file offset of where local symbols are stored
	LocalSymbolsSize          uint64                    `struc:"little"` // size of local symbols information
	UUID                      [16]uint8                 // unique value for each shared cache file
	CacheType                 uint64                    `struc:"little"` // 0 for development, 1 for production, 2 for multi-cache
	BranchPoolsOffset         RelativeAddress32         `struc:"little"` // file offset to table of uint64_t pool addresses
	BranchPoolsCount          uint32                    `struc:"little"` // number of uint64_t entries
	DyldInCacheMh             UnslidAddress             `struc:"little"` // (unslid) address of mach_header of dyld in cache
	DyldInCacheEntry          UnslidAddress             `struc:"little"` // (unslid) address of entry point (_dyld_start) of dyld in cache
	ImagesTextOffset          RelativeAddress64         `struc:"little"` // file offset to first DYLDCacheImageTextInfo
	ImagesTextCount           uint64                    `struc:"little"` // number of DYLDCacheImageTextInfo entries
	PatchInfoAddr             UnslidAddress             `struc:"little"` // (unslid) address of DYLDCachePatchInfo
	PatchInfoSize             uint64                    `struc:"little"` // Size of all of the patch information pointed to via the DYLDCachePatchInfo
	OtherImageGroupAddrUnused uint64                    `struc:"little"` // unused
	OtherImageGroupSizeUnused uint64                    `struc:"little"` // unused
	ProgClosuresAddr          UnslidAddress             `struc:"little"` // (unslid) address of list of program launch closures
	ProgClosuresSize          uint64                    `struc:"little"` // size of list of program launch closures
	ProgClosuresTrieAddr      UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes into program launch closures
	ProgClosuresTrieSize      uint64                    `struc:"little"` // size of trie of indexes into program launch closures
	Platform                  uint32                    `struc:"little"` // platform number (macOS=1, etc)
	BitField                  DYLDCacheHeaderV2BitField `struc:"little"`
	SharedRegionStart         uint64                    `struc:"little"` // base load address of cache if not slid
	SharedRegionSize          uint64                    `struc:"little"` // overall size required to map the cache and all subCaches, if any
	MaxSlide                  uint64                    `struc:"little"` // runtime slide of cache can be between zero and this value
	DylibsImageArrayAddr      UnslidAddress             `struc:"little"` // (unslid) address of ImageArray for dylibs in this cache
	DylibsImageArraySize      uint64                    `struc:"little"` // size of ImageArray for dylibs in this cache
	DylibsTrieAddr            UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes of all cached dylibs
	DylibsTrieSize            uint64                    `struc:"little"` // size of trie of cached dylib paths
	OtherImageArrayAddr       UnslidAddress             `struc:"little"` // (unslid) address of ImageArray for dylibs and bundles with dlopen closures
	OtherImageArraySize       uint64                    `struc:"little"` // size of ImageArray for dylibs and bundles with dlopen closures
	OtherTrieAddr             UnslidAddress             `struc:"little"` // (unslid) address of trie of indexes of all dylibs and bundles with dlopen closures
	OtherTrieSize             uint64                    `struc:"little"` // size of trie of dylibs and bundles with dlopen closures
	// End of V1 (not using an embedded struct to avoid losing the name and to use the right bitfield)
	MappingWithSlideOffset        RelativeAddress32 `struc:"little"` // file offset to first DYLDCacheMappingAndSlideInfo
	MappingWithSlideCount         uint32            `struc:"little"` // number of DYLDCacheMappingAndSlideInfo entries
	DylibsPblStateArrayAddrUnused uint64            `struc:"little"` // unused
	DylibsPblSetAddr              UnslidAddress     `struc:"little"` // (unslid) address of PrebuiltLoaderSet of all cached dylibs
	ProgramsPblSetPoolAddr        UnslidAddress     `struc:"little"` // (unslid) address of pool of PrebuiltLoaderSet for each program
	ProgramsPblSetPoolSize        uint64            `struc:"little"` // size of pool of PrebuiltLoaderSet for each program
	ProgramTrieAddr               UnslidAddress     `struc:"little"` // (unslid) address of trie mapping program path to PrebuiltLoaderSet
	ProgramTrieSize               uint32            `struc:"little"`
	OsVersion                     uint32            `struc:"little"` // OS Version of dylibs in this cache for the main platform
	AltPlatform                   uint32            `struc:"little"` // e.g. iOSMac on macOS
	AltOsVersion                  uint32            `struc:"little"` // e.g. 14.0 for iOSMac
	SwiftOptsOffset               RelativeAddress64 `struc:"little"` // VM offset from cache_header* to Swift optimizations header
	SwiftOptsSize                 uint64            `struc:"little"` // size of Swift optimizations header
	SubCacheArrayOffset           RelativeAddress32 `struc:"little"` // file offset to first dyld_subcache_entry
	SubCacheArrayCount            uint32            `struc:"little"` // number of subCache entries
	SymbolFileUuid                [16]uint8         // unique value for the shared cache file containing unmapped local symbols
	RosettaReadOnlyAddr           UnslidAddress     `struc:"little"` // (unslid) address of the start of where Rosetta can add read-only/executable data
	RosettaReadOnlySize           uint64            `struc:"little"` // maximum size of the Rosetta read-only/executable region
	RosettaReadWriteAddr          UnslidAddress     `struc:"little"` // (unslid) address of the start of where Rosetta can add read-write data
	RosettaReadWriteSize          uint64            `struc:"little"` // maximum size of the Rosetta read-write region
	ImagesOffset                  RelativeAddress32 `struc:"little"` // file offset to first DYLDCacheImageInfo
	ImagesCount                   uint32            `struc:"little"` // number of DYLDCacheImageInfo entries
}

// Apple considers both those structs to be the same, as V2 simply extends V1
// They just tell the difference by checking if the extra fields collide with the mappings
// as they are supposed to be right after the header
func (me DYLDCacheHeaderV2) V1() (*DYLDCacheHeaderV1, bool) {
	isV1 := uint32(me.MappingOffset) <= uint32(DYLDCacheHeaderV2SubCacheArrayOffsetOffset)
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
	CacheSubType       uint32            `struc:"little"` // 0 for development, 1 for production, when cacheType is multi-cache(2)
	ObjcOptsOffset     RelativeAddress64 `struc:"big"`    // VM offset from cache_header* to ObjC optimizations header
	ObjcOptsSize       uint64            `struc:"big"`    // size of ObjC optimizations header
	CacheAtlasOffset   RelativeAddress64 `struc:"big"`    // VM offset from cache_header* to embedded cache atlas for process introspection
	CacheAtlasSize     uint64            `struc:"big"`    // size of embedded cache atlas
	DynamicDataOffset  RelativeAddress64 `struc:"big"`    // VM offset from cache_header* to the location of DYLDCacheDynamicDataHeader
	DynamicDataMaxSize uint64            `struc:"big"`    // maximum size of space reserved from dynamic data
}

// Same deal as for V1 vs V2, V3 just superseeds V2 without having a dedicated struct
func (me DYLDCacheHeaderV3) V2() (*DYLDCacheHeaderV2, bool) {
	isV2 := uint32(me.MappingOffset) <= uint32(DYLDCacheHeaderV3CacheSubTypeOffset)
	if !isV2 {
		return nil, false
	}
	return (*DYLDCacheHeaderV2)(unsafe.Pointer(&me)), true
}

var DYLDCacheHeaderV3CacheSubTypeOffset = unsafe.Offsetof(DYLDCacheHeaderV3{}.CacheSubType)

// From Apple's dyld/common/CachePatching.h

type DYLDCachePatchInfoV1 struct {
	PatchTableArrayAddr     UnslidAddress `struc:"little"` // (unslid) address of array for DYLDCacheImagePatches for each image
	PatchTableArrayCount    int64         `struc:"little"` // count of patch table entries
	PatchExportArrayAddr    UnslidAddress `struc:"little"` // (unslid) address of array for patch exports for each image
	PatchExportArrayCount   int64         `struc:"little"` // count of patch exports entries
	PatchLocationArrayAddr  UnslidAddress `struc:"little"` // (unslid) address of array for patch locations for each patch
	PatchLocationArrayCount int64         `struc:"little"` // count of patch location entries
	PatchExportNamesAddr    int64         `struc:"little"` // blob of strings of export names for patches
	PatchExportNamesSize    int64         `struc:"little"` // size of string blob of export names for patches
}

type DYLDCacheImagePatchesV1 struct {
	PatchExportsStartIndex uint32 `struc:"little"`
	PatchExportsCount      uint32 `struc:"little"`
}

type DYLDCachePatchableExportV1 struct {
	CacheOffsetOfImpl        uint32 `struc:"little"`
	PatchLocationsStartIndex uint32 `struc:"little"`
	PatchLocationsCount      uint32 `struc:"little"`
	ExportNameOffset         uint32 `struc:"little"`
}

type DYLDCachePatchableLocationV1 struct {
	CacheOffset uint32                               `struc:"little"`
	BitField    DYLDCachePatchableLocationV1BitField `struc:"little"`
}

func (me DYLDCachePatchableLocationV1) GetAddend() int64 {
	return (int64(me.BitField.Addend()) << 52) >> 52
}

type DYLDCachePatchableLocationV1BitField int32

func (me DYLDCachePatchableLocationV1BitField) High7() int {
	return int(me>>25) & 0x7F
}

func (me DYLDCachePatchableLocationV1BitField) Addend() int {
	return int(me>>20) & 0x1F
}

func (me DYLDCachePatchableLocationV1BitField) Authenticated() bool {
	return me>>19&0x1 == 1
}

func (me DYLDCachePatchableLocationV1BitField) UsesAddressDiversity() bool {
	return me>>18&0x1 == 1
}

func (me DYLDCachePatchableLocationV1BitField) Key() int {
	return int(me>>16) & 0x3
}

func (me DYLDCachePatchableLocationV1BitField) Discriminator() int {
	return int(me & 0xFFFF)
}

func (me DYLDCachePatchableLocationV1BitField) String() string {
	return fmt.Sprintf("{High7: %d, Addend: %d, Authenticated: %t, UsesAddressDiversity: %t, Key: %d, Discriminator: %d}", me.High7(), me.Addend(), me.Authenticated(), me.UsesAddressDiversity(), me.Key(), me.Discriminator())
}

// Patches can be different kinds.  This lives in the high nibble of the exportNameOffset,
// so we restrict these to 4-bits
const (
	// Just a normal patch. Isn't one of ther other kinds
	PATCH_KIND_REGULAR = 0x0
	// One of { void* isa, uintptr_t }, from CF
	PATCH_KIND_CFOBJ2 = 0x1
	// objc patching was added before this enum exists, in just the high bit
	// of the 4-bit nubble.  This matches that bit layout
	PATCH_KIND_OBJCCLASS = 0x8
)

// This is the base for all v2 and newer info
type DYLDCachePatchInfo struct {
	PatchTableVersion uint32 `struc:"little"` // == 2 or 3 for now
}

type DYLDCachePatchInfoV2 struct {
	DYLDCachePatchInfo
	PatchLocationVersion         uint32        `struc:"little"` // == 0 for now
	PatchTableArrayAddr          UnslidAddress `struc:"little"` // (unslid) address of array for DYLDCacheImagePatchesV2 for each image
	PatchTableArrayCount         int64         `struc:"little"` // count of patch table entries
	PatchImageExportsArrayAddr   UnslidAddress `struc:"little"` // (unslid) address of array for DYLDCacheImageExportV2 for each image
	PatchImageExportsArrayCount  int64         `struc:"little"` // count of patch table entries
	PatchClientsArrayAddr        UnslidAddress `struc:"little"` // (unslid) address of array for DYLDCacheImageClientsV2 for each image
	PatchClientsArrayCount       int64         `struc:"little"` // count of patch clients entries
	PatchClientExportsArrayAddr  UnslidAddress `struc:"little"` // (unslid) address of array for patch exports for each client image
	PatchClientExportsArrayCount int64         `struc:"little"` // count of patch exports entries
	PatchLocationArrayAddr       UnslidAddress `struc:"little"` // (unslid) address of array for patch locations for each patch
	PatchLocationArrayCount      int64         `struc:"little"` // count of patch location entries
	PatchExportNamesAddr         UnslidAddress `struc:"little"` // blob of strings of export names for patches
	PatchExportNamesSize         int64         `struc:"little"` // size of string blob of export names for patches
}

type DYLDCacheImagePatchesV2 struct {
	PatchClientsStartIndex uint32 `struc:"little"`
	PatchClientsCount      uint32 `struc:"little"`
	PatchExportsStartIndex uint32 `struc:"little"` // Points to DYLDCacheImageExportV2[]
	PatchExportsCount      uint32 `struc:"little"`
}

type DYLDCacheImageExportV2 struct {
	DylibOffsetOfImpl uint32                         `struc:"little"` // Offset from the dylib we used to find a DYLDCacheImagePatchesV2
	BitField          DYLDCacheImageExportV2BitField `struc:"little"`
}

type DYLDCacheImageExportV2BitField int32

func (me DYLDCacheImageExportV2BitField) ExportNameOffset() uint32 {
	return uint32(me & 0xFFFFFFF)
}

func (me DYLDCacheImageExportV2BitField) PatchKind() uint32 {
	return uint32(me >> 28 & 0xF) // One of DyldSharedCache::patchKind
}

func (me DYLDCacheImageExportV2BitField) String() string {
	return fmt.Sprintf("{ExportNameOffset: %d, PatchKind: %d}", me.ExportNameOffset(), me.PatchKind())
}

type DYLDCacheImageClientsV2 struct {
	ClientDylibIndex       uint32 `struc:"little"`
	PatchExportsStartIndex uint32 `struc:"little"` // Points to DYLDCachePatchableExportV2[]
	PatchExportsCount      uint32 `struc:"little"`
}

type DYLDCachePatchableExportV2 struct {
	ImageExportIndex         uint32 `struc:"little"` // Points to DYLDCacheImageExportV2
	PatchLocationsStartIndex uint32 `struc:"little"` // Points to DYLDCachePatchableLocationV2[]
	PatchLocationsCount      uint32 `struc:"little"`
}

type DYLDCachePatchableLocationV2 struct {
	DylibOffsetOfUse uint32                               `struc:"little"` // Offset from the dylib we used to get a DYLDCacheImageClientsV2
	BitField         DYLDCachePatchableLocationV1BitField `struc:"little"`
}

func (me DYLDCachePatchableLocationV2) GetAddend() int64 {
	return (int64(me.BitField.Addend()) << 52) >> 52
}

type DYLDCachePatchInfoV3 struct {
	DYLDCachePatchInfoV2
	// uint32_t    patchTableVersion;       // == 3
	// ... other fields from DYLDCachePatchInfoV2
	GotClientsArrayAddr        UnslidAddress `struc:"little"` // (unslid) address of array for DYLDCacheImageGotClientsV3 for each image
	GotClientsArrayCount       int64         `struc:"little"` // count of got clients entries.  Should always match the patchTableArrayCount
	GotClientExportsArrayAddr  UnslidAddress `struc:"little"` // (unslid) address of array for patch exports for each GOT image
	GotClientExportsArrayCount int64         `struc:"little"` // count of patch exports entries
	GotLocationArrayAddr       UnslidAddress `struc:"little"` // (unslid) address of array for patch locations for each GOT patch
	GotLocationArrayCount      int64         `struc:"little"` // count of patch location entries
}

type DYLDCacheImageGotClientsV3 struct {
	PatchExportsStartIndex uint32 `struc:"little"` // Points to DYLDCachePatchableExportV3[]
	PatchExportsCount      uint32 `struc:"little"`
}

type DYLDCachePatchableExportV3 struct {
	ImageExportIndex         uint32 `struc:"little"` // Points to DYLDCacheImageExportV2
	PatchLocationsStartIndex uint32 `struc:"little"` // Points to DYLDCachePatchableLocationV3[]
	PatchLocationsCount      uint32 `struc:"little"`
}

type DYLDCachePatchableLocationV3 struct {
	CacheOffsetOfUse int64                                `struc:"little"` // Offset from the cache header
	BitField         DYLDCachePatchableLocationV1BitField `struc:"little"`
}

func (me DYLDCachePatchableLocationV3) GetAddend() int64 {
	return (int64(me.BitField.Addend()) << 52) >> 52
}
