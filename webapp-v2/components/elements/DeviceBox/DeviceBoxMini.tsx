import { CalendarIcon } from '@heroicons/react/outline';
import { XCircleIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import { ApiSrv } from 'types/apiSrv';

interface IDeviceBoxMini {
  device: ApiSrv.Device;
  onClick?: (device: ApiSrv.Device) => void;
  selected?: boolean;
}

const DeviceBoxMini: React.FC<IDeviceBoxMini> = ({
  device,
  device: { name },
  onClick,
  selected = false,
}) => {
  const handleClick = (): void => {
    onClick && onClick(device);
  };

  return (
    <div
      className={classNames('w-full border bg-white rounded-m p-4', {
        'opacity-40': selected,
      })}>
      <div className="flex items-center justify-between">
        <div className="flex items-center w-10/12 space-x-1">
          {device.appraisals?.length > 0 && <CalendarIcon height={20} color="#673355" />}
          <span className="text-base truncate w-5/6">{name}</span>
        </div>
        <div className="flex items-center space-x-1">
          <XCircleIcon
            onClick={handleClick}
            height={20}
            color="#B00020"
            className="cursor-pointer"
          />
        </div>
      </div>
    </div>
  );
};

export default DeviceBoxMini;
