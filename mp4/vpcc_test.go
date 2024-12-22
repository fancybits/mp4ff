package mp4

import (
	"testing"

	"github.com/Eyevinn/mp4ff/bits"
	"github.com/go-test/deep"
)

func TestVpcC(t *testing.T) {
	// Configure deep.Equal to treat nil slices as empty slices
	deep.NilSlicesAreEmpty = true

	testCases := []struct {
		name string
		box  *VpcCBox
	}{
		{
			name: "4k_vp9_profile0",
			box: &VpcCBox{
				Version:                     1,
				Flags:                       0,
				Profile:                     0,  // Profile 0 from vp09.00.50.08
				Level:                       50, // Level 5.0 from vp09.00.50.08
				BitDepth:                    8,  // 8-bit from vp09.00.50.08
				ChromaSubsampling:           0,  // 4:2:0 from test file
				VideoFullRangeFlag:          false,
				ColourPrimaries:             1, // BT.709
				TransferCharacteristics:     1, // BT.709
				MatrixCoefficients:          1, // BT.709
				CodecInitializationDataSize: 0,
				CodecInitializationData:     []byte{},
			},
		},
		{
			name: "high_quality_10bit_444",
			box: &VpcCBox{
				Version:                     1,
				Flags:                       0,
				Profile:                     2,  // Profile 2 (4:4:4 10/12-bit)
				Level:                       41, // Level 4.1
				BitDepth:                    10,
				ChromaSubsampling:           3, // 4:4:4
				VideoFullRangeFlag:          true,
				ColourPrimaries:             9,  // BT.2020
				TransferCharacteristics:     16, // PQ
				MatrixCoefficients:          9,  // BT.2020 non-constant luminance
				CodecInitializationDataSize: 0,
				CodecInitializationData:     []byte{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			boxDiffAfterEncodeAndDecode(t, tc.box)
		})
	}
}

func TestVpcCValidation(t *testing.T) {
	// Test invalid chromaSubsampling value
	invalidChroma := &VpcCBox{
		Version:           1,
		ChromaSubsampling: 4, // Invalid value > 3
	}
	_, err := encodeAndDecode(t, invalidChroma)
	if err == nil {
		t.Error("expected error for invalid ChromaSubsampling value")
	}

	// Test RGB matrix coefficients constraint
	invalidRGB := &VpcCBox{
		Version:            1,
		MatrixCoefficients: 0, // RGB
		ChromaSubsampling:  1, // Not 4:4:4
	}
	_, err = encodeAndDecode(t, invalidRGB)
	if err == nil {
		t.Error("expected error for RGB with non-444 chroma subsampling")
	}
}

// Helper function to encode and decode a box
func encodeBox(box Box) ([]byte, error) {
	sw := bits.NewFixedSliceWriter(int(box.Size()))
	err := box.EncodeSW(sw)
	if err != nil {
		return nil, err
	}
	return sw.Bytes(), nil
}

func decodeBox(data []byte) (Box, error) {
	sr := bits.NewFixedSliceReader(data)
	return DecodeBoxSR(0, sr)
}

func encodeAndDecode(t *testing.T, box Box) (Box, error) {
	bytes, err := encodeBox(box)
	if err != nil {
		return nil, err
	}

	decoded, err := decodeBox(bytes)
	if err != nil {
		return nil, err
	}

	// Convert nil slices to empty slices for comparison
	if src, ok := box.(*VpcCBox); ok {
		if src.CodecInitializationData == nil {
			src.CodecInitializationData = []byte{}
		}
	}
	if dst, ok := decoded.(*VpcCBox); ok {
		if dst.CodecInitializationData == nil {
			dst.CodecInitializationData = []byte{}
		}
	}

	return decoded, nil
}
