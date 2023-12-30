import { Menu, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/outline';
import { useSession } from 'hooks/useSession';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import React, { useContext } from 'react';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';

function OrganisationSelect(): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);

  const {
    session: {
      memberships,
      currentMembership,
      user: { role },
    },
    setCurrentMembership,
  } = useSession();

  const listLength = role === 'admin' ? 'max-h-[10rem]' : 'max-h-[7rem]';

  return (
    <Menu>
      <div className="w-full relative" data-testid="org-select-menu">
        <Menu.Button
          className="text-white bg-purple-400 p-3 px-3 w-full text-left flex justify-between items-center font-bold"
          data-qa="org-select"
          role="button">
          <span className="whitespace-nowrap overflow-ellipsis overflow-hidden">
            {currentMembership?.organisation?.name}
          </span>
          {memberships && memberships.length > 1 && <ChevronDownIcon height="20" />}
        </Menu.Button>
        <Transition>
          <Menu.Items
            as="ul"
            className={`text-white bg-purple-400 px-3 w-full text-left absolute divide-y z-10	overflow-auto ${listLength}`}>
            {memberships &&
              memberships
                .filter((membership) => membership.id != currentMembership.id)
                .map((membership) => (
                  <Menu.Item
                    as="li"
                    className="py-3 border-t cursor-pointer"
                    onClick={() => {
                      selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
                      setCurrentMembership(membership);
                    }}
                    key={membership.id}>
                    {membership?.organisation?.name}
                  </Menu.Item>
                ))}
          </Menu.Items>
        </Transition>
      </div>
    </Menu>
  );
}
export default OrganisationSelect;
