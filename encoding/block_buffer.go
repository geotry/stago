package encoding

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/geotry/rass/compute"
)

type WritableBlock interface {
	NewBlock(kind uint8) int
	EndBlock() *Block
	AppendBlock(b *Block)
	PutBool(v bool) int
	PutUint8(v uint8) int
	PutUint16(v uint16) int
	PutUint32(v uint32) int
	PutVector2Float32(x, y float32) int
	PutVector3Float32(x, y, z float32) int
	PutMatrix(m compute.Matrix) int
	NewArray()
	EndArray()
}

// BlockBuffer wraps a buffer to write contiguous blocks of data.
type BlockBuffer struct {
	buf                []byte
	capacity           int
	offset             int
	blocks             map[int]uint8
	currentBlockOffset int
	currentBlockSize   int
	currentArrayOffset int
	currentArraySize   int
}

// Block is a fixed-space in buffer that can be updated independantly
type Block struct {
	kind        uint8
	startOffset int
	offset      int
	size        int
	buf         *BlockBuffer
}

// 1 uint8 for type
// 1 uint32 for byte size
const BlockPrefixBytes = 5

func NewBlockBuffer(cap int) *BlockBuffer {
	return &BlockBuffer{
		buf:                make([]byte, cap),
		capacity:           cap,
		offset:             0,
		blocks:             make(map[int]uint8),
		currentBlockOffset: 0,
	}
}

func (b *BlockBuffer) PutUint8(v uint8) int {
	b.buf[b.offset] = v
	b.moveOffset(1)
	return b.offset
}

func (b *Block) PutUint8(v uint8) int {
	b.buf.buf[b.startOffset+b.offset] = v
	b.moveOffset(1)
	return b.offset
}

func (b *BlockBuffer) PutBool(v bool) int {
	if v {
		b.buf[b.offset] = 1
	} else {
		b.buf[b.offset] = 0
	}
	b.moveOffset(1)
	return b.offset
}

func (b *Block) PutBool(v bool) int {
	if v {
		b.buf.buf[b.startOffset+b.offset] = 1
	} else {
		b.buf.buf[b.startOffset+b.offset] = 0
	}
	b.moveOffset(1)
	return b.offset
}

func (b *BlockBuffer) PutUint16(v uint16) int {
	binary.BigEndian.PutUint16(b.buf[b.offset:], v)
	b.moveOffset(2)
	return b.offset
}

func (b *Block) PutUint16(v uint16) int {
	binary.BigEndian.PutUint16(b.buf.buf[b.startOffset+b.offset:], v)
	b.moveOffset(2)
	return b.offset
}

func (b *BlockBuffer) PutUint32(v uint32) int {
	binary.BigEndian.PutUint32(b.buf[b.offset:], v)
	b.moveOffset(4)
	return b.offset
}

func (b *Block) PutUint32(v uint32) int {
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset:], v)
	b.moveOffset(4)
	return b.offset
}

func (b *BlockBuffer) PutVector2Float32(x, y float32) int {
	binary.BigEndian.PutUint32(b.buf[b.offset:], math.Float32bits(x))
	binary.BigEndian.PutUint32(b.buf[b.offset+4:], math.Float32bits(y))
	b.moveOffset(8)
	return b.offset
}

func (b *Block) PutVector2Float32(x, y float32) int {
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset:], math.Float32bits(x))
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset+4:], math.Float32bits(y))
	b.moveOffset(8)
	return b.offset
}

func (b *BlockBuffer) PutVector3Float32(x, y, z float32) int {
	binary.BigEndian.PutUint32(b.buf[b.offset:], math.Float32bits(x))
	binary.BigEndian.PutUint32(b.buf[b.offset+4:], math.Float32bits(y))
	binary.BigEndian.PutUint32(b.buf[b.offset+8:], math.Float32bits(z))
	b.moveOffset(12)
	return b.offset
}

func (b *Block) PutVector3Float32(x, y, z float32) int {
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset:], math.Float32bits(x))
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset+4:], math.Float32bits(y))
	binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset+8:], math.Float32bits(z))
	b.moveOffset(12)
	return b.offset
}

func (b *BlockBuffer) NewArray() {
	b.EndArray()
	b.currentArrayOffset = b.offset
	b.PutUint32(uint32(0))
	b.currentArraySize = 0
}

func (b *Block) NewArray() {
	// Does nothing, skip offset
	// Blocks cannot modify size of arrays
	b.moveOffset(4)
}

func (b *BlockBuffer) EndArray() {
	if b.currentArrayOffset > 0 {
		binary.BigEndian.PutUint32(b.buf[b.currentArrayOffset:], uint32(b.currentArraySize))
	}
	b.currentArrayOffset = 0
	b.currentArraySize = 0
}

func (b *Block) EndArray() {
	// Does nothing, skip offset
	// Blocks cannot modify size of arrays
}

