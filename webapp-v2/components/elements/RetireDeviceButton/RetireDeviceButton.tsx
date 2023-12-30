import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import { useRetireDevice, USE_DEVICES_URL } from 'hooks/devices';
import { useState } from 'react';
import { mutate } from 'swr';

interface IRetireDeviceButtonProps {
  deviceId: string;
}

const RetireDeviceButton = ({ deviceId }: IRetireDeviceButtonProps): JSX.Element => {
  const [submitted, setSubmitted] = useState(false);

  const retireDevice = useRetireDevice();
  return (
    <ConfirmModal
      headline="Do you really want to retire this device?"
      confirmLabel="Retire"
      onConfirm={() => {
        setSubmitted(true);
        retireDevice
          .mutate({
            id: deviceId,
          })
          .then(() => {
            mutate(USE_DEVICES_URL).then(() => {
              location.reload();
            });
          });
      }}
      TriggerComponent={(props) => (
        <button disabled={submitted} className="block underline font-bold flex" {...props}>
          {submitted && <LoadingSpinner />}
          Retire device
        </button>
      )}
    />
  );
};
export default RetireDeviceButton;
