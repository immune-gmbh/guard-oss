// Copyright 2020 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package intelme

// MKHIHdr is the header for every MKHI command, see
// https://github.com/coreboot/coreboot/blob/b8b8ec832360ada5a313f10938bb6cfc310a11eb/src/soc/intel/common/block/include/intelblocks/cse.h#L64
type MkhiHdr [4]byte

// GroupID returns the 8-bit mkhi_hdr.group_id field
func (mh MkhiHdr) GroupID() uint8 {
	return mh[0]
}

// SetGroupID sets the GroupID field
func (mh *MkhiHdr) SetGroupID(v uint8) {
	mh[0] = v
}

// Command returns the 7-bit mkhi_hdr.command field
func (mh MkhiHdr) Command() uint8 {
	return mh[1] & 0x7f
}

// SetCommand sets the Command field. Only the first 7 bits will be used
func (mh *MkhiHdr) SetCommand(v uint8) {
	mh[1] = v & 0x7f
}

// IsResponse returns the 1-bit mkhi_hdr.is_resp field as boolean
func (mh MkhiHdr) IsResponse() bool {
	return mh[1]&0x80 == 0x80
}

// SetIsResponse sets the IsResponse field
func (mh *MkhiHdr) SetIsResponse(v bool) {
	if v {
		mh[1] |= 0x80
	} else {
		mh[1] &= ^byte(0x80)
	}
}

// RSVD returns the 8-bit mkhi_hdr.rsvd field
func (mh MkhiHdr) RSVD() uint8 {
	return mh[2]
}

// SetRSVD sets the RSVD field
func (mh *MkhiHdr) SetRSVD(v uint8) {
	mh[2] = v
}

// Result returns the 8-bit mkhi_hdr.result field
func (mh MkhiHdr) Result() uint8 {
	return mh[3]
}

// SetResult sets the Result field
func (mh *MkhiHdr) SetResult(v uint8) {
	mh[3] = v
}
