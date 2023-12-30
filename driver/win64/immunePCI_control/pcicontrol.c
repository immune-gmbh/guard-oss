#define POOL_NX_OPTIN 1
#include <ntddk.h>
#include <wdf.h>
#include <initguid.h>
#include <wdmguid.h>
#include <wdmsec.h>
#include "pcicontrol.h"

DRIVER_INITIALIZE DriverEntry;
EVT_WDF_DRIVER_UNLOAD immunePCICtrlEvtDriverUnload;
EVT_WDF_DEVICE_CONTEXT_CLEANUP immunePCICtrlEvtDriverContextCleanup;
EVT_WDF_DEVICE_SHUTDOWN_NOTIFICATION immunePCICtrlMachineShutdown;
EVT_WDF_IO_QUEUE_IO_DEVICE_CONTROL immunePCICtrlEvtIoDeviceControl;
EVT_WDF_DEVICE_FILE_CREATE immunePCICtrlEvtDeviceFileCreate;
EVT_WDF_FILE_CLOSE immunePCICtrlEvtFileClose;


// Don't use EVT_WDF_DRIVER_DEVICE_ADD for immunePCICtrlDeviceAdd even though 
// the signature is same because this is not an event called by the framework.
NTSTATUS immunePCICtrlDeviceAdd(IN WDFDRIVER Driver, IN PWDFDEVICE_INIT DeviceInit);
NTSTATUS GetPCIBusInterfaceStandard(IN WDFDEVICE Device, OUT PPCI_BUS_INTERFACE_STANDARD BusInterfaceStandard);
static NTSTATUS ReadPciConfig(IN WDFDEVICE Device, OUT PPHYMEM_PCI pPci, OUT PVOID OutputBuffer);


#ifdef ALLOC_PRAGMA
#pragma alloc_text( INIT, DriverEntry )
#pragma alloc_text( PAGE, immunePCICtrlDeviceAdd)
#pragma alloc_text( PAGE, immunePCICtrlEvtDriverContextCleanup)
#pragma alloc_text( PAGE, immunePCICtrlEvtDriverUnload)
#pragma alloc_text( PAGE, immunePCICtrlEvtDeviceFileCreate)
#pragma alloc_text( PAGE, immunePCICtrlEvtFileClose)
#pragma alloc_text( PAGE, immunePCICtrlEvtIoDeviceControl)
#endif // ALLOC_PRAGMA

//pci driver interface
PPCI_BUS_INTERFACE_STANDARD busInterface = NULL; 

void DebugPrintMsg(char* s)
{
	UNREFERENCED_PARAMETER(s);
	//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, s));
}

// main entry point called by OS
NTSTATUS DriverEntry(IN OUT PDRIVER_OBJECT DriverObject, IN PUNICODE_STRING RegistryPath)
{
	NTSTATUS                       status;
	WDF_DRIVER_CONFIG              config;
	WDFDRIVER                      hDriver;
	PWDFDEVICE_INIT                pInit = NULL;
	WDF_OBJECT_ATTRIBUTES          attributes;

	WDF_DRIVER_CONFIG_INIT(&config, WDF_NO_EVENT_CALLBACK);

	// we are a legacy style software device
	config.DriverInitFlags |= WdfDriverInitNonPnpDriver;
	config.EvtDriverUnload = immunePCICtrlEvtDriverUnload;

	// register cleanup callback
	WDF_OBJECT_ATTRIBUTES_INIT(&attributes);
	attributes.EvtCleanupCallback = immunePCICtrlEvtDriverContextCleanup;

	// create framework object
	status = WdfDriverCreate(DriverObject, RegistryPath, &attributes, &config, &hDriver);
	if (!NT_SUCCESS(status)) return status;

	pInit = WdfControlDeviceInitAllocate(hDriver, &SDDL_DEVOBJ_SYS_ALL_ADM_RWX_WORLD_RW_RES_R);
	if (pInit == NULL) {
		status = STATUS_INSUFFICIENT_RESOURCES;
		return status;
	}

	ExInitializeDriverRuntime(DrvRtPoolNxOptIn);

	return immunePCICtrlDeviceAdd(hDriver, pInit);
}


