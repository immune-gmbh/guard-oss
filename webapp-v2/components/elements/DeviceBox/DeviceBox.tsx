import classNames from 'classnames';
import ResurrectDeviceButton from 'components/elements/ResurrectDeviceButton/ResurrectDeviceButton';
import DeviceTag from 'components/elements/Tag/DeviceTag';
import NextJsRoutes from 'generated/NextJsRoutes';
import Link from 'next/link';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { useContext } from 'react';
import { ApiSrv } from 'types/apiSrv';

interface IDeviceBox {
  device: ApiSrv.Device;
  onCheckboxClick?: (device: ApiSrv.Device) => void;
  selected?: boolean;
  onTagClick?: (tag: ApiSrv.Tag) => void;
}

function DeviceBox({ device, onCheckboxClick, onTagClick, selected }: IDeviceBox): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);

  return (
    <div
      className={classNames(
        'grid grid-cols-[80px,1fr,40px] justify-center items-center w-full border rounded-m',
        {
          'opacity-40': selected,
          'border-dashed': selectedDevices.items.length > 0,
        },
      )}
      role="listbox">
      {onCheckboxClick && (
        <div
          className="h-full w-full flex justify-center items-center"
          onClick={() => onCheckboxClick(device)}>
          <input
            type="checkbox"
            className="border-gray-800 justify-self-center jsx-checkbox cursor-pointer"
            checked={selected}
            readOnly={true}
          />
        </div>
      )}
      <Link
        passHref
        href={{
          pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
          query: { did: device.id },
        }}>
        <a
          onClick={(e) => {
            if (selectedDevices.items.length > 0) {
              e.preventDefault();
              e.stopPropagation();
              onCheckboxClick(device);
            }
          }}
          className={classNames('cursor-pointer', {
            'flex col-span-2 justify-between': !onCheckboxClick,
          })}>
          <div className="relative grid py-4 grid-rows-1">
            <div className="flex flex-col, space-x-4">
              <span className="font-bold text-lg truncate">{device.name}</span>
              <div className="flex flex-row space-x-2">
                {device.tags?.map((tag, index) => (
                  <DeviceTag
                    tag={tag}
                    key={`d-${device.id}-t-${index}`}
                    onClick={() => {
                      onTagClick && onTagClick(tag);
                    }}
                  />
                ))}
              </div>
            </div>
          </div>
          <div>
            {device.state === 'resurrectable' && <ResurrectDeviceButton deviceId={device.id} />}
          </div>
        </a>
      </Link>
    </div>
  );
}

export default DeviceBox;
