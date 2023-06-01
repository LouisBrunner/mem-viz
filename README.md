# `mem-viz`

This tool allows to display computer memory maps using various formats. It is intended as both a generic way to display memory from any system, language, library, etc (back-end) or as a way to read/parse specific memory maps and display them (front-end).

## Installation

```sh
go install github.com/LouisBrunner/mem-viz/cmd/mem-viz@latest
```

## Usage

```
Usage of mem-viz:
      --from-json ./blocks.json          use the JSON output from a previous run, e.g. ./blocks.json or `-` for stdin
      --from-json-text {"Name": "foo"}   use the JSON output from a previous run, e.g. {"Name": "foo"}
  -h, --help                             show this help message and exit
      --logging-level string             logrus log level for internal debugging, e.g. "debug" (default "error")
      --output string                    output format, one of: "graphviz", "latex", "markdown", "text", "ascii", "json" (default "text")
  -o, --output-file ./blocks.dot         output file, e.g. ./blocks.dot, defaults to stdout
```

In order to use `mem-viz` directly, you will need to provide a JSON document. See [Front-ends > JSON](#json) for more information.

## Front-ends

These are the possible ways to input memory maps for `mem-viz` to render.

### JSON

A JSON document saved from a previous run of a front-end, generated by another tool or handcrafted can be used as input.

Here is a rough schema of the format (see [here](pkg/contracts/memory.go) for full details):

```json
{
  "Name": "DSC",
  "Address": 7314554880,
  "Size": 0,
  "ParentOffset": 0,
  "Content": [
    {
      "Name": "Main Header Area",
      "Address": 7314554880,
      "Size": 0,
      "ParentOffset": 0,
      "Content": [
        {
          "Name": "Main Header (V3)",
          "Address": 7314554880,
          "Size": 512,
          "ParentOffset": 0,
          "Content": null,
          "Values": [
            {
              "Name": "Magic",
              "Offset": 0,
              "Size": 16,
              "Value": "dyld_v1  arm64e",
              "Links": null
            },
            {
              "Name": "MappingOffset",
              "Offset": 16,
              "Size": 4,
              "Value": "512 (0x200)",
              "Links": [
                {
                  "Name": "points to",
                  "TargetAddress": 7314555392
                }
              ]
            }
          ]
        },
        {
          "Name": "Mappings (6)",
          "Address": 7314555392,
          "Size": 0,
          "ParentOffset": 512,
          "Content": [
            {
              "Name": "Mapping 1/6",
              "Address": 7314555392,
              "Size": 32,
              "ParentOffset": 0,
              "Content": null,
              "Values": [
                {
                  "Name": "Address",
                  "Offset": 0,
                  "Size": 8,
                  "Value": "6442450944 (0x180000000)",
                  "Links": null
                },
                {
                  "Name": "Size",
                  "Offset": 8,
                  "Size": 8,
                  "Value": "1414856704 (0x54550000)",
                  "Links": null
                },
                {
                  "Name": "FileOffset",
                  "Offset": 16,
                  "Size": 8,
                  "Value": "0 (0x0)",
                  "Links": null
                },
                {
                  "Name": "MaxProt",
                  "Offset": 24,
                  "Size": 4,
                  "Value": "5 (0x5)",
                  "Links": null
                },
                {
                  "Name": "InitProt",
                  "Offset": 28,
                  "Size": 4,
                  "Value": "5 (0x5)",
                  "Links": null
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### `dsc-viz`

This tool allows to display the format of a macOS/iOS dyld shared cache (DSC) file. It is also able to read the kernel-provided DSC in its memory.

Install it using:

```sh
go install github.com/LouisBrunner/mem-viz/cmd/dsc-viz@latest
```

Usage:

```
Usage of dsc-viz:
      --from-arch string                 architecture of the file to load
      --from-current-arch                load the file for the current architecture
      --from-file string                 file to load
      --from-json ./blocks.json          use the JSON output from a previous run, e.g. ./blocks.json or `-` for stdin
      --from-json-text {"Name": "foo"}   use the JSON output from a previous run, e.g. {"Name": "foo"}
      --from-memory                      load the memory from the current process
  -h, --help                             show this help message and exit
      --logging-level string             logrus log level for internal debugging, e.g. "debug" (default "error")
      --output string                    output format, one of: "graphviz", "latex", "markdown", "text", "ascii", "json" (default "text")
  -o, --output-file ./blocks.dot         output file, e.g. ./blocks.dot, defaults to stdout
```

You can use `--from-memory` or `--from-current-arch` to let the tool fetch the DSC from your system (respectively from memory or from a file on disk). Otherwise you can use `--from-file` to specify a file to fetch from or `--from-arch` to scan your system but for a specific architecture.

Other options are the same as `mem-viz` (same output formats supported, possibility to save/load JSON, etc).

### `macho-viz`

This tool allows to display the format of a macOS/iOS Mach-O file.

Install it using:

```sh
go install github.com/LouisBrunner/mem-viz/cmd/macho-viz@latest
```

Usage:

```
Usage of dsc-viz:
      --file string                      file to load
      --from-json ./blocks.json          use the JSON output from a previous run, e.g. ./blocks.json or `-` for stdin
      --from-json-text {"Name": "foo"}   use the JSON output from a previous run, e.g. {"Name": "foo"}
  -h, --help                             show this help message and exit
      --logging-level string             logrus log level for internal debugging, e.g. "debug" (default "error")
      --output string                    output format, one of: "graphviz", "latex", "markdown", "text", "ascii", "json" (default "text")
  -o, --output-file ./blocks.dot         output file, e.g. ./blocks.dot, defaults to stdout
```

You can use `--file` to specify a file to read from disk.

Other options are the same as `mem-viz` (same output formats supported, possibility to save/load JSON, etc).

## Output formats

A wide-range of output formats is supported.

### `text` (default)

It will display the memory map in a linear, human-readable format.

Example (partial output of DSC):

```
0x00000001b3fdaa00-0x00000001b3fdaa20 [  32 B]       Images Text 2450/2458 {UUID:61b5e92a-e5d5-3280-889c-7d02e46bf544,LA:0x211329000,TSS:0x1fff,PO:0x570c9}
0x00000001b3fdaa20-0x00000001b3fdaa40 [  32 B]       Images Text 2451/2458 {UUID:6093beae-8e58-333f-852b-3d90f389ddbf,LA:0x21132b000,TSS:0x7ff6,PO:0x57100}
0x00000001b3fdaa40-0x00000001b3fdaa60 [  32 B]       Images Text 2452/2458 {UUID:9adf4c59-f0cb-3698-b058-8131864f9d6e,LA:0x211333000,TSS:0x3000,PO:0x57138}
0x00000001b3fdaa60-0x00000001b3fdaa80 [  32 B]       Images Text 2453/2458 {UUID:a06d4152-36d8-328c-b3eb-ff12aa211dfb,LA:0x211336000,TSS:0x1000,PO:0x57171}
0x00000001b3fdaa80-0x00000001b3fdaaa0 [  32 B]       Images Text 2454/2458 {UUID:f2f30ca7-222b-313b-b1ab-df9899bf6342,LA:0x211337000,TSS:0x3ffc,PO:0x571af}
0x00000001b3fdaaa0-0x00000001b3fdaac0 [  32 B]       Images Text 2455/2458 {UUID:68c2b690-d94f-39da-9efb-c4245d6bacd4,LA:0x21133b000,TSS:0x2000,PO:0x571e7}
0x00000001b3fdaac0-0x00000001b3fdaae0 [  32 B]       Images Text 2456/2458 {UUID:4509ebe4-44e9-3d39-80a4-5661ccdf9df2,LA:0x21133d000,TSS:0x1ffa,PO:0x57220}
0x00000001b3fdaae0-0x00000001b3fdab00 [  32 B]       Images Text 2457/2458 {UUID:9cc9f303-f02c-390f-b0ba-f895d6f63701,LA:0x21133f000,TSS:0x4ffc,PO:0x57258}
0x00000001b3fdab00-0x00000001b3fdab20 [  32 B]       Images Text 2458/2458 {UUID:05e52166-6c40-3639-a636-7ff7d3989b57,LA:0x211344000,TSS:0xbfff,PO:0x5728e}
0x00000001b3fdab20-0x00000001b3fdab58 [  56 B]     Subcache Entries (1) <- Main Header (V3).SubCacheArrayOffset, Main Header (V3).SubCacheArrayCount
0x00000001b3fdab20-0x00000001b3fdab58 [  56 B]       Subcache Entry (V2) {UUID:1d33c45e-9da4-35c1-beeb-4dcaa582d0ca,CVO:0x63444000,FS:.01}
0x00000001b3fdab58-0x00000001b4bc1eca [ 12 MB]     UNUSED
0x00000001b4bc1eca-0x00000001b4fc1eca [4.2 MB]     DYLD Cache Dynamic Data <- Main Header (V3).DynamicDataOffset, Main Header (V3).DynamicDataMaxSize
0x00000001b4fc1eca-0x00000002133f8000 [1.6 GB]     UNUSED
0x00000002133f8000-0x00000002136f4000 [3.1 MB]     Code Signature <- Main Header (V3).CodeSignatureOffset, Main Header (V3).CodeSignatureSize
0x00000002136f4000-0x00000002173f8000 [ 64 MB]   UNUSED
0x00000002173f8000-0x000000027b0c0000 [1.7 GB]   Sub Cache .01 Area <- Subcache Entry (V2).CacheVmOffset, Sub Cache .01 (V3).DyldInCacheMh, Sub Cache .01 (V3).DyldInCacheEntry, Sub Cache .01 (V3).DylibsPblSetAddr
0x00000002173f8000-0x00000002173f8200 [ 512 B]     Sub Cache .01 (V3) {M:dyld_v1  arm64e,MO:0x200,MC:0x6,IOO:0x0,ICO:0x0,DBA:0x0,CSO:0x639a8000,CSS:0x320000,SIOU:0x0,SISU:0x0,LSO:0x0,LSS:0x0,UUID:1d33c45e-9da4-35c1-beeb-4dcaa582d0ca,CT:0x0,BPO:0x0,BPC:0x0,DICM:0x0,DICE:0x0,ITO:0x137e0,ITC:0x99a,PIA:0x0,PIS:0x0,OIGAU:0x0,OIGSU:0x0,PCA:0x0,PCS:0x0,PCTA:0x0,PCTS:0x0,P:0x1,BF:{FormatVersion: 0, DylibsExpectedOnDisk: false, Simulator: false, LocallyBuiltCache: false, BuiltFromChainedFixups: false, Padding: 0},SRS:0x1e3444000,SRS:0x0,MS:0x0,DIAA:0x0,DIAS:0x0,DTA:0x0,DTS:0x0,OIAA:0x0,OIAS:0x0,OTA:0x0,OTS:0x0,MWSO:0x350,MWSC:0x6,DPSAAU:0x0,DPSA:0x0,PPSPA:0x0,PPSPS:0x0,PTA:0x0,PTS:0x0,OV:0x0,AP:0x0,AOV:0x0,SOO:0x0,SOS:0x0,SCAO:0x26b20,SCAC:0x0,SFU:,RROA:0x0,RROS:0x0,RRWA:0x0,RRWS:0x0,IO:0x4a0,IC:0x99a,CST:0x1,OOO:0x0,OOS:0x0,CAO:0x0,CAS:0x0,DDO:0x0,DDMS:0x0}
0x00000002173f8200-0x00000002173f82c0 [ 192 B]     Mappings (6) <- Sub Cache .01 (V3).MappingOffset, Sub Cache .01 (V3).MappingCount
0x00000002173f8200-0x00000002173f8220 [  32 B]       Mappings 1/6 {A:0x1e3444000,S:0x2e0e8000,FO:0x0,MP:0x5,IP:0x5}
0x00000002173f8220-0x00000002173f8240 [  32 B]       Mappings 2/6 {A:0x21152c000,S:0x1184000,FO:0x2e0e8000,MP:0x3,IP:0x1}
0x00000002173f8240-0x00000002173f8260 [  32 B]       Mappings 3/6 {A:0x2146b0000,S:0x2268000,FO:0x2f26c000,MP:0x3,IP:0x3}
0x00000002173f8260-0x00000002173f8280 [  32 B]       Mappings 4/6 {A:0x216918000,S:0x1660000,FO:0x314d4000,MP:0x3,IP:0x3}
0x00000002173f8280-0x00000002173f82a0 [  32 B]       Mappings 5/6 {A:0x217f78000,S:0x140c000,FO:0x32b34000,MP:0x3,IP:0x1}
0x00000002173f82a0-0x00000002173f82c0 [  32 B]       Mappings 6/6 {A:0x21b384000,S:0x2fa68000,FO:0x33f40000,MP:0x1,IP:0x1}
0x00000002173f82c0-0x00000002173f8350 [ 144 B]     UNUSED
0x00000002173f8350-0x00000002173f84a0 [ 336 B]     Mappings With Slide (6) <- Sub Cache .01 (V3).MappingWithSlideOffset, Sub Cache .01 (V3).MappingWithSlideCount
0x00000002173f8350-0x00000002173f8388 [  56 B]       Mappings With Slide 1/6 {A:0x1e3444000,S:0x2e0e8000,FO:0x0,SIFO:0x0,SIFS:0x0,F:0x0,MP:0x5,IP:0x5}
0x00000002173f8388-0x00000002173f83c0 [  56 B]       Mappings With Slide 2/6 {A:0x21152c000,S:0x1184000,FO:0x2e0e8000,SIFO:0x33f444e8,SIFS:0x2320,F:0x4,MP:0x3,IP:0x1}
```

Notes:

- Block values are abbreviated for readability
- Links are grouped if they points to the same value
- `UNUSED` blocks are added where no memory has been mapped
- `UNUSED` blocks are broken up if a link happens in the middle (to show where missing mappings might be)

### `ascii`

TODO: unimplemented

### `graphviz`

TODO: unimplemented

### `markdown`

TODO: unimplemented

### `latex`

TODO: unimplemented, see memory-maps in https://texdoc.org/serve/bytefield.pdf/0

### `json`

When using `mem-viz` directly, outputting `json` will just act as a noop: whatever you passed in will be echoed back. Do note that it will parse your input and check it, removing any unknown field. This is a good way to check if JSON format is valid though.

When using another front-end, it will output the JSON representation of the memory map, which can be saved then loaded later with one of the `--from-json*` flags.
