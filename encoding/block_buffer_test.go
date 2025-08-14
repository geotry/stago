package encoding

import (
	"fmt"
	"testing"

	"github.com/geotry/stago/compute"
)

func TestNewBlock(t *testing.T) {
	buf := NewBlockBuffer(255)

	// size=3
	buf.NewBlock(1)

	// size=2
	buf.PutUint8(1)
	buf.PutUint8(2)

	// size=2
	buf.PutUint16(30000)

	// size=4
	buf.PutUint32(1000000)

	buf.PutVector2Float32(.23545523, .230190)
	// size=12
	buf.PutVector3Float32(.98767667, .298399, .198923)

	// size=68
	buf.PutMatrix(compute.NewMatrix4().Out)

	block := buf.EndBlock()

	expectedSize := BlockPrefixBytes + 2 + 2 + 4 + 8 + 12 + 68

	if block.offset != 0 {
		t.Errorf("expected block offset to be 0")
	}
	if block.startOffset != BlockPrefixBytes {
		t.Errorf("expected block start offset to be %v", BlockPrefixBytes)
	}
	if int(block.kind) != 1 {
		t.Errorf("expected block kind to be %v, got %v", 1, block.kind)
	}
	if block.size != expectedSize {
		t.Errorf("expected block size to be %v, got %v", expectedSize, block.size)
	}
}

func TestArray(t *testing.T) {
	buf := NewBlockBuffer(255)

	buf.NewBlock(1)
	buf.PutUint16(1)
	buf.PutUint16(2)
	buf.PutUint16(3)
	buf.EndBlock()

	buf.NewBlock(1)
	buf.NewArray()
	buf.PutVector2Float32(1.0, 1.0)
	buf.PutVector2Float32(1.0, 1.0)
	buf.PutVector2Float32(1.0, 1.0)
	buf.EndArray()
	block := buf.EndBlock()

	arrayLength := block.Uint32At(0)
	if arrayLength != 24 {
		t.Errorf("expected block array size to be 24, got %v", arrayLength)
	}

	buf.NewBlock(1)
	buf.PutMatrix(compute.NewMatrix4().Out)
	block = buf.EndBlock()

	arrayLength = block.Uint32At(0)
	if arrayLength != 64 {
		t.Errorf("expected block array size to be 64, got %v", arrayLength)
	}
}

func TestReadBlock(t *testing.T) {
	buf := NewBlockBuffer(255)

	buf.NewBlock(1)
	buf.PutUint8(1)
	buf.PutUint8(2)
	buf.PutVector2Float32(2.0, 1.0)
	block := buf.EndBlock()

	if block.Uint8At(0) != 1 {
		t.Errorf("expected block at offset %v to be %v, got %v", 0, 1, block.Uint8At(0))
	}
	if block.Uint8At(1) != 2 {
		t.Errorf("expected block at offset %v to be %v, got %v", 1, 2, block.Uint8At(1))
	}
	x, y := block.Vector2Float32At(2)
	v := fmt.Sprintf("%.2f %.2f", x, y)
	if v != "2.00 1.00" {
		t.Errorf("expected block at offset 2 to be 2.00 1.00, got %v", v)
	}
}

func TestUpdateBlock(t *testing.T) {
	buf := NewBlockBuffer(255)

	buf.NewBlock(1)
	buf.PutUint8(1)
	buf.PutUint8(2)
	block := buf.EndBlock()

	block.PutUint8(3)
	block.PutUint8(4)

	if block.offset != 2 {
		t.Errorf("expected block offset to be 2")
	}

	tmp := make([]byte, block.Size())
	block.Copy(tmp)

	expected := []byte{1, 0, 0, 0, 7, 3, 4}
	for i := range tmp {
		if tmp[i] != expected[i] {
			t.Errorf("expected block at index %v to be %v, got %v (%v)", i, expected[i], tmp[i], tmp)
		}
	}
}
