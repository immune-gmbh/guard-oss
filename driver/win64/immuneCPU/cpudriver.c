#include <ntddk.h>
#include <wdf.h>
#include <intrin.h>
#include "cpudriver.h"
#include "cpu.h"

DRIVER_INITIALIZE DriverEntry;
EVT_WDF_DRIVER_UNLOAD immuneCPUEvtDriverUnload;
EVT_WDF_DEVICE_CONTEXT_CLEANUP immuneCPUEvtDriverContextCleanup;
EVT_WDF_DEVICE_SHUTDOWN_NOTIFICATION immuneCPUMachineShutdown;
EVT_WDF_IO_QUEUE_IO_DEVICE_CONTROL immuneCPUEvtIoDeviceControl;
EVT_WDF_DEVICE_FILE_CREATE immuneCPUEvtDeviceFileCreate;
EVT_WDF_FILE_CLOSE immuneCPUEvtFileClose;


// Don't use EVT_WDF_DRIVER_DEVICE_ADD for immuneCPUDeviceAdd even though 
// the signature is same because this is not an event called by the framework.
NTSTATUS immuneCPUDeviceAdd(IN WDFDRIVER Driver, IN PWDFDEVICE_INIT DeviceInit);

#ifdef ALLOC_PRAGMA
#pragma alloc_text( INIT, DriverEntry )
#pragma alloc_text( PAGE, immuneCPUDeviceAdd)
#pragma alloc_text( PAGE, immuneCPUEvtDriverContextCleanup)
#pragma alloc_text( PAGE, immuneCPUEvtDriverUnload)
#pragma alloc_text( PAGE, immuneCPUEvtDeviceFileCreate)
#pragma alloc_text( PAGE, immuneCPUEvtFileClose)
#pragma alloc_text( PAGE, immuneCPUEvtIoDeviceControl)
#endif // ALLOC_PRAGMA

void DebugPrintMsg(char* s)
{
	UNREFERENCED_PARAMETER(s);
	//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, s));
}

// main entry point called by OS
NTSTATUS DriverEntry(IN OUT PDRIVER_OBJECT DriverObject, IN PUNICODE_STRING RegistryPath )
{
	NTSTATUS                       status;
	WDF_DRIVER_CONFIG              config;
	WDFDRIVER                      hDriver;
	PWDFDEVICE_INIT                pInit = NULL;
	WDF_OBJECT_ATTRIBUTES          attributes;

	WDF_DRIVER_CONFIG_INIT(&config, WDF_NO_EVENT_CALLBACK);

	// we are a legacy style software device
	config.DriverInitFlags |= WdfDriverInitNonPnpDriver;
	config.EvtDriverUnload = immuneCPUEvtDriverUnload;

	// register cleanup callback
	WDF_OBJECT_ATTRIBUTES_INIT(&attributes);
	attributes.EvtCleanupCallback = immuneCPUEvtDriverContextCleanup;

	// create framework object
	status = WdfDriverCreate(DriverObject, RegistryPath, &attributes, &config, &hDriver);
	if (!NT_SUCCESS(status)) return status;

	pInit = WdfControlDeviceInitAllocate(hDriver, &SDDL_DEVOBJ_SYS_ALL_ADM_RWX_WORLD_RW_RES_R);
	if (pInit == NULL) {
		status = STATUS_INSUFFICIENT_RESOURCES;
		return status;
	}

	return immuneCPUDeviceAdd(hDriver, pInit);
}


NTSTATUS immuneCPUDeviceAdd(IN WDFDRIVER Driver, IN PWDFDEVICE_INIT DeviceInit)
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

	RtlInitUnicodeString(&DeviceNameU, L"\\Device\\immuneCPU");
    status = WdfDeviceInitAssignName(DeviceInit, &DeviceNameU);

    if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immuneCPU] WdfDeviceInitAssignName failed %u", status));
        goto exit;
    }

	status = WdfDeviceInitAssignSDDLString(DeviceInit, &SDDL_DEVOBJ_SYS_ALL_ADM_ALL);
	if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immuneCPU] WdfDeviceInitAssignSDDLString failed %u", status));
		goto exit;
	}

	// we are required to register for shutdown notifications
    WdfControlDeviceInitSetShutdownNotification(DeviceInit, immuneCPUMachineShutdown, WdfDeviceShutdown);

	// configure a file object so our device gets a device file that can be opened by drivers
    WDF_FILEOBJECT_CONFIG_INIT(&fileConfig, immuneCPUEvtDeviceFileCreate, immuneCPUEvtFileClose, WDF_NO_EVENT_CALLBACK);
    WdfDeviceInitSetFileObjectConfig(DeviceInit, &fileConfig, WDF_NO_OBJECT_ATTRIBUTES);

    // create device
	WDF_OBJECT_ATTRIBUTES_INIT(&attributes);
    status = WdfDeviceCreate(&DeviceInit, &attributes, &controlDevice);
    if (!NT_SUCCESS(status)) {
		//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immuneCPU] WdfDeviceCreate failed %u", status));
        goto exit;
    }

    // use an IO queue to react to IOCTLs
    WDF_IO_QUEUE_CONFIG_INIT_DEFAULT_QUEUE(&ioQueueConfig, WdfIoQueueDispatchSequential);
    ioQueueConfig.EvtIoDeviceControl = immuneCPUEvtIoDeviceControl;

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
        //KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "[immuneCPU] WdfIoQueueCreate failed %u", status));
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


