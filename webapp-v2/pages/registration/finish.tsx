import { Elements } from '@stripe/react-stripe-js';
import { loadStripe } from '@stripe/stripe-js';
import BillingInfoForm from 'components/elements/BillingInfoForm/BillingInfoForm';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import { deserialise } from 'kitsu-core';
import getConfig from 'next/config';
import React from 'react';
import useSWR from 'swr';
import { handleLoadError } from 'utils/errorHandling';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { fetchAuthToken, fetchTempValue } from 'utils/session';
import { SerializedMembership } from 'utils/types';

interface ISession {
  id: string;
  invited: boolean;
  memberships: SerializedMembership[];
}

const fetcher = async (url: string, token: string): Promise<ISession> => {
  const data = (await authenticatedJsonapiRequest(url, {
    headers: { Authorization: `Bearer ${token}` },
  })) as { session: any; error?: any };
  return await deserialise(data.session).data;
};

export default function RegisterFinish(): JSX.Element {
  const { publicRuntimeConfig } = getConfig();

  const { data, error } = useSWR<ISession>(
    [`${publicRuntimeConfig.hosts.authSrv}/v2/session`, fetchAuthToken()],
    fetcher,
  );

  if (error) return handleLoadError('Session');
  if (!data) return <Spinner />;

  const membershipId = fetchTempValue('membershipId');
  const owner = data.memberships.find((mem) => mem.id === membershipId)?.role == 'owner';
  const stripePromise = loadStripe(publicRuntimeConfig.stripeApiKey);

  return (
    <DashboardLoggedOut>
      <Elements stripe={stripePromise}>
        <div className="flex flex-col items-center justify-center min-h-screen w-2/3">
          <h1 className="text-5xl my-10">You’re successfully registered!</h1>
          <BillingInfoForm
            invited={data.invited}
            userId={data.id}
            membershipId={data.invited && membershipId != 'undefined' ? membershipId : ''}
            owner={owner}
          />
        </div>
      </Elements>
    </DashboardLoggedOut>
  );
}
