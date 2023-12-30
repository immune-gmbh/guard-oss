import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import { useResurrectDevice, USE_DEVICES_URL } from 'hooks/devices';
import { useState } from 'react';
import { mutate } from 'swr';

interface IResurrectDeviceLinkProps {
  deviceId: string;
}

const ResurrectDeviceLink = ({ deviceId }: IResurrectDeviceLinkProps): JSX.Element => {
  const [submitted, setSubmitted] = useState(false);

  const resurectDevice = useResurrectDevice();
  return (
    <button
      onClick={() => {
        setSubmitted(true);
        resurectDevice
          .mutate({
            id: deviceId,
          })
          .then(() => {
            mutate(USE_DEVICES_URL).then(() => {
              location.reload();
            });
          });
      }}
      className="block underline font-bold flex">
      {submitted && <LoadingSpinner />}
      Undo
    </button>
  );
};
export default ResurrectDeviceLink;
