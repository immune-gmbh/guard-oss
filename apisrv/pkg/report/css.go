package report

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/9elements/converged-security-suite/v2/pkg/registers"
	"github.com/9elements/go-linux-lowlevel-hw/pkg/hwapi"
	"github.com/digitalocean/go-smbios/smbios"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/api"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/evidence"
)

func hostbridgeConfigSpace(values *evidence.Values) (*api.Buffer, bool, error) {
	cfgSpc, err := values.PCIConfigSpace(0, 0, 0)
	if err != nil {
		return nil, false, err
	}

	if values.PCIVendorID(cfgSpc) != 0x8086 {
		return nil, false, fmt.Errorf("Hostbridge is not made by Intel")
	}
	deviceid := values.PCIDeviceID(cfgSpc)

	var found bool
	sandyAndNewer := false
	for _, id := range hwapi.HostbridgeIDsSandyCompatible {
		if id == deviceid {
			found = true
			sandyAndNewer = true
			break
		}
	}
	if !found {
		for _, id := range hwapi.HostbridgeIDsBroadwellDE {
			if id == deviceid {
				found = true
				break
			}
		}
	}

	if !found {
		return cfgSpc, sandyAndNewer, fmt.Errorf("Hostbridge is unsupported")
	} else {
		return cfgSpc, sandyAndNewer, nil
	}
}

type cssAPI struct {
	flash  api.ErrorBuffer
	values *evidence.Values
	txtAPI hwapi.LowLevelHardwareInterfaces
}

func (a cssAPI) VersionString() string {
	_, ebx, ecx, edx, err := a.values.CPUIDLeaf(0, 0)
	if err != nil {
		return ""
	}

	vendor := make([]byte, 12)
	binary.LittleEndian.PutUint32(vendor, ebx)
	binary.LittleEndian.PutUint32(vendor[4:], edx)
	binary.LittleEndian.PutUint32(vendor[8:], ecx)
	return string(vendor)
}

func (a cssAPI) HasSMX() bool {
	_, _, c, _, err := a.values.CPUIDLeaf(1, 0)
	return err == nil && c&(1<<6) != 0
}

func (a cssAPI) HasVMX() bool {
	_, _, c, _, err := a.values.CPUIDLeaf(1, 0)
	return err == nil && c&(1<<5) != 0
}

func (a cssAPI) HasMTRR() bool {
	_, _, _, d, err := a.values.CPUIDLeaf(1, 0)
	return err == nil && d&(1<<12) != 0
}

func (a cssAPI) ProcessorBrandName() string {
	buf := bytes.NewBuffer([]byte{})

	for i := uint32(0x80000002); i < 0x80000005; i += 1 {
		eax, ebx, ecx, edx, err := a.values.CPUIDLeaf(i, 0)
		if err != nil {
			return "unknown"
		}
		binary.Write(buf, binary.BigEndian, eax)
		binary.Write(buf, binary.BigEndian, ebx)
		binary.Write(buf, binary.BigEndian, ecx)
		binary.Write(buf, binary.BigEndian, edx)
	}

	return buf.String()
}

func (a cssAPI) CPUSignature() uint32 {
	eax, _, _, _, _ := a.values.CPUIDLeaf(1, 0)
	return eax
}

func (a cssAPI) CPUSignatureFull() (uint32, uint32, uint32, uint32) {
	eax, ebx, ecx, edx, _ := a.values.CPUIDLeaf(1, 0)
	return eax, ebx, ecx, edx
}

func (a cssAPI) CPULogCount() uint32 {
	_, b, _, _, _ := a.values.CPUIDLeaf(1, 0)
	return uint32((b >> 16) & 0xFF)
}

func (a cssAPI) IsReservedInE820(start uint64, end uint64) (bool, error) {
	return a.values.RangeReserved(start, end)
}

func (a cssAPI) LookupIOAddress(addr uint64, regs hwapi.VTdRegisters) ([]uint64, error) {
	return []uint64{}, fmt.Errorf("Not implemented")
}

