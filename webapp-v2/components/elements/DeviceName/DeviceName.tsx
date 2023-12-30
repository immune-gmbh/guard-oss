import { CheckIcon, PencilIcon } from '@heroicons/react/solid';
import { useRenameDevice } from 'hooks/devices';
import React, { KeyboardEvent, useEffect, useRef, useState } from 'react';
import { toast } from 'react-toastify';
import { ApiSrv } from 'types/apiSrv';

interface IDeviceName {
  device: ApiSrv.Device;
}

const DeviceName: React.FC<IDeviceName> = ({ device }) => {
  const [inEditModus, setEditModus] = useState(false);
  const [deviceName, setDeviceName] = useState(device?.name);
  const [textWidth, setTextWidth] = useState(0);
  const spanElement = useRef<HTMLInputElement>(null);
  const inputElement = useRef<HTMLInputElement>(null);
  const renameDevice = useRenameDevice();

  const handleClick = (): void => {
    if (inEditModus) {
      if (deviceName.length === 0) {
        toast.error('Please enter at least one letter');
        inputElement.current?.focus();
        return;
      } else {
        renameDevice
          .mutate({ id: device.id, name: deviceName })
          .then((response?: { code: number }) => {
            if (response?.code && response?.code === 200) {
              toast.success('Device name was successfully updated.');
            } else {
              toast.error('Something went wrong');
            }
          });
      }
    } else {
      setTextWidth(spanElement.current?.getBoundingClientRect().width);
    }

    setEditModus(!inEditModus);
  };

  useEffect(() => {
    if (spanElement.current) {
      setTextWidth(spanElement.current.getBoundingClientRect().width);
    }
  }, [spanElement]);

  useEffect(() => {
    if (inEditModus) {
      inputElement.current?.focus();
    }
  }, [inEditModus]);

  const handleKeyUp = (event: KeyboardEvent<HTMLInputElement>): void => {
    if (event.key === 'Enter') {
      handleClick();
    }
  };

  const CurrentIcon = inEditModus ? CheckIcon : PencilIcon;
  return (
    <div className="grid grid-cols-[max-content,1fr] items-center">
      {inEditModus ? (
        <input
          type="text"
          className="border-0 border-b-2 text-3xl p-0 m-0 text-purple-500 font-bold"
          value={deviceName}
          onChange={(event) => setDeviceName(event.currentTarget.value)}
          onKeyUp={handleKeyUp}
          style={{ width: `${textWidth}px` }}
          ref={inputElement}
        />
      ) : (
        <span
          ref={spanElement}
          className="text-3xl p-0 m-0 border-b-2 border-transparent text-purple-500 font-bold">
          {deviceName}
        </span>
      )}
      <CurrentIcon
        className="ml-4 h-6 text-purple-500 hover:text-primary cursor-pointer"
        onClick={handleClick}
      />
    </div>
  );
};

export default DeviceName;
