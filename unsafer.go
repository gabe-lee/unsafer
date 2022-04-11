package unsafer

import (
	"unsafe"
)

/*********************************************************************************
	THE FOLLOWING TYPES AND CONSTANTS ARE ANALOGOUS WITH GO'S INTERNAL SOURCE CODE
	AND SHOULD BE TREATED AS SUCH FOR LICENSING AND REDISTRIBUTION PURPOSES
*********************************************************************************/

// Size of a pointer for the target architecture
//
// Unsafety Rating: ☆☆☆☆☆ (perfectly safe)
const SystemPointerSize = 4 << (^uintptr(0) >> 63)

// Internal structure of a string
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
type StringInternal struct {
	Data unsafe.Pointer // Pointer to beginning of string data
	Len  int            // Length of data in bytes
}

// Internal structure of a slice
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
type SliceInternal struct {
	Data unsafe.Pointer // Pointer to beginning of slice data
	Len  int            // Length of slice
	Cap  int            // Capacity of slice
}

// Additional info about the type
type TypeFlag uint8

const (
	TFlagUncommon      TypeFlag = 1 << 0 // ??
	TFlagExtraStar     TypeFlag = 1 << 1 // Whether the Name field has an extra superfluous star in front of it
	TFlagNamed         TypeFlag = 1 << 2 // Type has a defined name
	TFlagRegularMemory TypeFlag = 1 << 3 // Whether the type can be treated in its entirety as contiguous block of Size bytes
)

type NameOffset int32 // int32 offset from specific TypeInternal pointer to its string name
type TypeOffset int32 // int32 offset from specific TypeInternal pointer to the type that is a POINTER-TO the type

const (
	NameExported                       byte = 1
	NameFollowedByTagData              byte = 2
	TagDataFollowedByPkgPathNameOffset byte = 4
)

// Encoded type name with additional data.
//
// First byte has flags describing the name, followed by varint-encoded length,
// followed by the name itself.
//
// If NameFollowedByTagData is set, the name is followed by
// a varint-encoded length followed by the tag data itself.
//
// If TagDataFollowedByPkgPathNameOffset is set, tag data is followed by a NameOffset
// (int32)
//
// Unsafety Rating: ★★★☆☆ (clearly unsafe)
type EncodedName struct {
	Bytes *byte // pointer to first byte of encoded name
}

// Internals of Go's type system for most types
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type TypeInternal struct {
	Size        uintptr                                   // How large the type is. Does not include the size of any data POINTED TO by this type or any fields/elements of this type
	PtrData     uintptr                                   // size of memory prefix holding all pointers
	Hash        uint32                                    // Precomputed hash of this type
	TypeFlags   TypeFlag                                  // Additional type data
	Align       uint8                                     // Byte alignment of a variable of this type
	FieldAlign  uint8                                     // Byte alignment of a struct field of this type
	kind        Kind                                      // Base category this type falls under
	Equals      func(unsafe.Pointer, unsafe.Pointer) bool // Function for comparing equality between two variables of this type
	GCData      *byte                                     // GarbageCollectionData: for the truly insane, see src/runtime/mbitmap.go
	Name        NameOffset                                // Offset to the plain-text name for this type as a string
	Pointertype TypeOffset                                // Offset to the type that is a POINTER-TO this type (*T)
}

// Flags for special map states
type MapFlag uint8

const (
	BeingUsedByIterator    MapFlag = 1 // An iterator may be using the currrent buckets
	OldBeingUsedByIterator MapFlag = 2 // An iterator may be using the old buckets
	BeingWrittenTo         MapFlag = 4 // A goroutine is writing to the map
	GrowingToSameSize      MapFlag = 8 // The current grow operation is growing to a map of the same size
)

