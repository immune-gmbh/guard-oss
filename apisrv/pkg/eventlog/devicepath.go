package eventlog

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
)

func (dp efiPCIDevicePath) String() string {
	return fmt.Sprintf("Pci(0x%x,0x%x)", dp.Device, dp.Function)
}

func (dp efiMMIODevicePath) String() string {
	return fmt.Sprintf("MemoryMapped(0x%x,0x%x,0x%x)", dp.MemoryType, dp.StartAddress, dp.EndAddress)
}

func (dp efiHardDriveDevicePath) String() string {
	switch dp.SignatureType {
	case mbr:
		return fmt.Sprintf("HD(%d,MBR,0x%08x,0x%x,0x%x)", dp.Partition, dp.PartitionSignature, dp.PartitionStart, dp.PartitionSize)
	case guid:
		guid := EFIGUID{}
		buf := bytes.NewReader(dp.PartitionSignature[:])
		binary.Read(buf, binary.LittleEndian, &guid)
		guidString := guid.String()
		return fmt.Sprintf("HD(%d,GPT,%s,0x%x,0x%x)", dp.Partition, guidString, dp.PartitionStart, dp.PartitionSize)
	default:
		return fmt.Sprintf("HD(%d,%d,0,0x%x,0x%x)", dp.Partition, dp.SignatureType, dp.PartitionStart, dp.PartitionSize)
	}
}

func (dp efiMacMessagingDevicePath) String() string {
	hwAddressSize := len(dp.MAC)
	if dp.IfType == 0x01 || dp.IfType == 0x00 {
		hwAddressSize = 6
	}

	output := "MAC("
	for i := 0; i < hwAddressSize; i++ {
		output += fmt.Sprintf("%02x", dp.MAC[i])
	}
	output += fmt.Sprintf(",0x%x)", dp.IfType)
	return output
}

func (dp efiIpv4MessagingDevicePath) String() string {
	output := "IPv4("
	output += fmt.Sprintf("%d.%d.%d.%d:%d,", dp.RemoteAddress[0], dp.RemoteAddress[1], dp.RemoteAddress[2], dp.RemoteAddress[3], dp.RemotePort)
	switch dp.Protocol {
	case 6:
		output += "TCP,"
	case 17:
		output += "UDP,"
	default:
		output += fmt.Sprintf("0x%x,", dp.Protocol)
	}
	if dp.StaticIP == 0 {
		output += "DHCP,"
	} else {
		output += "Static,"
	}
	output += fmt.Sprintf("%d.%d.%d.%d:%d,", dp.LocalAddress[0], dp.LocalAddress[1], dp.LocalAddress[2], dp.LocalAddress[3], dp.LocalPort)
	output += fmt.Sprintf("%d.%d.%d.%d,", dp.GatewayAddress[0], dp.GatewayAddress[1], dp.GatewayAddress[2], dp.GatewayAddress[3])
	output += fmt.Sprintf("%d.%d.%d.%d)", dp.SubnetMask[0], dp.SubnetMask[1], dp.SubnetMask[2], dp.SubnetMask[3])
	return output
}

func (dp efiIpv6MessagingDevicePath) String() string {
	output := "IPv6("
	output += fmt.Sprintf("%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x:%d,", dp.RemoteAddress[:2], dp.RemoteAddress[2:4], dp.RemoteAddress[4:6], dp.RemoteAddress[6:8], dp.RemoteAddress[8:10], dp.RemoteAddress[10:12], dp.RemoteAddress[12:14], dp.RemoteAddress[14:16], dp.RemotePort)
	switch dp.Protocol {
	case 6:
		output += "TCP,"
	case 17:
		output += "UDP,"
	default:
		output += fmt.Sprintf("0x%x,", dp.Protocol)
	}
	switch dp.AddressOrigin {
	case 0:
		output += "Static,"
	case 1:
		output += "StatelessAutoConfigure,"
	default:
		output += "StatefulAutoConfigure,"
	}
	output += fmt.Sprintf("%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x:%d,", dp.LocalAddress[0:2], dp.LocalAddress[2:4], dp.LocalAddress[4:6], dp.LocalAddress[6:8], dp.LocalAddress[8:10], dp.LocalAddress[10:12], dp.LocalAddress[12:14], dp.LocalAddress[14:16], dp.LocalPort)
	output += fmt.Sprintf("0x%x,", dp.PrefixLength)
	output += fmt.Sprintf("%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x)", dp.GatewayIP[0:2], dp.GatewayIP[2:4], dp.GatewayIP[4:6], dp.GatewayIP[6:8], dp.GatewayIP[8:10], dp.GatewayIP[10:12], dp.GatewayIP[12:14], dp.GatewayIP[14:16])
	return output
}

