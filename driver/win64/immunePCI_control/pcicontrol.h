#ifndef	__PCIFILTER_H
#define	__PCIFILTER_H

#define	FILE_DEVICE_PHYMEM	0x8000

#define	IOCTL_PHYMEM_GETPCI	\
	CTL_CODE(FILE_DEVICE_PHYMEM, 0x804,\
			 METHOD_OUT_DIRECT, FILE_READ_DATA | FILE_WRITE_DATA)

typedef struct tagPHYMEM_PCI
{
	ULONG dwBusNum;		//bus number: 0-255
	ULONG dwDevNum;		//device number: 0-31
	ULONG dwFuncNum;	//function number: 0-7
	ULONG dwRegOff;		//register offset: 0-255
	ULONG dwBytes;		//bytes to read or write
} PHYMEM_PCI, * PPHYMEM_PCI;

#endif	//__PCIFILTERZ_H#pragma once