// Internal structure of a map
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type MapInternal struct {
	Count          int            // Number of Key-Value pairs currently active
	Flags          MapFlag        // Flags for special map states
	NumBucketsLog2 uint8          // Log (base 2) of the number of buckets
	NumOverflow    uint16         // (Approximate) Number of overflow buckets
	HashSeed       uint32         // Seed for the hashing algorithm
	Buckets        unsafe.Pointer // Bucket array with length = NumBucketsLog2^2, may be nil if Count == 0
	OldBuckets     unsafe.Pointer // Old bucket array of half the current size, only non-nil when in the process of growing
	NumEvacuated   uintptr        // progress counter for evacuation (buckets less than this have been evacuated)
	MapOverflow    *MapOverflow   // Holds pointers to overflow buckets for map types that require them
}

// Holds pointers to overflow buckets for map types that require them
//
// Unsafety Rating: ★★★☆☆ (clearly unsafe)
type MapOverflow struct {
	OverflowBuckets    *[]*BucketInternal // Pointer to master list of current overflow buckets to keep them alive
	OldOverflowBuckets *[]*BucketInternal // Pointer to master list of old overflow buckets to keep them alive
	NextOverflowBucket *BucketInternal    // Next free overflow bucket
}

// How many Key/Value pairs a bucket can hold
//
// Unsafety Rating: ☆☆☆☆☆ (perfectly safe)
const BucketSize = 8

// The offset from a bucket's location in memory to where its Key/Value pairs begin
//
// Unsafety Rating: ★★★★☆ (highly dangerous)
const BucketDataStart = unsafe.Offsetof(struct {
	b BucketInternal
	v int64
}{}.v)

const (
	LastEmptyCell         uint8 = 0 // Special TopHash: This cell is empty, and there are no more non-empty cells after this one.
	EmptyCell             uint8 = 1 // Special TopHash: This cell is empty
	EvacuatedToFirstHalf  uint8 = 2 // Special TopHash: Key/Value pair is valid, but it has been evacuated to the first half of a larger bucket
	EvacuatedToSecondHalf uint8 = 3 // Special TopHash: Key/Value pair is valid, but it has been evacuated to the second half of a larger bucket
	EvacuatedAndEmpty     uint8 = 4 // Special TopHash: This cell is empty and the entire bucket is evacuated
	MinimumTopHash        uint8 = 5 // Minimum TopHash value for a normal, non-evacuated cell
)

// Internals of a map bucket.
// Immediately following the bucket's place in memory are 8(BucketSize) keys then 8(BucketSize) values,
// followed by a pointer to an overflow bucket
//
// Unsafety Rating: ★★★☆☆ (clearly unsafe)
type BucketInternal struct {
	// Normally holds the top byte of the hash value for each key in the bucket,
	// or a special value for empty cells or evacuation state.
	TopHash [BucketSize]uint8
}

// Internals of an interface that defines methods
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type InterfaceInternal struct {
	IDescription *InterfaceDescription // Description of the interface
	Data         unsafe.Pointer        // Pointer to the concrete data
}

// Description of an interface that defines methods
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type InterfaceDescription struct {
	IType *ITypeInternal // The type of the interface definition itself
	Type  *TypeInternal  // The type of the concrete data held by the interface
	Hash  uint32         // Same as Type.Hash
	_     [4]byte        // Padding
	// Pointers to the concrete functions the interface describes.
	// The size of this array is variable, contrary to what is listed.
	// If first index == 0, Type does not implement IType
	FunctionPointers [1]uintptr
}

// TypeInternal wrapper for interfaces
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type ITypeInternal struct {
	Type         TypeInternal      // Basic type data for the interface
	PackagePath  EncodedName       // An EncodedName describing the package path of the interface
	MethodHeader []InterfaceMethod // A list of method types the interface has
}

// Type describing the type of a method on an interface
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type InterfaceMethod struct {
	Name NameOffset // Offset pointing to the name of the function
	Type TypeOffset // Offset pointing to the type of the function
}

