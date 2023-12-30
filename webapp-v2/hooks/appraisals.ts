import getConfig from 'next/config';
import useSWR, { useSWRConfig } from 'swr';
import { ApiSrv } from 'types/apiSrv';

const { publicRuntimeConfig } = getConfig();

interface IAppraisalResponse {
  appraisal: ApiSrv.AppraisalData;
  isLoading: boolean;
  isError: boolean;
}

interface IAppraisalsConfig {
  mutate: (did: string) => Promise<unknown>;
}

const useAppraisalsForDeviceConfig = (): IAppraisalsConfig => {
  const { mutate: mutateSWR } = useSWRConfig();
  const mutate = (did: string): Promise<unknown> =>
    mutateSWR(`${publicRuntimeConfig.hosts.apiSrv}/v2/devices/${did}/appraisals`);

  return { mutate };
};

const useAppraisalForDevice = (did: string): IAppraisalResponse => {
  const { data, error } = useSWR(`/v2/devices/${did}/appraisals`);

  const appraisals = (data || []) as Array<ApiSrv.AppraisalData>;
  const appraisal = appraisals ? appraisals[0] : null;

  return {
    appraisal,
    isLoading: !error && !data,
    isError: error,
  };
};

export { useAppraisalForDevice, useAppraisalsForDeviceConfig };