NTSTATUS immunePCICtrlDeviceAdd(IN WDFDRIVER Driver, IN PWDFDEVICE_INIT DeviceInit)
{
	NTSTATUS					status;
	WDF_OBJECT_ATTRIBUTES		attributes;
	UNICODE_STRING				DeviceNameU;
	WDF_IO_QUEUE_CONFIG			ioQueueConfig;
	WDF_FILEOBJECT_CONFIG		fileConfig;
	WDFQUEUE					queue;
	WDFDEVICE					controlDevice;

	UNREFERENCED_PARAMETER(Driver);
	PAGED_CODE();

	// exclusive access to device by one application at a time
	WdfDeviceInitSetExclusive(DeviceInit, TRUE);
	WdfDeviceInitSetIoType(DeviceInit, WdfDeviceIoBuffered);

	RtlInitUnicodeString(&DeviceNameU, L"\\Device\\immunePCI_ctrl");
	status = WdfDeviceInitAssignName(DeviceInit, &DeviceNameU);

	if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immunePCICtrl] WdfDeviceInitAssignName failed %u", status));
		goto exit;
	}

	status = WdfDeviceInitAssignSDDLString(DeviceInit, &SDDL_DEVOBJ_SYS_ALL_ADM_ALL);
	if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immunePCICtrl] WdfDeviceInitAssignSDDLString failed %u", status));
		goto exit;
	}

	// we are required to register for shutdown notifications
	WdfControlDeviceInitSetShutdownNotification(DeviceInit, immunePCICtrlMachineShutdown, WdfDeviceShutdown);

	// configure a file object so our device gets a device file that can be opened by drivers
	WDF_FILEOBJECT_CONFIG_INIT(&fileConfig, immunePCICtrlEvtDeviceFileCreate, immunePCICtrlEvtFileClose, WDF_NO_EVENT_CALLBACK);
	WdfDeviceInitSetFileObjectConfig(DeviceInit, &fileConfig, WDF_NO_OBJECT_ATTRIBUTES);

	// create device
	WDF_OBJECT_ATTRIBUTES_INIT(&attributes);
	status = WdfDeviceCreate(&DeviceInit, &attributes, &controlDevice);
	if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immunePCICtrl] WdfDeviceCreate failed %u", status));
		goto exit;
	}

	// use an IO queue to react to IOCTLs
	WDF_IO_QUEUE_CONFIG_INIT_DEFAULT_QUEUE(&ioQueueConfig, WdfIoQueueDispatchSequential);
	ioQueueConfig.EvtIoDeviceControl = immunePCICtrlEvtIoDeviceControl;

	//initialize pci bus interface buffer
	//XXX maybe this should be done per root complex device ??
	busInterface = (PPCI_BUS_INTERFACE_STANDARD)ExAllocatePoolZero(NonPagedPool, sizeof(PCI_BUS_INTERFACE_STANDARD), 'IMM1');
	if (busInterface == NULL)
	{
		return STATUS_INSUFFICIENT_RESOURCES;
	}
	// epxlicitly zero memory b/c ExAllocatePoolZero has issues
	//RtlZeroMemory(busInterface, sizeof(BUS_INTERFACE_STANDARD));
	RtlZeroMemory(busInterface, sizeof(PCI_BUS_INTERFACE_STANDARD));

	//
	// By default, Static Driver Verifier (SDV) displays a warning if it 
	// doesn't find the EvtIoStop callback on a power-managed queue. 
	// The 'assume' below causes SDV to suppress this warning. If the driver 
	// has not explicitly set PowerManaged to WdfFalse, the framework creates
	// power-managed queues when the device is not a filter driver.  Normally 
	// the EvtIoStop is required for power-managed queues, but for this driver
	// it is not needed b/c the driver doesn't hold on to the requests or 
	// forward them to other drivers. This driver completes the requests 
	// directly in the queue's handlers. If the EvtIoStop callback is not 
	// implemented, the framework waits for all driver-owned requests to be
	// done before moving in the Dx/sleep states or before removing the 
	// device, which is the correct behavior for this type of driver.
	// If the requests were taking an indeterminate amount of time to complete,
	// or if the driver forwarded the requests to a lower driver/another stack,
	// the queue should have an EvtIoStop/EvtIoResume.
	//
	WDF_OBJECT_ATTRIBUTES_INIT(&attributes);
	__analysis_assume(ioQueueConfig.EvtIoStop != 0);
	status = WdfIoQueueCreate(controlDevice, &ioQueueConfig, &attributes, &queue);
	__analysis_assume(ioQueueConfig.EvtIoStop == 0);
	if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immunePCICtrl] WdfIoQueueCreate failed %u", status));
		goto exit;
	}

	// Control devices must notify WDF when they are done initializing.   I/O is
	// rejected until this call is made.
	WdfControlFinishInitializing(controlDevice);