func (dp efiUsbMessagingDevicePath) String() string {
	return fmt.Sprintf("USB(0x%x,0x%x)", dp.ParentPort, dp.Interface)
}

func (dp efiSataMessagingDevicePath) String() string {
	return fmt.Sprintf("Sata(0x%x,0x%x,0x%x)", dp.HBA, dp.PortMultiplier, dp.LUN)
}

func (dp efiNvmMessagingDevicePath) String() string {
	output := fmt.Sprintf("NVMe(0x%x,", dp.Namespace)
	for _, id := range dp.EUI {
		output += fmt.Sprintf("%02x-", id)
	}
	output = strings.TrimSuffix(output, "-")
	output += ")"
	return output
}

func (dp efiACPIDevicePath) String() string {
	if (dp.HID & 0xffff) != 0x41d0 {
		return fmt.Sprintf("Acpi(0x%08x,0x%x)", dp.HID, dp.UID)
	}
	switch dp.HID >> 16 {
	case 0x0a03:
		return fmt.Sprintf("PciRoot(0x%x)", dp.UID)
	case 0x0a08:
		return fmt.Sprintf("PcieRoot(0x%x)", dp.UID)
	case 0x0604:
		return fmt.Sprintf("Floppy(0x%x)", dp.UID)
	case 0x0301:
		return fmt.Sprintf("Keyboard(0x%x)", dp.UID)
	case 0x0501:
		return fmt.Sprintf("Serial(0x%x)", dp.UID)
	case 0x0401:
		return fmt.Sprintf("ParallelPort(0x%x)", dp.UID)
	default:
		return fmt.Sprintf("Acpi(PNP%04x,0x%x)", dp.HID>>16, dp.UID)
	}
}

func (dp efiExpandedACPIDevicePath) String() string {
	if dp.Fixed.HID>>16 == 0x0a08 || dp.Fixed.CID>>16 == 0x0a08 {
		if dp.Fixed.UID == 0 {
			return fmt.Sprintf("PcieRoot(%s)", dp.UIDStr)
		}
		return fmt.Sprintf("PcieRoot(0x%x)", dp.Fixed.UID)
	}

	HID := fmt.Sprintf("%c%c%c%04x", ((dp.Fixed.HID>>10)&0x1f)+0x40,
		((dp.Fixed.HID>>5)&0x1f)+0x40,
		(dp.Fixed.HID&0x1f)+0x40,
		dp.Fixed.HID>>16)
	CID := fmt.Sprintf("%c%c%c%04x", ((dp.Fixed.CID>>10)&0x1f)+0x40,
		((dp.Fixed.CID>>5)&0x1f)+0x40,
		(dp.Fixed.CID&0x1f)+0x40,
		dp.Fixed.CID>>16)

	if dp.HIDStr == "" && dp.CIDStr == "" && dp.UIDStr == "" {
		if dp.Fixed.CID == 0 {
			return fmt.Sprintf("AcpiExp(%s,0,%s)", HID, dp.UIDStr)
		}
		return fmt.Sprintf("AcpiExp(%s,%s,%s)", HID, CID, dp.UIDStr)
	}

	return fmt.Sprintf("AcpiExp(%s, %s, 0x%x, %s, %s, %s)", HID, CID, dp.Fixed.UID, dp.HIDStr, dp.CIDStr, dp.UIDStr)
}

func (dp efiAdrACPIDevicePath) String() string {
	output := "AcpiAdr("
	for _, adr := range dp.ADRs {
		output += fmt.Sprintf("0x%x,", adr)
	}
	output = strings.TrimSuffix(output, ",")
	output += ")"
	return output
}

func (dp efiPiwgFileDevicePath) String() string {
	return fmt.Sprintf("FvFile(%s)", dp.GUID)
}

func (dp efiPiwgVolumeDevicePath) String() string {
	return fmt.Sprintf("Fv(%s)", dp.GUID)
}

func (dp efiOffsetDevicePath) String() string {
	return fmt.Sprintf("Offset(0x%x, 0x%x)", dp.StartOffset, dp.EndOffset)
}

func (dp efiBBSDevicePath) String() string {
	output := "BBS("
	description := strings.TrimSuffix(string(dp.Description), "\u0000")
	switch dp.Fixed.DeviceType {
	case 0x01:
		output += fmt.Sprintf("Floppy,%s", description)
	case 0x02:
		output += fmt.Sprintf("HD,%s", description)
	case 0x03:
		output += fmt.Sprintf("CDROM,%s", description)
	case 0x04:
		output += fmt.Sprintf("PCMCIA,%s", description)
	case 0x05:
		output += fmt.Sprintf("USB,%s", description)
	case 0x06:
		output += fmt.Sprintf("Network,%s", description)
	default:
		output += fmt.Sprintf("0x%x,%s", dp.Fixed.DeviceType, description)
	}
	output += fmt.Sprintf(",0x%x)", dp.Fixed.Status)
	return output
}

