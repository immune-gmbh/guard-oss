// Copyright 2020 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package intelme

// FWupHdr is the header for every FWUpdate command, see
// https://github.com/intel/dynamic-application-loader-host-interface/blob/master/common/FWUpdate/FwuCommon.h
type FWupHdr [4]byte

// GroupID returns the 8-bit mkhi_hdr.group_id field
func (fh FWupHdr) MessageID() uint8 {
	return fh[0]
}

// SetGroupID sets the GroupID field
func (fh *FWupHdr) SetMessageID(v uint8) {
	fh[0] = v
}

// Result returns the 8-bit mkhi_hdr.result field
func (fh FWupHdr) Result() uint8 {
	return fh[3]
}

// SetResult sets the Result field
func (fh *FWupHdr) SetResult(v uint8) {
	fh[3] = v
}
