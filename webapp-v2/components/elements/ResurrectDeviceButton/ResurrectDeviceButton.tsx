import Button, { ImmuneButtonProps } from 'components/elements/Button/Button';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useResurrectDevice, USE_DEVICES_URL } from 'hooks/devices';
import { useState } from 'react';
import { mutate } from 'swr';

interface IResurrectDeviceButtonProps extends ImmuneButtonProps {
  deviceId: string;
}

const ResurrectDeviceButton = ({ deviceId, ...rest }: IResurrectDeviceButtonProps): JSX.Element => {
  const [submitted, setSubmitted] = useState(false);

  const resurectDevice = useResurrectDevice();
  return (
    <Button
      disabled={submitted}
      onClick={() => {
        setSubmitted(true);
        resurectDevice
          .mutate({
            id: deviceId,
          })
          .then(() => {
            mutate(USE_DEVICES_URL).then(() => {
              window.location.href = NextJsRoutes.dashboardDevicesIndexPath;
            });
          });
      }}
      theme="MAIN"
      {...rest}>
      {submitted && <LoadingSpinner />}
      Restore
    </Button>
  );
};
export default ResurrectDeviceButton;
