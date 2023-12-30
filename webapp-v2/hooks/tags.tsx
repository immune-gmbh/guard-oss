import getConfig from 'next/config';
import useSWR, { KeyedMutator } from 'swr';
import useSWRImmutable from 'swr/immutable';
import { ApiSrv } from 'types/apiSrv';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { deserialiseWithoutMeta, fetchGoAuthToken } from 'utils/session';

const { publicRuntimeConfig } = getConfig();

const fetcher = async (url: string): Promise<any> => {
  const response = await authenticatedJsonapiRequest(url, {
    headers: {
      Authorization: `Bearer ${fetchGoAuthToken()}`,
    },
  });
  return { data: response };
};

interface ITagResponse {
  tag: ApiSrv.Tag;
  isLoading: boolean;
  isError: boolean;
}

const useTag = ({ id }: { id: string }): ITagResponse => {
  const { data, error } = useSWR(`${publicRuntimeConfig.hosts.apiSrv}/v2/tags/${id}`, fetcher);

  const tag = (
    data ? deserialiseWithoutMeta((data as Record<string, unknown>).data) : undefined
  ) as ApiSrv.Tag;

  return {
    tag,
    isLoading: !error && !data,
    isError: error,
  };
};

interface ITagsResponse {
  tags: ApiSrv.Tag[];
  isLoading: boolean;
  isError: boolean;
  mutate: KeyedMutator<any>;
}

const useSearchTags = ({ query }: { query: string }): ITagsResponse => {
  const { data, error, mutate } = useSWRImmutable(
    `${publicRuntimeConfig.hosts.apiSrv}/v2/tags?search=${query}`,
    fetcher,
  );

  const tags = (
    data ? deserialiseWithoutMeta((data as Record<string, unknown>).data) : []
  ) as ApiSrv.Tag[];

  return {
    tags,
    isLoading: !error && !data,
    isError: error,
    mutate,
  };
};

export { useTag, useSearchTags };