exit:
	//
	// If the device is created successfully, framework would clear the
	// DeviceInit value. Otherwise device create must have failed so we
	// should free the memory ourself.
	//
	if (DeviceInit != NULL) {
		WdfDeviceInitFree(DeviceInit);
	}

	return status;

}


VOID immunePCICtrlEvtDriverContextCleanup(IN WDFOBJECT Driver)
{
	UNREFERENCED_PARAMETER(Driver);
	PAGED_CODE();
	// no cleanup
}



VOID immunePCICtrlEvtDeviceFileCreate(IN WDFDEVICE Device, IN WDFREQUEST Request, IN WDFFILEOBJECT FileObject)
{
	UNREFERENCED_PARAMETER(Device);
	UNREFERENCED_PARAMETER(FileObject);
	PAGED_CODE();

	// just complete the request as we do not keep any state
	WdfRequestComplete(Request, STATUS_SUCCESS);

	return;
}

// no cleanup necessary as we do not keep any state
VOID immunePCICtrlEvtFileClose(IN WDFFILEOBJECT FileObject)
{
	UNREFERENCED_PARAMETER(FileObject);
	PAGED_CODE();
	return;
}

// dummy function as we do not have any state
VOID immunePCICtrlMachineShutdown(WDFDEVICE Device)
{
	UNREFERENCED_PARAMETER(Device);
	return;
}

VOID immunePCICtrlEvtDriverUnload(IN WDFDRIVER Driver)
{
	UNREFERENCED_PARAMETER(Driver);
	PAGED_CODE();

	if (busInterface && busInterface->InterfaceDereference)
	{
		(*busInterface->InterfaceDereference)(busInterface->Context);
		ExFreePool(busInterface);
	}

	return;
}

// gets the bus interface standard information from the PDO.
NTSTATUS GetPCIBusInterfaceStandard(IN WDFDEVICE Device, OUT PPCI_BUS_INTERFACE_STANDARD BusInterfaceStandard)
{
	NTSTATUS status;

	WDF_OBJECT_ATTRIBUTES  ioTargetAttrib;
	WDFIOTARGET  ioTarget;
	WDF_IO_TARGET_OPEN_PARAMS  openParams;
	UNICODE_STRING devNameU;

	WDF_OBJECT_ATTRIBUTES_INIT(&ioTargetAttrib);

	status = WdfIoTargetCreate(Device, &ioTargetAttrib, &ioTarget);

	if (!NT_SUCCESS(status))
		return status;

	RtlInitUnicodeString(&devNameU, L"\\Device\\immunePCI_flt");
	WDF_IO_TARGET_OPEN_PARAMS_INIT_OPEN_BY_NAME(
		&openParams,
		&devNameU,
		STANDARD_RIGHTS_ALL
	);

	status = WdfIoTargetOpen(ioTarget, &openParams);

	if (!NT_SUCCESS(status))
	{
		WdfObjectDelete(ioTarget);
		return status;
	}

	status = WdfIoTargetQueryForInterface(
		ioTarget,
		(LPGUID)&GUID_PCI_BUS_INTERFACE_STANDARD,
		(PINTERFACE)BusInterfaceStandard,
		sizeof(PCI_BUS_INTERFACE_STANDARD),
		PCI_BUS_INTERFACE_STANDARD_VERSION,
		NULL
	);

	if (!NT_SUCCESS(status))
		WdfObjectDelete(ioTarget);

	// the ioTarget will be registered with our device object and
	// the framework handles its lifecycle automatically, it will
	// close and free it on unloading and re-open it when removal is canceled

	return status;
}