func (b *BlockBuffer) PutMatrix(m compute.Matrix) int {
	b.NewArray()
	for i := range m {
		binary.BigEndian.PutUint32(b.buf[b.offset:], math.Float32bits(float32(m[i])))
		b.moveOffset(4)
	}
	b.EndArray()
	return b.offset
}

func (b *Block) PutMatrix(m compute.Matrix) int {
	b.NewArray()
	for i := range m {
		binary.BigEndian.PutUint32(b.buf.buf[b.startOffset+b.offset:], math.Float32bits(float32(m[i])))
		b.moveOffset(4)
	}
	b.EndArray()
	return b.offset
}

func (b *BlockBuffer) Reset() {
	b.offset = 0
	b.currentBlockOffset = 0
	b.currentBlockSize = 0
	b.currentArrayOffset = 0
	b.currentArraySize = 0
	for k := range b.blocks {
		delete(b.blocks, k)
	}
}

func (b *Block) Reset() {
	b.offset = 0
}

func (b *BlockBuffer) Copy(buf []byte) int {
	copy(buf, b.buf[0:b.offset])
	return b.offset
}

func (b *Block) Copy(buf []byte) int {
	copy(buf, b.buf.buf[b.startOffset-BlockPrefixBytes:b.startOffset+b.size])
	return b.size
}

func (b *Block) Free() {
	// Todo: instead shift all blocks after this block in the buffer
	b.Reset()
	for range b.size {
		b.PutUint8(0)
	}
	b.Reset()
}

func (b *BlockBuffer) Offset() int {
	return b.offset
}

func (b *BlockBuffer) Capacity() int {
	return b.capacity
}

// Declare a new block in buffer
func (b *BlockBuffer) NewBlock(kind uint8) int {
	b.EndArray()
	b.blocks[b.offset] = kind
	b.currentBlockOffset = b.offset
	b.currentBlockSize = 0
	b.PutUint8(kind)
	b.PutUint32(0) // placeholder for block size
	return b.offset
}

func (b *Block) NewBlock(kind uint8) int {
	b.offset = 0
	return b.offset
}

// End current block and returns a reference to it.
// The block can be updated but its size cannot change.
func (b *BlockBuffer) EndBlock() *Block {
	b.EndArray()

	blockOffset := b.currentBlockOffset
	blockSize := b.currentBlockSize
	b.currentBlockOffset = 0
	b.currentBlockSize = 0

	binary.BigEndian.PutUint32(b.buf[blockOffset+1:], uint32(blockSize))

	return &Block{
		kind:        b.buf[blockOffset],
		offset:      0,
		startOffset: blockOffset + BlockPrefixBytes,
		size:        blockSize,
		buf:         b,
	}
}

func (b *Block) EndBlock() *Block {
	b.offset = 0
	return b
}

func (b *BlockBuffer) AppendBlock(block *Block) {
	b.offset += block.Copy(b.buf[b.offset:])
}

func (b *Block) AppendBlock(block *Block) {
	// Do nothing
}

func (b *BlockBuffer) BlockCount() int {
	return len(b.blocks)
}

func (b *Block) Size() int {
	return b.size
}

func (b *Block) Uint8At(offset int) uint8 {
	return b.buf.buf[b.startOffset+offset]
}

func (b *Block) Uint8() uint8 {
	defer b.moveOffset(1)
	return b.buf.buf[b.startOffset+b.offset]
}

func (b *Block) Uint16At(offset int) uint16 {
	return binary.BigEndian.Uint16(b.buf.buf[b.startOffset+offset:])
}

func (b *Block) Uint16() uint16 {
	defer b.moveOffset(2)
	return binary.BigEndian.Uint16(b.buf.buf[b.startOffset+b.offset:])
}

func (b *Block) Uint32At(offset int) uint32 {
	return binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset:])
}

func (b *Block) Uint32() uint32 {
	defer b.moveOffset(4)
	return binary.BigEndian.Uint32(b.buf.buf[b.startOffset+b.offset:])
}

func (b *Block) Vector2Float32At(offset int) (float32, float32) {
	x := binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset:])
	y := binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset+4:])
	xf, yf := math.Float32frombits(x), math.Float32frombits(y)
	return xf, yf
}

func (b *Block) Vector3Float32At(offset int) (float32, float32, float32) {
	return math.Float32frombits(binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset:])),
		math.Float32frombits(binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset+4:])),
		math.Float32frombits(binary.BigEndian.Uint32(b.buf.buf[b.startOffset+offset+8:]))
}

func (b *Block) String() string {
	return fmt.Sprintf("block=%d offset=%d,%d size=%v", b.kind, b.startOffset-BlockPrefixBytes, b.offset, b.size)
}

func (b *BlockBuffer) moveOffset(amount int) {
	b.offset += amount
	b.currentBlockSize += amount
	if b.currentArrayOffset > 0 {
		b.currentArraySize += amount
	}
}

func (b *Block) moveOffset(amount int) {
	b.offset += amount
}
