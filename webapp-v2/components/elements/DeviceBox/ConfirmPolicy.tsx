import Button from 'components/elements/Button/Button';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import Toggle from 'components/elements/Toggle/Toggle';
import { useUpdateDevice, USE_DEVICES_URL } from 'hooks/devices';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { useContext, useState, useEffect } from 'react';
import { toast } from 'react-toastify';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { mutate } from 'swr';

export default function ConfirmPolicy(): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);
  const updateDevice = useUpdateDevice();

  const [endpointProtection, setEndpointProtection] = useState(false);
  const [intelTsc, setIntelTsc] = useState(false);

  useEffect(() => {
    const previousEndpointProtection =
      selectedDevices.items.every((device) => device.policy.endpoint_protection === 'on') &&
      !!selectedDevices.items.length;
    const previousIntelTsc =
      selectedDevices.items.every((device) => device.policy.intel_tsc === 'on') &&
      !!selectedDevices.items.length;

    setEndpointProtection(previousEndpointProtection);
    setIntelTsc(previousIntelTsc);
  }, [selectedDevices.items.length]);

  return (
    <ConfirmModal
      headline="Set Security Policy"
      confirmLabel={
        selectedDevices.items.length > 0
          ? 'Set Security Policy for multiple devices'
          : 'Set Security Policy'
      }
      onConfirm={() => {
        Promise.all(
          selectedDevices.items.map((device) =>
            updateDevice.mutate({
              id: device.id,
              attributes: {
                policy: {
                  intel_tsc: intelTsc ? 'on' : 'if-present',
                  endpoint_protection: endpointProtection ? 'on' : 'if-present',
                },
              },
            }),
          ),
        ).then(() => {
          mutate(USE_DEVICES_URL)
            .then((response: { data: any; error: any }) => {
              if (response?.error || !response?.data) {
                toast.error('Something went wrong');
              } else {
                toast.success('Device policy was successfully updated.');
              }
            })
            .then(() => {
              selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
              location.reload();
            });
        });
      }}
      TriggerComponent={(props) => (
        <Button theme="CTA" full={true} {...props} disabled={selectedDevices.items.length == 0}>
          Set Security Policy
        </Button>
      )}>
      <div className="-mt-4">
        <b>{selectedDevices.items.length}</b> Devices selected
        <div className="mt-8 mb-4">
          Enforces the following security features to be enabled before marking a device trusted.
        </div>
        <div className="flex space-y-4 flex-col">
          <div className="grid grid-cols-[auto,1fr] items-start gap-x-2">
            <Toggle
              checked={endpointProtection}
              label=" "
              onChange={() => setEndpointProtection((current) => !current)}
            />
            {/* " " label is needed */}
            <div>
              <b className="block">Require Endpoint Protection</b>
              <span>
                Only mark the device as trusted if an Endpoint Protection solution is installed and
                enabled on it.
              </span>
            </div>
          </div>
          <div className="grid grid-cols-[auto,1fr] items-start gap-x-2">
            <Toggle
              checked={intelTsc}
              label=" "
              onChange={() => setIntelTsc((current) => !current)}
            />
            {/* " " label is needed */}
            <div>
              <b className="block">Require Intel Transparent Supply Chain</b>
              <span>
                Only mark the device trusted if its serial number is registered to Intel’s
                Transparent Supply Chain database. immune Guard will make sure the device’s
                components are matching those in the database.
              </span>
            </div>
          </div>
        </div>
      </div>
    </ConfirmModal>
  );
}