func (a cssAPI) AddressRangesIsDMAProtected(first, end uint64) (bool, error) {
	var regs hwapi.VTdRegisters

	regSet := &a.values.VTdRegisterSet
	if regSet.Error != "" {
		return false, errors.New(string(regSet.Error))
	}

	rd := bytes.NewReader([]byte(regSet.Data))
	err := binary.Read(rd, binary.LittleEndian, &regs)
	if err != nil {
		return false, err
	}

	loDMAprotection := regs.Capabilities&(1<<5) != 0
	hiDMAprotection := regs.Capabilities&(1<<6) != 0
	enableDMAprotection := regs.Capabilities&1 != 0
	enable2DMAprotection := regs.Capabilities&(1<<31) != 0

	if enableDMAprotection && enable2DMAprotection && loDMAprotection && uint64(regs.ProtectedLowMemoryBase) <= first && uint64(regs.ProtectedLowMemoryLimit) >= end {
		return true, nil
	}

	if enableDMAprotection && enable2DMAprotection && hiDMAprotection && regs.ProtectedHighMemoryBase <= first && regs.ProtectedHighMemoryLimit >= end {
		return true, nil
	}

	/*
		for addr := first & 0xffffffffffff0000; addr < end; addr += 4096 {
			vas, err := t.LookupIOAddress(addr, regs)
			if err != nil {
				return false, err
			}

			if len(vas) < 0 {
				return false, nil
			}
		}
	*/

	return false, nil
}

func (a cssAPI) HasSMRR() (bool, error) {
	mtrrcap, err := a.values.MachineSpecReg(api.MSRMTRRCap)
	if err != nil {
		return false, fmt.Errorf("Cannot access MSR IA32_MTRRCAP: %s", err)
	}

	return (mtrrcap>>11)&1 != 0, nil
}

func (a cssAPI) GetSMRRInfo() (hwapi.SMRR, error) {
	var ret hwapi.SMRR

	smrrPhysbase, err := a.values.MachineSpecReg(api.MSRSMRRPhysBase)
	if err != nil {
		return ret, fmt.Errorf("Cannot access MSR IA32_SMRR_PHYSBASE: %s", err)
	}

	smrrPhysmask, err := a.values.MachineSpecReg(api.MSRSMRRPhysMask)
	if err != nil {
		return ret, fmt.Errorf("Cannot access MSR IA32_SMRR_PHYSMASK: %s", err)
	}

	ret.Active = (smrrPhysmask>>11)&1 != 0
	ret.PhysBase = (smrrPhysbase >> 12) & 0xfffff
	ret.PhysMask = (smrrPhysmask >> 12) & 0xfffff

	return ret, nil
}

func (a cssAPI) IA32FeatureControlIsLocked() (bool, error) {
	featCtrl, err := a.values.MachineSpecReg(api.MSRFeatureControl)
	if err != nil {
		return false, fmt.Errorf("Cannot access MSR IA32_FEATURE_CONTROL: %s", err)
	}

	return featCtrl&1 != 0, nil
}

func (a cssAPI) IA32PlatformID() (uint64, error) {
	pltID, err := a.values.MachineSpecReg(api.MSRPlatformID)
	if err != nil {
		return 0, fmt.Errorf("Cannot access MSR IA32_PLATFORM_ID: %s", err)
	}

	return pltID, nil
}

func (a cssAPI) AllowsVMXInSMX() (bool, error) {
	featCtrl, err := a.values.MachineSpecReg(api.MSRFeatureControl)
	if err != nil {
		return false, fmt.Errorf("Cannot access MSR IA32_FEATURE_CONTROL: %s", err)
	}

	var mask uint64 = (1 << 1) & (1 << 5) & (1 << 6)
	return (mask & featCtrl) == mask, nil
}

func (a cssAPI) TXTLeavesAreEnabled() (bool, error) {
	featCtrl, err := a.values.MachineSpecReg(api.MSRFeatureControl)
	if err != nil {
		return false, fmt.Errorf("Cannot access MSR IA32_FEATURE_CONTROL: %s", err)
	}

	txtBits := (featCtrl >> 8) & 0x1ff
	return (txtBits&0xff == 0xff) || (txtBits&0x100 == 0x100), nil
}

func (a cssAPI) IA32DebugInterfaceEnabledOrLocked() (*hwapi.IA32Debug, error) {
	var debugMSR hwapi.IA32Debug
	debugInterfaceCtrl, err := a.values.MachineSpecReg(api.MSRIA32DebugInterface)
	if err != nil {
		return nil, fmt.Errorf("Cannot access MSR IA32_DEBUG_INTERFACE: %s", err)
	}

	debugMSR.Enabled = (debugInterfaceCtrl>>0)&1 != 0
	debugMSR.Locked = (debugInterfaceCtrl>>30)&1 != 0
	debugMSR.PCHStrap = (debugInterfaceCtrl>>31)&1 != 0
	return &debugMSR, nil
}

func (a cssAPI) UsableMemoryBelow4G() (uint64, error) {
	panic("not implemented")
}

func (a cssAPI) UsableMemoryAbove4G() (uint64, error) {
	panic("not implemented")
}

func (a cssAPI) ReadMSRAllCores(int64) (uint64, error) {
	panic("not implemented")
}

