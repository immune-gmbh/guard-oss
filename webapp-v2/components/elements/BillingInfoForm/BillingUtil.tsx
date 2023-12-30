import getConfig from 'next/config';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { fetchAuthToken } from 'utils/session';

const { publicRuntimeConfig } = getConfig();

interface ISubscription {
  status: boolean;
  current_devices_amount: number;
  period_start: string;
  period_end: string;
  tax_rate: number;
}
interface IUser {
  name: string;
  email: string;
  invited: string;
  activation_state: string;
  address: string;
  organisations: Array<unknown>;
  subscriptions: Array<ISubscription>;
  errors: Array<unknown>;
}
interface ISubscriptionSetup {
  membershipId: string;
  setup_intent: {
    client_secret: string;
    status: 'requires_action' | 'succeeded';
  };
  error: string;
}
export const setupSubscription = async ({
  membershipId,
  name,
  password,
  paymentMethodId,
  street_and_number,
  city,
  postal_code,
  country,
  notifyInvoice,
  notifyDeviceUpdate,
  vatNumber,
}: {
  membershipId: string;
  name: string;
  password: string;
  paymentMethodId: string;
  street_and_number: string;
  city: string;
  postal_code: string;
  country: string;
  notifyInvoice: string;
  notifyDeviceUpdate: string;
  vatNumber: string;
}): Promise<ISubscriptionSetup> => {
  const data = await authenticatedJsonapiRequest(
    `${publicRuntimeConfig.hosts.authSrv}/v2/subscriptions/setup`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${fetchAuthToken()}`,
      },
      body: JSON.stringify({
        membership: {
          membershipId,
          notify_invoice: notifyInvoice,
          notify_device_update: notifyDeviceUpdate,
        },
        user: {
          name,
          password,
        },
        subscription: {
          paymentMethodId,
          street_and_number,
          city,
          postal_code,
          country,
          vat_number: vatNumber,
        },
      }),
    },
  );
  return data as ISubscriptionSetup;
};
export const updateUser = async ({
  street_and_number,
  city,
  postal_code,
  country,
  name,
  userId,
  password,
}: {
  street_and_number: string;
  city: string;
  postal_code: string;
  country: string;
  name: string;
  userId: string;
  password: string;
}): Promise<IUser> => {
  const data = (await authenticatedJsonapiRequest(
    `${publicRuntimeConfig.hosts.authSrv}/v2/users/${userId}`,
    {
      method: 'PATCH',
      body: JSON.stringify({
        user: {
          name,
          password,
          address_attributes: {
            street_and_number,
            city,
            postal_code,
            country,
          },
        },
      }),
    },
  )) as IUser;

  return data;
};
export const createSubscription = async ({
  membershipId,
  paymentMethodId,
}: {
  membershipId: string;
  paymentMethodId: string;
}): Promise<ISubscription> => {
  const data = await authenticatedJsonapiRequest(
    `${publicRuntimeConfig.hosts.authSrv}/v2/subscriptions`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${fetchAuthToken()}` },
      body: JSON.stringify({
        membership: { membershipId },
        subscription: { paymentMethodId },
      }),
    },
  );
  return data as ISubscription;
};
