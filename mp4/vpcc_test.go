package mp4

import (
	"testing"
)

func TestVpcC(t *testing.T) {
	vpcc := &VpcCBox{
		Profile:                 1,
		Level:                   10,
		BitDepth:                8,
		ChromaSubsampling:       1,
		VideoFullRangeFlag:      true,
		ColourPrimaries:         1,
		TransferCharacteristics: 1,
		MatrixCoefficients:      1,
		CodecInitializationData: []byte{0x01, 0x02, 0x03},
	}

	boxDiffAfterEncodeAndDecode(t, vpcc)
}
