import { TrashIcon } from '@heroicons/react/outline';
import classNames from 'classnames';
import DeviceBoxMini from 'components/elements/DeviceBox/DeviceBoxMini';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { useContext } from 'react';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';

import ConfirmPolicy from './ConfirmPolicy';
import ConfirmRetire from './ConfirmRetire';
import ConfirmTags from './ConfirmTags';

export default function DevicesSidebar(): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);

  return (
    <aside
      className={classNames(
        'sticky top-8 p-8 bg-purple-600 rounded-md h-[min-content] max-h-[calc(100vh-4rem)]',
        {
          'overflow-hidden': selectedDevices.items?.length === 0,
        },
      )}>
      <div className="flex items-center justify-between">
        <strong className="text-4xl">Selected</strong>
        <span className="text-xl">{selectedDevices.items?.length} devices</span>
      </div>
      <section className="space-y-4 mt-6 mb-2 max-h-[calc(100vh-32rem)] min-h-[70px] overflow-y-auto">
        {selectedDevices.items?.map((device) => (
          <DeviceBoxMini
            key={device.id}
            device={device}
            onClick={() =>
              selectedDevices.dispatch({
                type: DeviceActionType.TOGGLE,
                device,
              })
            }
          />
        ))}
      </section>
      <div className="flex justify-end">
        <div
          className="flex cursor-pointer space-x-2"
          onClick={() => selectedDevices.dispatch({ type: DeviceActionType.CLEAR })}>
          <TrashIcon width={20} />
          <span className="underline text-xl font-bold">clear</span>
        </div>
      </div>
      <div className="space-y-4 mt-6">
        <ConfirmRetire />
        <ConfirmTags />
        <ConfirmPolicy />
      </div>
    </aside>
  );
}
