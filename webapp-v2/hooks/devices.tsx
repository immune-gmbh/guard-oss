import _ from 'lodash';
import getConfig from 'next/config';
import { useMemo } from 'react';
import useSWR, { useSWRConfig } from 'swr';
import { ApiSrv } from 'types/apiSrv';
import { ApiMutationHook, useMutation } from 'utils/api';
import { addFingerprintMiddleware, memoizeLastRequest } from 'utils/caching';
import { authenticatedJsonapiRequest, retryPromiseFn } from 'utils/fetcher';
import { deserialiseWithoutMeta, fetchGoAuthToken } from 'utils/session';

const { publicRuntimeConfig } = getConfig();

const DEVICE_QUERY_PARAMS =
  '?include=appraisals&fields[appraisals]=id,appraised,verdict,report,issues';

export const USE_DEVICES_URL = publicRuntimeConfig.hosts.apiSrv + '/v2/devices?';

export const useDeviceUrl = (id: string) =>
  publicRuntimeConfig.hosts.apiSrv + `/v2/devices/${id}` + DEVICE_QUERY_PARAMS;

interface IDevicesResponse {
  devices: ApiSrv.Device[];
  meta: IDeviceMeta;
  isLoading: boolean;
  isError: boolean;
}

interface IDeviceResponse {
  device: ApiSrv.Device;
  isLoading: boolean;
  isError: boolean;
}

interface IDevicesConfig {
  mutate: () => Promise<unknown>;
}

export interface IDeviceMeta {
  retiredDevices: ApiSrv.Device[];
}

const fetcher = async (url: string): Promise<any> => {
  return retryPromiseFn(async () => {
    const response = await authenticatedJsonapiRequest(url, {
      headers: {
        Authorization: `Bearer ${fetchGoAuthToken()}`,
      },
    });
    return { data: addFingerprintMiddleware(response) };
  }, 3);
};

export const deviceToName = (device: ApiSrv.Device): string => {
  const smbios = device.appraisals[0].report?.values?.smbios;
  if (smbios) {
    const manufacturer = smbios.manufacturer || '';
    const product = smbios.product || '';

    return `${_.capitalize(manufacturer)} ${product}`;
  } else {
    return 'unknown';
  }
};

export const IGNORED_DEVICE_STATES: ApiSrv.Device['state'][] = ['resurrectable', 'retired'];

export const getUrlWithFilters = ({
  stateFilter,
  filterTags,
  issueFilter,
}: {
  stateFilter?: ApiSrv.Device['state'];
  filterTags?: Array<ApiSrv.Tag>;
  issueFilter?: string;
}) => {
  const params = [];
  if (stateFilter) params.push(['filter[state]', stateFilter]);
  if (issueFilter) params.push(['filter[issue]', issueFilter]);
  if (filterTags?.length) {
    const tagIds = filterTags.map((tag) => tag.id).join(',');
    params.push(['filter[tags]', tagIds]);
  }

  const queryParams = new URLSearchParams(params);
  return params.length ? `${USE_DEVICES_URL}${queryParams.toString()}` : USE_DEVICES_URL;
};

export const calculateDevicesMeta = (devices: ApiSrv.Device[]): IDeviceMeta => {
  const retiredDevices = devices.filter(
    (device: ApiSrv.Device) => device.state === 'resurrectable' && !device.replaced_by,
  );

  return { retiredDevices };
};

const useDevicesConfig = (): IDevicesConfig => {
  const { mutate: mutateSWR } = useSWRConfig();
  const mutate = (): Promise<unknown> => mutateSWR(USE_DEVICES_URL);

  return { mutate };
};

const useDevices = (): IDevicesResponse => {
  const { data, error } = useSWR(USE_DEVICES_URL, fetcher);

  const devices = useMemo(() => {
    return memoizeLastRequest<ApiSrv.Device[]>(
      'devices',
      data?.data?.requestFingerprint,
      () => (data ? deserialiseWithoutMeta(data.data.data) : []) as ApiSrv.Device[],
    );
  }, [data]);

  const meta = useMemo(() => {
    return memoizeLastRequest<IDeviceMeta>('devicesMeta', data?.data?.requestFingerprint, () =>
      calculateDevicesMeta(devices),
    );
  }, [data, devices]);

  return {
    devices,
    meta,
    isLoading: !error && !data,
    isError: error,
  };
};

const useDevice = (id: string): IDeviceResponse => {
  const { data, error } = useSWR(useDeviceUrl(id), fetcher);
  const device = deserialiseWithoutMeta(data?.data?.data) as ApiSrv.Device;

  return {
    device,
    isLoading: !error && !data,
    isError: error,
  };
};

export { useDevices, useDevice, useDevicesConfig };

export const useRetireDevice = (): ApiMutationHook<ApiSrv.Device> =>
  useMutation<ApiSrv.Device>(
    'PATCH',
    ({ id }) => `/v2/devices/${id}`,
    'API',
    ({ id }) => ({ data: [{ type: 'devices', id, attributes: { state: 'retired' } }] }),
  );
export const useRenameDevice = (): ApiMutationHook<ApiSrv.Device> =>
  useMutation<ApiSrv.Device>(
    'PATCH',
    ({ id }) => `/v2/devices/${id}`,
    'API',
    ({ id, name }) => ({ data: [{ type: 'devices', id, attributes: { name } }] }),
  );
export const useChangeTagsOfDevice = (): ApiMutationHook<ApiSrv.Device> =>
  useMutation<ApiSrv.Device>(
    'PATCH',
    ({ id }) => `/v2/devices/${id}`,
    'API',
    ({ id, tags }) => ({ data: [{ type: 'devices', id, attributes: { tags } }] }),
  );
export const useUpdateDevice = (): ApiMutationHook<Partial<ApiSrv.Device>> =>
  useMutation<ApiSrv.Device>(
    'PATCH',
    ({ id }) => `/v2/devices/${id}`,
    'API',
    ({ id, attributes }) => ({ data: [{ type: 'devices', id, attributes }] }),
  );

export const useResurrectDevice = (): ApiMutationHook<ApiSrv.Device> =>
  useMutation<ApiSrv.Device>('POST', ({ id }) => `/v2/devices/${id}/resurect`, 'API');

export const UNRETIREABLE_STATES = ['retired', 'resurrectable'];

export const exportForTesting = {
  fetcher,
};
