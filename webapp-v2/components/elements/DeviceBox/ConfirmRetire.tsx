import Button from 'components/elements/Button/Button';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import { UNRETIREABLE_STATES, useRetireDevice, USE_DEVICES_URL } from 'hooks/devices';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { useContext, useMemo } from 'react';
import { toast } from 'react-toastify';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { mutate } from 'swr';

export default function ConfirmRetire(): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);

  const retireDevice = useRetireDevice();

  const retireableDevices = useMemo(
    () => selectedDevices.items?.filter((dev) => !UNRETIREABLE_STATES.includes(dev.state)),
    [selectedDevices],
  );

  return (
    <ConfirmModal
      headline={`Do you really want to retire the following ${retireableDevices.length} devices?`}
      confirmLabel="Retire"
      onConfirm={() => {
        Promise.all(
          retireableDevices.map((device) =>
            retireDevice.mutate({
              id: device.id,
            }),
          ),
        ).then(() => {
          mutate(USE_DEVICES_URL).then(() => {
            toast.success(`Retired ${retireableDevices.length} devices`);
            selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
          });
        });
      }}
      TriggerComponent={(props) => (
        <Button theme="SECONDARY" full={true} {...props} disabled={retireableDevices.length == 0}>
          Retire {retireableDevices.length} devices
        </Button>
      )}>
      <ul className="list-disc list-inside">
        {retireableDevices.map((device) => (
          <li key={device.id}>{device.name}</li>
        ))}
      </ul>
    </ConfirmModal>
  );
}
