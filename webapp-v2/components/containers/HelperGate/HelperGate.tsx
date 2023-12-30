import { useSession } from 'hooks/useSession';
import { useRouter } from 'next/router';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import React, { useContext, useEffect } from 'react';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { SerializedMembership } from 'utils/types';

type IHelperGateContainer = React.HTMLProps<HTMLDivElement> & { mutateFn?: () => void };

// This Gate could be used as first logic point after the SessionProvider, to:
// - handle jump in points e.g.: Email Links
// - catch errors and handle them
export default function HelperGate({ children, mutateFn }: IHelperGateContainer): JSX.Element {
  const selectedDevices = useContext(SelectedDevicesContext);

  const {
    session: { memberships, currentMembership },
    setCurrentMembership,
  } = useSession();
  const {
    query: { organisation },
  } = useRouter();

  // Handle setting the correct organisation
  useEffect(() => {
    let membership: SerializedMembership;

    if (organisation) {
      // if ?organisation is set, find the organisation and set it
      membership = memberships.find((membership) => membership?.organisation?.id === organisation);
    } else {
      // if there is no organisation selected yet, take the first one
      if (!currentMembership && memberships.length > 0) {
        membership = memberships[0];
      }
    }

    if (membership && membership.organisation.id !== currentMembership.organisation.id) {
      selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
      setCurrentMembership(membership);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [organisation, memberships]);

  useEffect(() => {
    if (mutateFn) {
      mutateFn();
    }
  }, [currentMembership]);

  return <>{children}</>;
}
