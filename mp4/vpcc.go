package mp4

import (
	"io"

	"github.com/Eyevinn/mp4ff/bits"
)

// VpcCBox - VP9 Codec Configuration Box
// See https://www.webmproject.org/vp9/mp4/
type VpcCBox struct {
	Profile                 uint8
	Level                   uint8
	BitDepth                uint8
	ChromaSubsampling       uint8
	VideoFullRangeFlag      bool
	ColourPrimaries         uint8
	TransferCharacteristics uint8
	MatrixCoefficients      uint8
	CodecInitializationData []byte
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

	// Skip first byte (reserved + version)
	sr.SkipBytes(1)

	// Read profile and level
	b.Profile = sr.ReadUint8()
	b.Level = sr.ReadUint8()

	// Read bit depth and color config
	colorConfig := sr.ReadUint8()
	b.BitDepth = (colorConfig >> 4) & 0x0f
	b.ChromaSubsampling = (colorConfig >> 1) & 0x07
	b.VideoFullRangeFlag = (colorConfig & 0x01) == 1

	// Read color description
	b.ColourPrimaries = sr.ReadUint8()
	b.TransferCharacteristics = sr.ReadUint8()
	b.MatrixCoefficients = sr.ReadUint8()

	// VP9 config box has a fixed size structure
	remaining := hdr.payloadLen() - 7 // 7 bytes read so far
	if remaining > 0 {
		b.CodecInitializationData = sr.ReadBytes(remaining)
	}

	return b, sr.AccError()
}

// Type - box type
func (b *VpcCBox) Type() string {
	return "vpcC"
}

// Size - calculated size of box
func (b *VpcCBox) Size() uint64 {
	// Size includes:
	// - boxHeaderSize (8 bytes)
	// - version (1 byte)
	// - profile (1 byte)
	// - level (1 byte)
	// - colorConfig (1 byte)
	// - colorPrimaries (1 byte)
	// - transferCharacteristics (1 byte)
	// - matrixCoefficients (1 byte)
	// - codecInitializationData (variable)
	return uint64(boxHeaderSize + 7 + len(b.CodecInitializationData))
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

	// Write version byte (always 1)
	sw.WriteUint8(1)

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

	// Write codec initialization data if any
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