// Internals of 'any' (also known as the empty interface, interface{})
//
// Unsafety Rating: ★★☆☆☆ (use caution)
type AnyInternal struct {
	Type *TypeInternal  // Pointer to the TYPE of the concrete data
	Data unsafe.Pointer // Pointer to the concrete data
}

type Kind uint8

const (
	KindBool Kind = 1 + iota
	KindInt
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindUintptr
	KindFloat32
	KindFloat64
	KindComplex64
	KindComplex128
	KindArray
	KindChan
	KindFunc
	KindInterface
	KindMap
	KindPointer
	KindSlice
	KindString
	KindStruct
	KindUnsafePointer

	KindDirectIface Kind = 1 << 5       // Whether the type is stored directly in an interface
	KindGCProg      Kind = 1 << 6       // Whether the value pointed to by TypeInternal.GCData is a GCProgram
	KindMask        Kind = (1 << 5) - 1 // Mask for base kinds without special flags
)

// NoEscape hides a pointer from escape analysis. NoEscape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  NoEscape is inlined and currently
// compiles down to zero instructions.
//
// Unsafety Rating: ★★★☆☆ (clearly unsafe)
//go:nosplit
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

/*********************************************************************************
	THE FOLLOWING FUNCTIONS AND TYPES ARE ADDED BY THE AUTHOR TO MAKE USE OF THE
	ABOVE TYPES IN NEW, INTERESTING, AND POSSIBLY UNSAFER WAYS, LICENSED UNDER
	THE PERMISIVE BSD 2-CLAUSE LICENSE.
*********************************************************************************/

// Return the unique type pointer of the supplied value.
// This is THE definitive address where the type's definition resides,
// and will not change for the duration of the program.
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
func GetTypePointer(t any) (pointer uintptr) {
	tt := (*AnyInternal)(unsafe.Pointer(&t))
	return uintptr(unsafe.Pointer(tt.Type))
}

// Return the unique type hash of the supplied value.
// In most cases the hash is enough to uniqely identify a type, but
// collisions may exists. This value will not change
// for a given type for the duration of the program.
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
func GetTypeHash(t any) (hash uint32) {
	tt := (*AnyInternal)(unsafe.Pointer(&t))
	return tt.Type.Hash
}

// Get the basic kind of variable this type embodies
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
func GetKind(t any) Kind {
	tt := (*AnyInternal)(unsafe.Pointer(&t))
	return tt.Type.kind & KindMask
}

// Tell Go that t1 is *actually* of the same type as t2, as in
// a type assertion on the returned 'any' value will resolve to
// the same type as t2 using the data pointed to by t1.
//
// Unsafety Rating: ★★★★☆ (highly dangerous)
func Spoof(t1 any, t2 any) any {
	tt := (*AnyInternal)(unsafe.Pointer(&t1))
	tt.Type = (*TypeInternal)(unsafe.Pointer(GetTypePointer(t2)))
	return t1
}

// Invent an 'any' value from the memory pointed to by data,
// and the type located at typePointer. Use GetTypePointer(t any) to
// find type pointer addresses.
//
// Unsafety Rating: ★★★★★ (C U R S E D)
func Invent(data unsafe.Pointer, typePointer uintptr) (value any) {
	a := (*AnyInternal)(unsafe.Pointer(&value))
	a.Data = data
	a.Type = (*TypeInternal)(unsafe.Pointer(typePointer))
	return value
}

// Return a string that mirrors the data in slice,
// with a fixed length matching the length of the slice at time of calling.
//
// Any changes to the bytes in range slice[0:lenAtCallTime] will be reflected by the string
// in future reads.
//
// Assignment to the resulting string, assignment to the byte slice,
// or an append call that reallocates the byte slice
// will break the relationship, however since string still has pointer to
// old data it will persist in the state it was in prior to breaking.
//
// Unsafety Rating: ★☆☆☆☆ (relatively safe)
func ByteString(slice []byte) string {
	return *(*string)(unsafe.Pointer(&slice))
}
