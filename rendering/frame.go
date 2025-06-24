package rendering

import (
	"encoding/binary"
)

type Frame []uint8

// Return RLE-blocks (3 byte length) from the diff between two frames, whether there was a diff,
// and number of blocks
func ComputeRLEDiffBlocks(old Frame, new Frame) ([]uint8, bool, int) {
	const prefixByteLength = 3

	var blocks []uint8 = make([]uint8, prefixByteLength)
	var blockSizeBytes []byte = make([]byte, 4)

	var currentBlockSize uint32 = 0
	var currentColor uint8 = new[0]

	var startPos, endPos int
	var deltaFound = false

	for i := range new {
		if new[i] != old[i] {
			if !deltaFound {
				startPos = i
				deltaFound = true
			}
			endPos = len(blocks) + 4
		}

		if !deltaFound {
			continue
		}

		if new[i] == currentColor {
			currentBlockSize++
		} else {
			binary.BigEndian.PutUint32(blockSizeBytes, currentBlockSize)
			blocks = append(blocks, blockSizeBytes[1:]...)
			blocks = append(blocks, currentColor)
			currentColor = new[i]
			currentBlockSize = 1
		}
	}

	// Add last block
	if currentBlockSize > 0 {
		binary.BigEndian.PutUint32(blockSizeBytes, currentBlockSize)
		blocks = append(blocks, blockSizeBytes[1:]...)
		blocks = append(blocks, currentColor)
		endPos += 4
	}

	// Store first position in first bytes
	if startPos > 0 {
		var startPosBytes []byte = make([]byte, 4)
		binary.BigEndian.PutUint32(startPosBytes, uint32(startPos))
		// Skip first byte (24bit)
		blocks[0] = startPosBytes[1]
		blocks[1] = startPosBytes[2]
		blocks[2] = startPosBytes[3]
	}

	// Trim last blocks
	if endPos < len(blocks) {
		blocks = blocks[0:endPos]
	}

	blocksLength := (len(blocks) - prefixByteLength) / 3

	return blocks, deltaFound, blocksLength
}

func ComputeRLEBlocks(data Frame) ([]uint8, int) {
	const prefixByteLength = 3

	var blocks []uint8 = make([]uint8, prefixByteLength)
	var blockSizeBytes []byte = make([]byte, 4)

	var currentBlockSize uint32 = 0
	var currentColor uint8 = data[0]

	for i := range data {
		if data[i] == currentColor {
			currentBlockSize++
		} else {
			binary.BigEndian.PutUint32(blockSizeBytes, currentBlockSize)
			blocks = append(blocks, blockSizeBytes[1:]...)
			blocks = append(blocks, currentColor)
			currentColor = data[i]
			currentBlockSize = 1
		}
	}

	// Add last block
	if currentBlockSize > 0 {
		binary.BigEndian.PutUint32(blockSizeBytes, currentBlockSize)
		blocks = append(blocks, blockSizeBytes[1:]...)
		blocks = append(blocks, currentColor)
	}

	blocksLength := (len(blocks) - prefixByteLength) / 3

	return blocks, blocksLength
}
