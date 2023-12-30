// Intel Unique Platform ID
package intelme

import (
	"context"
)

const (
	// Commands
	UPIDGetFeatureState          = 0x01
	UPIDSetFeatureState          = 0x02
	UPIDGetFeatureOSControlState = 0x03
	UPIDSetFeatureOSControlState = 0x04
	UPIDGetPlatformID            = 0x05
	UPIDGetRefurbishCounter      = 0x06
	UPIDUpdateOEMPlatformID      = 0x07

	UPIDStatusSuccess                  = 0
	UPIDStatusFeatureNotSupported      = 1
	UPIDStatusInvalidInputParameter    = 2
	UPIDStatusInternalError            = 3
	UPIDStatusNotAllowedAfterEOP       = 4
	UPIDStatusNotAllowedAfterManufLock = 5
	UPIDStatusMaxCounterExceeded       = 6
	UPIDStatusPlatformIDDisabled       = 7
	UPIDStatusNotAllowedAfterCBD       = 9

	UPIDFileID     = 0x1001b900
	UPIDFileOffset = 0
	UPIDFileSize   = 1
)

type UPIDHeader struct {
	Feature   uint8
	Command   uint8
	ByteCount uint16
}

func EncodeHasUPID() []byte {
	return EncodeReadFileEx(UPIDFileID, UPIDFileOffset, UPIDFileSize)
}

func DecodeHasUPID(ctx context.Context, buf []byte) (bool, error) {
	data, err := DecodeReadFileEx(ctx, buf)
	if err != nil || len(data) != 1 {
		return false, ErrFormat
	}
	return data[0] == 1, nil
}