func dumpEfiDevicePath(buf io.Reader, dp efiDevicePathHeader, prefix string) string {
	data := make([]byte, dp.Length-4)
	binary.Read(buf, binary.LittleEndian, &data)
	if prefix == "" {
		return fmt.Sprintf("Path(%d,%d,%02x)", dp.Type, dp.SubType, data[:])
	}
	return fmt.Sprintf("%s(%d,%02x)", prefix, dp.SubType, data[:])
}

// efiDevicePath translates an EFI Device Path into the canonical string representation
func efiDevicePath(b []byte) (string, error) {
	buf := bytes.NewReader(b)
	offset := 0
	dp := efiDevicePathHeader{}
	output := ""

	for offset < len(b) {
		buf.Seek(int64(offset), io.SeekStart)
		binary.Read(buf, binary.LittleEndian, &dp)
		offset += int(dp.Length)
		if offset == 0 || offset > len(b) {
			return "", fmt.Errorf("malformed device path")
		}
		switch dp.Type {
		case hardwareDevicePath:
			switch hwDPType(dp.SubType) {
			case pciHwDevicePath:
				path := efiPCIDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case mmioHwDevicePath:
				path := efiMMIODevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			default:
				output += dumpEfiDevicePath(buf, dp, "HardwarePath")
			}
		case acpiDevicePath:
			switch acpiDPType(dp.SubType) {
			case normalACPIDevicePath:
				path := efiACPIDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case expandedACPIDevicePath:
				path := efiExpandedACPIDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path.Fixed)
				data := make([]byte, dp.Length-16)
				buf.Read(data)
				path.HIDStr = string(data)
				path.UIDStr = string(data[len(path.HIDStr)+1:])
				path.CIDStr = string(data[len(path.HIDStr)+len(path.UIDStr)+2:])
				output += path.String()
			case adrACPIDevicePath:
				path := efiAdrACPIDevicePath{}
				path.ADRs = make([]uint32, (dp.Length-4)/4)
				binary.Read(buf, binary.LittleEndian, &path.ADRs)
				output += path.String()
			default:
				output += dumpEfiDevicePath(buf, dp, "AcpiPath")
			}
		case messagingDevicePath:
			switch messagingDPType(dp.SubType) {
			case usbMessagingDevicePath:
				path := efiUsbMessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case vendorMessagingDevicePath:
				path := efiVendorMessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path.GUID)
				path.Data = make([]byte, dp.Length-20)
				buf.Read(path.Data)
				GUIDString := path.GUID.String()
				output += fmt.Sprintf("VenMsg(%s)", GUIDString)
			case macMessagingDevicePath:
				path := efiMacMessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case ipv4MessagingDevicePath:
				path := efiIpv4MessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case ipv6MessagingDevicePath:
				path := efiIpv6MessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case sataMessagingDevicePath:
				path := efiSataMessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case nvmMessagingDevicePath:
				path := efiNvmMessagingDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			default:
				output += dumpEfiDevicePath(buf, dp, "Msg")
			}
		case mediaDevicePath:
			switch mediaDPType(dp.SubType) {
			case hardDriveDevicePath:
				path := efiHardDriveDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case filePathDevicePath:
				path := make([]uint16, (dp.Length-4)/2)
				binary.Read(buf, binary.LittleEndian, &path)
				filename := strings.TrimSuffix(string(utf16.Decode(path)), "\u0000")
				output += filename
			case piwgFileDevicePath:
				path := efiPiwgFileDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case piwgVolumeDevicePath:
				path := efiPiwgVolumeDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			case offsetDevicePath:
				path := efiOffsetDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path)
				output += path.String()
			default:
				output += dumpEfiDevicePath(buf, dp, "MediaPath")
			}
		case bbsDevicePath:
			switch bbsDPType(dp.SubType) {
			case bbs101DevicePath:
				path := efiBBSDevicePath{}
				binary.Read(buf, binary.LittleEndian, &path.Fixed)
				path.Description = make([]byte, dp.Length-8)
				buf.Read(path.Description)
				output += path.String()
			default:
				output += dumpEfiDevicePath(buf, dp, "BbsPath")
			}
		case endDevicePath:
			switch endDPType(dp.SubType) {
			case endThisDevicePath:
				output += ","
			case endEntireDevicePath:
				output = strings.TrimSuffix(output, "/")
				output += ","
				continue
			default:
				output += fmt.Sprintf("Unknown end subtype %d",
					dp.SubType)
			}
		default:
			output += dumpEfiDevicePath(buf, dp, "")
		}
		output += "/"
	}

	output = strings.TrimSuffix(output, "/")
	return output, nil
}