func (a cssAPI) ReadMSR(int64) (uint64, error) {
	panic("not implemented")
}

func (a cssAPI) PCIEnumerateVisibleDevices(func(hwapi.PCIDevice) bool) error {
	panic("not implemented")
}

func (a cssAPI) PCIReadConfigSpace(bus int, device int, devFn int, off int, buf interface{}) error {
	cfgsp, err := a.values.PCIConfigSpace(uint16(bus), uint16(device), uint8(devFn))
	if err != nil {
		return err
	}

	rd := bytes.NewReader([]byte(*cfgsp))
	if _, err := rd.Seek(int64(off), io.SeekStart); err != nil {
		return err
	}
	return binary.Read(rd, binary.LittleEndian, buf)
}

func (a cssAPI) PCIReadConfig8(dev hwapi.PCIDevice, off int) (uint8, error) {
	var reg8 uint8
	err := a.PCIReadConfigSpace(dev.Bus, dev.Device, dev.Function, off, &reg8)
	if err != nil {
		return 0, err
	}

	return reg8, nil
}

func (a cssAPI) PCIReadConfig16(dev hwapi.PCIDevice, off int) (uint16, error) {
	var reg16 uint16
	err := a.PCIReadConfigSpace(dev.Bus, dev.Device, dev.Function, off, &reg16)
	if err != nil {
		return 0, err
	}

	return reg16, nil
}

func (a cssAPI) PCIReadConfig32(dev hwapi.PCIDevice, off int) (uint32, error) {
	var reg32 uint32
	err := a.PCIReadConfigSpace(dev.Bus, dev.Device, dev.Function, off, &reg32)
	if err != nil {
		return 0, err
	}

	return reg32, nil
}

func (a cssAPI) PCIReadVendorID(dev hwapi.PCIDevice) (uint16, error) {
	id, err := a.PCIReadConfig16(dev, 0)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a cssAPI) PCIReadDeviceID(dev hwapi.PCIDevice) (uint16, error) {
	id, err := a.PCIReadConfig16(dev, 2)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a cssAPI) PCIWriteConfig8(dev hwapi.PCIDevice, off int, data uint8) error {
	panic("not implemented")
}

func (a cssAPI) PCIWriteConfig16(dev hwapi.PCIDevice, off int, data uint16) error {
	panic("not implemented")
}

func (a cssAPI) PCIWriteConfig32(dev hwapi.PCIDevice, off int, data uint32) error {
	panic("not implemented")
}

func (a cssAPI) ReadHostBridgeTseg() (uint32, uint32, error) {
	var tsegBaseOff int
	var tsegLimitOff int

	var tsegbase uint32
	var tseglimit uint32

	cfgSpc, sandyAndNewer, err := hostbridgeConfigSpace(a.values)
	if err != nil {
		return 0, 0, err
	}

	if sandyAndNewer {
		tsegBaseOff = hwapi.TsegPCIRegSandyAndNewer
		tsegLimitOff = hwapi.TsegPCIRegSandyAndNewer + 4
	} else {
		tsegBaseOff = hwapi.TSEGPCIBroadwellde
		tsegLimitOff = hwapi.TSEGPCIBroadwellde + 4
	}

	tsegbase = binary.LittleEndian.Uint32([]byte(*cfgSpc)[tsegBaseOff:])
	if err != nil {
		return 0, 0, err
	}

	tseglimit = binary.LittleEndian.Uint32([]byte(*cfgSpc)[tsegLimitOff:])
	if err != nil {
		return 0, 0, err
	}

	if !sandyAndNewer {
		// On BroadwellDe TSEG limit lower 19bits are don't care, thus add 1 MiB.
		tseglimit += 1024 * 1024
	}

	return tsegbase, tseglimit, nil
}

func (a cssAPI) ReadHostBridgeDPR() (hwapi.DMAProtectedRange, error) {

	cfgSpc, sandyAndNewer, err := hostbridgeConfigSpace(a.values)
	if err != nil {
		return hwapi.DMAProtectedRange{}, err
	}

	var dprOff int
	if sandyAndNewer {
		dprOff = hwapi.DPRPCIRegSandyAndNewer
	} else {
		dprOff = hwapi.DPRPciRegBroadwellDE
	}

	u32 := binary.LittleEndian.Uint32([]byte(*cfgSpc)[dprOff:])
	if err != nil {
		return hwapi.DMAProtectedRange{}, err
	}

	return hwapi.DMAProtectedRange{
		Lock: u32&1 != 0,
		Size: uint8((u32 >> 4) & 0xff),    // 11:4
		Top:  uint16((u32 >> 20) & 0xfff), // 31:20
	}, nil
}