VOID immuneCPUEvtDriverContextCleanup(IN WDFOBJECT Driver)
{
	UNREFERENCED_PARAMETER(Driver);
    PAGED_CODE();
	// no cleanup
}



VOID immuneCPUEvtDeviceFileCreate(IN WDFDEVICE Device, IN WDFREQUEST Request, IN WDFFILEOBJECT FileObject)
{
	UNREFERENCED_PARAMETER(Device);
    UNREFERENCED_PARAMETER(FileObject);
    PAGED_CODE();

	// just complete the request as we do not keep any state
    WdfRequestComplete(Request, STATUS_SUCCESS);

    return;
}

// no cleanup necessary as we do not keep any state
VOID immuneCPUEvtFileClose(IN WDFFILEOBJECT FileObject)
{
	UNREFERENCED_PARAMETER(FileObject);
    PAGED_CODE();
    return;
}

// dummy function as we do not have any state
VOID immuneCPUMachineShutdown(WDFDEVICE Device)
{
	UNREFERENCED_PARAMETER(Device);
	return;
}

// dummy function as we do not have any state
VOID immuneCPUEvtDriverUnload(IN WDFDRIVER Driver)
{
	UNREFERENCED_PARAMETER(Driver);
	PAGED_CODE();
	return;
}

NTSTATUS ReadPhysicalMemory(PHYSICAL_ADDRESS pa, unsigned int len, void* pData)
{
	void* va = MmMapIoSpace(pa, len, MmCached);
	if (!va)
	{
		DebugPrintMsg("ERROR: no space for mapping\n");
		return STATUS_UNSUCCESSFUL;
	}
	//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "reading %u bytes from physical address 0x%08x_%08x (virtual = %#010x)", len, pa.HighPart, pa.LowPart, (UINTN)va));
	RtlCopyMemory(pData, va, len);
	MmUnmapIoSpace(va, len);
	return STATUS_SUCCESS;
}

NTSTATUS CheckMSRAllowed(UINT32 msr_addr)
{
	// hide some registers that hint where the kernel is located in memory
	switch (msr_addr) {
		// AMD syscall / sysret
		case 0xC0000081: // STAR
		case 0xC0000082: // LSTAR
		case 0xC0000083: // CSTAR

		// Intel systenter / sysexit
		case 0x174: // IA32_SYSENTER_CS
		case 0x175: // IA32_SYSENTER_ESP
		case 0x176: // IA32_SYSENTER_EIP

		return STATUS_INVALID_PARAMETER;

		default:
		break;
	}
	return STATUS_SUCCESS;
}