//read pci configuration
static NTSTATUS ReadPciConfig(IN WDFDEVICE Device, OUT PPHYMEM_PCI pPci, OUT PVOID OutputBuffer)
{
	NTSTATUS status = 0;

	//get bus interface
	if (busInterface == NULL || busInterface->ReadConfig == NULL)
	{
		status = GetPCIBusInterfaceStandard(Device, busInterface);

		if (!NT_SUCCESS(status))
		{
			//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "Get pci bus driver interface failed, code=0x%x\n", ntStatus));
		}
	}

	if (NT_SUCCESS(status) && busInterface != NULL && busInterface->ReadConfig != NULL)
	{
		PCI_SLOT_NUMBER slot;
		ULONG ulRet;
		KIRQL oldIRQL;

		slot.u.AsULONG = 0;
		slot.u.bits.DeviceNumber = pPci->dwDevNum;
		slot.u.bits.FunctionNumber = pPci->dwFuncNum;

		oldIRQL = KeGetCurrentIrql();
		if (oldIRQL < DISPATCH_LEVEL)
			oldIRQL = KeRaiseIrqlToDpcLevel();
		ulRet = (*busInterface->ReadConfig)(busInterface->Context, //context
			(UCHAR)pPci->dwBusNum, //busoffset
			slot.u.AsULONG,		   //slot
			OutputBuffer,		   //buffer
			pPci->dwRegOff,		   //offset
			pPci->dwBytes);		   //length
		if(oldIRQL < DISPATCH_LEVEL)
			KeLowerIrql(oldIRQL);

		if (ulRet == pPci->dwBytes)
		{
			status = STATUS_SUCCESS;

			//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "Read %d bytes from Bus %d Dev %d Fun %d\n", ulRet, pPci->dwBusNum, pPci->dwDevNum, pPci->dwFuncNum));
		}
		else
			status = STATUS_UNSUCCESSFUL;
	}
	else
		status = STATUS_INVALID_PARAMETER;

	return status;
}

/*++
Routine Description:

	This event is called when the framework receives IRP_MJ_DEVICE_CONTROL
	requests from the system.
--*/
VOID immunePCICtrlEvtIoDeviceControl(IN WDFQUEUE Queue, IN WDFREQUEST Request, IN size_t OutputBufferLength, IN size_t InputBufferLength, IN ULONG IoControlCode)
{
	NTSTATUS            status = STATUS_SUCCESS;// Assume success
	PCHAR               inBuf = NULL, outBuf = NULL; // pointer to Input and output buffer
	size_t              bufSize;

	UNREFERENCED_PARAMETER(Queue);

	PAGED_CODE();

	if (!OutputBufferLength || !InputBufferLength)
	{
		WdfRequestComplete(Request, STATUS_INVALID_PARAMETER);
		return;
	}

	status = WdfRequestRetrieveInputBuffer(Request, 0, &inBuf, &bufSize);
	if (!NT_SUCCESS(status)) {
		status = STATUS_INSUFFICIENT_RESOURCES;
		goto exit;
	}
	ASSERT(bufSize == InputBufferLength);
	ASSERT(inBuf != NULL);

	status = WdfRequestRetrieveOutputBuffer(Request, 0, &outBuf, &bufSize);
	if (!NT_SUCCESS(status)) {
		status = STATUS_INSUFFICIENT_RESOURCES;
		goto exit;
	}
	ASSERT(bufSize == OutputBufferLength);
	ASSERT(outBuf != NULL);

	switch (IoControlCode)
	{
		case IOCTL_PHYMEM_GETPCI:
		{
			PPHYMEM_PCI pPci = (PPHYMEM_PCI)inBuf;

			//register offset + bytes to read cannnot exceed 4096 (pci config space limit)
			if (InputBufferLength == sizeof(PHYMEM_PCI) &&
				((pPci->dwRegOff + pPci->dwBytes) <= 4096) && (OutputBufferLength >= pPci->dwBytes))
			{
				status = ReadPciConfig(WdfIoQueueGetDevice(Queue), pPci, (PVOID)outBuf);

				if (NT_SUCCESS(status))
					WdfRequestSetInformation(Request, pPci->dwBytes);
			}
			else
				status = STATUS_INVALID_PARAMETER;

			break;
		}

		default:
			status = STATUS_INVALID_DEVICE_REQUEST;
			DebugPrintMsg("Error: Unknown IO CONTROL CODE\n");
			break;
	}

exit:
	WdfRequestComplete(Request, status);
}