func (a cssAPI) ReadPhys(addr int64, data hwapi.UintN) error {
	buf := make([]byte, data.Size())
	if err := a.ReadPhysBuf(addr, buf); err != nil {
		return err
	}

	rd := bytes.NewReader((buf))
	return binary.Read(rd, binary.LittleEndian, data)
}

func (a cssAPI) ReadPhysBuf(addr int64, buf []byte) error {
	flash := &a.flash
	txtPublic := &a.values.TXTPublicSpace
	end := addr + int64(len(buf))

	var flashEnd int64 = 0x1_0000_0000
	var flashStart int64 = flashEnd - int64(len(flash.Data))
	var txtPublicStart int64 = 0xfed3_0000
	var txtPublicEnd int64 = txtPublicStart + int64(len(txtPublic.Data))
	// XXX: txtSinitStart := 0
	// XXX: txtSinitEnd := 0
	// XXX: txtHeapStart := 0
	// XXX: txtHeapEnd := 0

	var rd io.Reader
	if addr >= flashStart && end <= flashEnd && flash.Error == "" {
		rd = bytes.NewReader(flash.Data[addr-flashStart:])
	} else if addr >= txtPublicStart && end <= txtPublicEnd && txtPublic.Error == "" {
		rd = bytes.NewReader(txtPublic.Data[addr-txtPublicStart:])
	} else {
		return notFoundErr
	}

	l, err := rd.Read(buf)
	if err != nil {
		return err
	}
	if l != len(buf) {
		return notFoundErr
	}
	return nil
}

// Unused in CSS
func (a cssAPI) WritePhys(addr int64, data hwapi.UintN) error {
	return fmt.Errorf("Not implemented")
}

func (t cssAPI) GetMSRRegisters() (registers.Registers, error) {
	return registers.ReadMSRRegisters(t)
}

func (t cssAPI) Read(msr int64) (uint64, error) {
	return t.values.MachineSpecReg(uint32(msr))
}

func (a cssAPI) NewTPM() (*hwapi.TPM, error) {
	return nil, nil
}

func (a cssAPI) NVLocked(tpmCon *hwapi.TPM) (bool, error) {
	return false, nil
}

func (a cssAPI) ReadNVPublic(tpmCon *hwapi.TPM, index uint32) ([]byte, error) {
	if pub, err := a.values.NVPublic(index); err == nil {
		return tpmutil.Pack(tpm2.NVPublic(*pub))
	} else {
		return []byte{}, err
	}
}

func (a cssAPI) NVReadValue(tpmCon *hwapi.TPM, index uint32, password string, size, offhandle uint32) ([]byte, error) {
	if password != "" {
		return []byte{}, errors.New("invalid password")
	}

	if buf, err := a.values.NVValue(index); err == nil {
		return []byte(*buf)[offhandle:(offhandle + size)], nil
	} else {
		return []byte{}, err
	}
}

func (a cssAPI) ReadPCR(tpmCon *hwapi.TPM, pcr uint32) ([]byte, error) {
	if bank, ok := a.values.PCR["11"]; ok {
		if pcr, ok := bank[fmt.Sprintf("%d", pcr)]; ok {
			return hex.DecodeString(pcr)
		}
	}
	return []byte{}, notFoundErr
}

func (a cssAPI) IterateOverSMBIOSTablesType17(func(*hwapi.SMBIOSType17) bool) (bool, error) {
	panic("not implemented")
}

func (a cssAPI) IterateOverSMBIOSTablesType0(func(*hwapi.SMBIOSType0) bool) (bool, error) {
	panic("not implemented")
}

func (a cssAPI) IterateOverSMBIOSTables(uint8, func(*smbios.Structure) bool) (bool, error) {
	panic("not implemented")
}

func (a cssAPI) GetACPITableSysFS(arg string) ([]byte, error) {
	if buf, err := a.values.ACPITable(arg); err == nil {
		return []byte(*buf), nil
	} else {
		return []byte{}, err
	}
}

func (a cssAPI) GetACPITableDevMem(arg string) ([]byte, error) {
	if buf, err := a.values.ACPITable(arg); err == nil {
		return []byte(*buf), nil
	} else {
		return []byte{}, err
	}
}

func (a cssAPI) GetACPITable(arg string) ([]byte, error) {
	if buf, err := a.values.ACPITable(arg); err == nil {
		return []byte(*buf), nil
	} else {
		return []byte{}, err
	}
}

func NewCSSAPI(values *evidence.Values) hwapi.LowLevelHardwareInterfaces {
	return cssAPI{
		values: values,
		txtAPI: hwapi.GetAPI(),
	}
}
