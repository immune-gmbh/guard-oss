import { JsRoutesRails } from 'generated/authsrvRoutes';
import useSWR from 'swr';
import { ApiResponse } from 'utils/api';
import { swrFetcher } from 'utils/fetcher';
import { SerializedSubscription } from 'utils/types';

export const useSubscription = (id: string): ApiResponse<SerializedSubscription> => {
  const { data, error } = useSWR(id ? JsRoutesRails.v2_subscription_path({ id }) : null, swrFetcher, {
    errorRetryCount: 3,
  });

  return {
    data: data as SerializedSubscription,
    isLoading: !error && !data,
    isError: error,
  };
};