/*++
Routine Description:

    This event is called when the framework receives IRP_MJ_DEVICE_CONTROL
    requests from the system.

Arguments:

    Queue - Handle to the framework queue object that is associated
            with the I/O request.
    Request - Handle to a framework request object.

    OutputBufferLength - length of the request's output buffer,
                        if an output buffer is available.
    InputBufferLength - length of the request's input buffer,
                        if an input buffer is available.

    IoControlCode - the driver-defined or system-defined I/O control code
                    (IOCTL) that is associated with the request.

Return Value:

   VOID

--*/
VOID immuneCPUEvtIoDeviceControl(IN WDFQUEUE Queue, IN WDFREQUEST Request, IN size_t OutputBufferLength, IN size_t InputBufferLength, IN ULONG IoControlCode)
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
		case IOCTL_IMMUNE_GETFLASH:
		{
			// First 16MB of platform flash is mapped @ 4GB - 16MB
			UINT32 len = 0x1000000;
			PHYSICAL_ADDRESS phys_addr = { .HighPart = 0x0,
											.LowPart = 0xFF000000 };

			if (OutputBufferLength < len)
			{
				status = STATUS_BUFFER_TOO_SMALL;
				break;
			}

			__try
			{
				status = ReadPhysicalMemory(phys_addr, len, outBuf);
			}
			__except (EXCEPTION_EXECUTE_HANDLER)
			{
				status = GetExceptionCode();
				//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "IOCTL_IMMUNE_GETFLASH exception code 0x%X\n", status));
				break;
			}

			if (NT_SUCCESS(status))
			{
				WdfRequestSetInformation(Request, len);
			}
			break;
		}

		case IOCTL_IMMUNE_GETTXTPUB:
		{
			// Get 64KiB of TXT public space
			UINT32 len = 0x10000;
			PHYSICAL_ADDRESS phys_addr = { .HighPart = 0x0,
											.LowPart = 0xFED30000 };
			UINT32 cpuidValue[4];
			unsigned char* cv;
			UINT64 msrValue;
			UINT32* pMsrValue = (UINT32*)&msrValue;

			if (OutputBufferLength < len)
			{
				status = STATUS_BUFFER_TOO_SMALL;
				break;
			}

			// check if this is an Intel CPU
			__cpuidex((int*)cpuidValue, CPUID_VENDOR, 0);
			cv = (unsigned char*)(&cpuidValue[1]);
			if ((cv[0] != 'G') ||
				(cv[1] != 'e') ||
				(cv[2] != 'n') ||
				(cv[3] != 'u') ||
				(cv[8] != 'i') ||
				(cv[9] != 'n') ||
				(cv[10] != 'e') ||
				(cv[11] != 'I') ||
				(cv[4] != 'n') ||
				(cv[5] != 't') ||
				(cv[6] != 'e') ||
				(cv[7] != 'l'))
			{
				status = STATUS_NOT_IMPLEMENTED;
				break;
			}

			// check if SMX is supported (ECX bit 6)
			__cpuidex((int*)cpuidValue, CPUID_MODEL, 0);
			if ((cpuidValue[CPUID_ECX] & CPUID_MODEL_SMX_bm) != CPUID_MODEL_SMX_bm)
			{
				status = STATUS_NOT_IMPLEMENTED;
				break;
			}

			// check if this system has global SENTER enabled and at leats one GETSEC leaf instruction enabled
			_rdmsr(MSR_IA32_FEATURE_CONTROL, &pMsrValue[0], &pMsrValue[1]);
			if (((msrValue & SENTER_GLOBAL_ENABLE_bm) != SENTER_GLOBAL_ENABLE_bm) || ((msrValue & SENTER_LEAF_ENABLES_bm) == 0))
			{
				status = STATUS_NOT_IMPLEMENTED;
				break;
			}

			// after passing all those checks we can probably safely assume that we're going to read the TXT public space
			__try
			{
				status = ReadPhysicalMemory(phys_addr, len, outBuf);
			}
			__except (EXCEPTION_EXECUTE_HANDLER)
			{
				status = GetExceptionCode();
				break;
			}

			if (NT_SUCCESS(status))
			{
				WdfRequestSetInformation(Request, len);
			}
			break;
		}

		case IOCTL_IMMUNE_GETMSR:
		{
			UINT32				_eax = 0, _edx = 0, _msr_addr = 0;
			unsigned int		new_cpu_thread_id = 0;
			ULONG				_num_active_cpus = 0;
			KAFFINITY			_kaffinity = 0;
			PROCESSOR_NUMBER	_proc_number = { 0, 0, 0 };

			_num_active_cpus = KeQueryActiveProcessorCountEx(ALL_PROCESSOR_GROUPS);
			KeGetCurrentProcessorNumberEx(&_proc_number);
			_kaffinity = KeQueryGroupAffinity(_proc_number.Group);

			if (!inBuf)
			{
				status = STATUS_INVALID_PARAMETER;
				break;
			}
			if (InputBufferLength < 2 * sizeof(UINT32))
			{
				status = STATUS_INVALID_PARAMETER;
				break;
			}

			RtlCopyBytes(&new_cpu_thread_id, (BYTE*)inBuf, sizeof(UINT32));
			if (new_cpu_thread_id >= _num_active_cpus)
			{
				//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "IOCTL_IMMUNE_GETMSR Invalid thread id %d\n", new_cpu_thread_id));
				status = STATUS_INVALID_PARAMETER;
				break;
			}

			_kaffinity = KeSetSystemAffinityThreadEx((KAFFINITY)(1ULL << new_cpu_thread_id));

			RtlCopyBytes(&_msr_addr, (BYTE*)inBuf + sizeof(UINT32), sizeof(UINT32));
			status = CheckMSRAllowed(_msr_addr);
			if (status != STATUS_SUCCESS) {
				DebugPrintMsg("requested protected MSR");
				break;
			}

			__try
			{
				_rdmsr(_msr_addr, &_eax, &_edx);
			}
			__except (EXCEPTION_EXECUTE_HANDLER)
			{
				status = GetExceptionCode();
				//KdPrintEx((DPFLTR_IHVDRIVER_ID, DPFLTR_INFO_LEVEL, "IOCTL_IMMUNE_GETMSR exception code 0x%X\n", status));
				break;
			}

			if (OutputBufferLength >= 2 * sizeof(UINT32))
			{
				RtlCopyBytes(outBuf, (VOID*)&_eax, sizeof(UINT32));
				RtlCopyBytes(((UINT8*)outBuf) + sizeof(UINT32), (VOID*)&_edx, sizeof(UINT32));
				WdfRequestSetInformation(Request, 2 * sizeof(UINT32));
				status = STATUS_SUCCESS;
			}
			else
			{
				status = STATUS_BUFFER_TOO_SMALL;
			}
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
