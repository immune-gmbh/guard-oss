import { compareDesc } from 'date-fns';
import getConfig from 'next/config';
import useSWR from 'swr';
import { ApiSrv } from 'types/apiSrv';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { deserialiseWithoutMeta, fetchGoAuthToken } from 'utils/session';
import { SerializedUser } from 'utils/types';

const fetcher = async (url: string): Promise<any> => {
  const response = await authenticatedJsonapiRequest(url, {
    headers: {
      Authorization: `Bearer ${fetchGoAuthToken()}`,
    },
  });
  return deserialiseWithoutMeta(response);
};

interface IChangesResponse {
  changes: Array<ApiSrv.Change>;
  isLoading: boolean;
  isError: boolean;
}

const useChanges = (users: SerializedUser[]): IChangesResponse => {
  const { publicRuntimeConfig } = getConfig();

  const { data, error } = useSWR(`${publicRuntimeConfig.hosts.apiSrv}/v2/changes`, fetcher);

  const changes = (data || []) as ApiSrv.Change[];

  changes
    .sort((a, b) => compareDesc(new Date(a.timestamp), new Date(b.timestamp)))
    .forEach((change: ApiSrv.Change) => {
      const match = change?.actor?.match(/^tag:immu.ne,2021:user\/([0-9a-fA-F-]+$)/);
      if (match) {
        const user = users.find((user) => user.id === match[1]);
        if (user) {
          change.actor = user.name;
        }
      }
    });

  return {
    changes,
    isLoading: !error && !data,
    isError: error,
  };
};

export { useChanges };
