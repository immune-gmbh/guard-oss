// Copyright 2020 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package intelme

import "github.com/google/uuid"

const (
	PCIDevice   = 0x16
	PCIFunction = 0
)

var (
	// MEI Guids use the MS/Intel encoding:
	// encoding [0,1,2,3,4,5,6,7,8,9,0xa,0xb,0xc,0xd,0xe,0xf]
	// becomes 03020100-0504-0706-0809-0a0b0c0d0e0f
	// i.e. (32 bit le)-(16 bit le)-(16 bit le)-(16 bit be)-(6 bytes, be obv.)

	// HECIGuids maps the known HECI GUIDs to their values. The MEI interface wants
	// little-endian. See all the GUIDs at
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	CSME_MKHIGuid = uuid.MustParse("8e6a6715-9abc-4043-88ef-9e39c6f63e0f")
	SPS_MKHIGuid  = uuid.MustParse("55213584-9a29-4916-badf-0fb7ed682aeb")
	FWUpdateGuid  = uuid.MustParse("309DCDE8-CCB1-4062-8F78-600115A34327")

	ICCGuid1 = uuid.MustParse("42b3ce2f-bd9f-485a-96ae-26406230b1ff")
	ICCGuid2 = uuid.MustParse("bf3cb4da-4045-4f9b-838d-8cbcfb21a107") // debugfs, Ivy Bridge

	MCHIGuid1 = uuid.MustParse("6861EC7B-D07A-4673-856C-7F22B4D55769")

	// Intel AMT Implementation and reference guide
	AMTGuid1 = uuid.MustParse("e2d1ff34-3458-49a9-88da-8e6915ce9be5")
	AMTGuid2 = uuid.MustParse("bb875e12-cb58-4d14-ae93-8566183c66c7") // debugfs
	AMTGuid3 = uuid.MustParse("12f80028-b4b7-4b2d-aca8-46e0ff65814c") // https://github.com/mjg59/mei-amt-check
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	WatchdogGuid = uuid.MustParse("05b79a6f-4628-4d7f-899d-a91514cb32ab")

	// Trusted Device Setup
	TDSGuid = uuid.MustParse("2FD631c9-1309-4fe6-bec6-85732778ab01")

	Unknown1Guid = uuid.MustParse("01e88543-8050-4380-9d6f-4f9cec704917") // debugfs
	Unknown2Guid = uuid.MustParse("cea154ea-8ff5-4f94-9290-0bb7355a34db") // debugfs
	Unknown3Guid = uuid.MustParse("fa8f55e8-ab22-42dd-b916-7dce39002574") // debugfs, Ivy Brigde, client 5
	MCHIGuid2    = uuid.MustParse("d2ea63bc-5f04-4997-9454-8cadf4e3ef8a") // debugfs, Ivy Bridge

	// HOTHAM
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	HOTHAMGuid = uuid.MustParse("082ee5a7-7c25-470a-9643-0c06f0466ea1")

	// LME
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	LMEGuid = uuid.MustParse("6733a4db-0476-4e7b-b3af-bcfc29bee7a7")

	// Platform Service Record
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	PSRGuid = uuid.MustParse("ED6703FA-D312-4E8B-9DDD-2155BB2DEE65")

	// Unique Platform ID
	// https://github.com/intel/lms/blob/master/MEIClient/Include/HECI_if.h
	UPIDGuid = uuid.MustParse("92136C79-5FEA-4CFD-980e-23BE07FA5E9F")
)
