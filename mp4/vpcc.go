package mp4

import (
	"fmt"
	"io"

	"github.com/Eyevinn/mp4ff/bits"
)

// VpcCBox - VP Codec Configuration Box (vpcC)
// Extends FullBox per spec
// See https://www.webmproject.org/vp9/mp4/
type VpcCBox struct {
	Version                     uint8
	Flags                       uint32
	Profile                     uint8
	Level                       uint8
	BitDepth                    uint8
	ChromaSubsampling           uint8
	VideoFullRangeFlag          bool
	ColourPrimaries             uint8
	TransferCharacteristics     uint8
	MatrixCoefficients          uint8
	CodecInitializationDataSize uint16
	CodecInitializationData     []byte
}

// DecodeVpcC - box-specific decode
func DecodeVpcC(hdr BoxHeader, startPos uint64, r io.Reader) (Box, error) {
	data, err := readBoxBody(r, hdr)
	if err != nil {
		return nil, err
	}
	sr := bits.NewFixedSliceReader(data)
	return DecodeVpcCSR(hdr, startPos, sr)
}

// DecodeVpcCSR - box-specific decode from SliceReader
func DecodeVpcCSR(hdr BoxHeader, startPos uint64, sr bits.SliceReader) (Box, error) {
	b := &VpcCBox{}

	// Read FullBox version and flags
	b.Version = sr.ReadUint8()
	if b.Version != 1 {
		return nil, fmt.Errorf("vpcC version must be 1, got %d", b.Version)
	}
	b.Flags = sr.ReadUint24()

	// Read profile and level
	b.Profile = sr.ReadUint8()
	b.Level = sr.ReadUint8()

	// Read bit depth and color config
	colorConfig := sr.ReadUint8()
	b.BitDepth = (colorConfig >> 4) & 0x0f
	b.ChromaSubsampling = (colorConfig >> 1) & 0x07
	b.VideoFullRangeFlag = (colorConfig & 0x01) == 1

	// Validate chromaSubsampling
	if b.ChromaSubsampling > 3 {
		return nil, fmt.Errorf("invalid chromaSubsampling value: %d", b.ChromaSubsampling)
	}

	// Read color description
	b.ColourPrimaries = sr.ReadUint8()
	b.TransferCharacteristics = sr.ReadUint8()
	b.MatrixCoefficients = sr.ReadUint8()

	// Validate matrixCoefficients and chromaSubsampling combination
	if b.MatrixCoefficients == 0 && b.ChromaSubsampling != 3 {
		return nil, fmt.Errorf("when matrixCoefficients is 0 (RGB), chromaSubsampling must be 3 (4:4:4)")
	}

	// Read codec initialization data size and data
	b.CodecInitializationDataSize = sr.ReadUint16()
	if b.CodecInitializationDataSize > 0 {
		b.CodecInitializationData = sr.ReadBytes(int(b.CodecInitializationDataSize))
	}

	return b, sr.AccError()
}

// Type - box type
func (b *VpcCBox) Type() string {
	return "vpcC"
}

// Size - calculated size of box
func (b *VpcCBox) Size() uint64 {
	return uint64(boxHeaderSize + 12 + len(b.CodecInitializationData)) // 12 = version(1) + flags(3) + fixed fields(8)
}

// Encode - write box to w
func (b *VpcCBox) Encode(w io.Writer) error {
	sw := bits.NewFixedSliceWriter(int(b.Size()))
	err := b.EncodeSW(sw)
	if err != nil {
		return err
	}
	_, err = w.Write(sw.Bytes())
	return err
}

// EncodeSW - box-specific encode to SliceWriter
func (b *VpcCBox) EncodeSW(sw bits.SliceWriter) error {
	err := EncodeHeaderSW(b, sw)
	if err != nil {
		return err
	}

	// Write FullBox version (1) and flags (0)
	sw.WriteUint8(1)
	sw.WriteUint24(0)

	// Write profile and level
	sw.WriteUint8(b.Profile)
	sw.WriteUint8(b.Level)

	// Write bit depth and color config
	colorConfig := (b.BitDepth << 4) | (b.ChromaSubsampling << 1)
	if b.VideoFullRangeFlag {
		colorConfig |= 0x01
	}
	sw.WriteUint8(colorConfig)

	// Write color description
	sw.WriteUint8(b.ColourPrimaries)
	sw.WriteUint8(b.TransferCharacteristics)
	sw.WriteUint8(b.MatrixCoefficients)

	// Write codec initialization data size and data
	sw.WriteUint16(uint16(len(b.CodecInitializationData)))
	if len(b.CodecInitializationData) > 0 {
		sw.WriteBytes(b.CodecInitializationData)
	}

	return sw.AccError()
}

// Info - write box info
func (b *VpcCBox) Info(w io.Writer, specificBoxLevels, indent, indentStep string) error {
	bd := newInfoDumper(w, indent, b, -1, 0)
	bd.write(" - Profile: %d", b.Profile)
	bd.write(" - Level: %d", b.Level)
	bd.write(" - BitDepth: %d", b.BitDepth)
	bd.write(" - ChromaSubsampling: %d", b.ChromaSubsampling)
	bd.write(" - VideoFullRangeFlag: %t", b.VideoFullRangeFlag)
	bd.write(" - ColourPrimaries: %d", b.ColourPrimaries)
	bd.write(" - TransferCharacteristics: %d", b.TransferCharacteristics)
	bd.write(" - MatrixCoefficients: %d", b.MatrixCoefficients)
	return bd.err
}
