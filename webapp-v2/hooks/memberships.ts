import { JsRoutesRails } from 'generated/authsrvRoutes';
import useSWR from 'swr';
import { ApiMutationHook, ApiResponse, useMutation } from 'utils/api';
import { swrFetcher } from 'utils/fetcher';
import { SerializedMembership } from 'utils/types';

export const useMemberships = (): ApiResponse<SerializedMembership[]> => {
  const { data, error } = useSWR(JsRoutesRails.v2_memberships_path, swrFetcher);

  return {
    data: data as SerializedMembership[],
    isLoading: !error && !data,
    isError: error,
  };
};
export const useMembership = (id): ApiResponse<SerializedMembership> => {
  const { data, error } = useSWR(JsRoutesRails.v2_membership_path({ id }), swrFetcher);

  return {
    data: data as SerializedMembership,
    isLoading: !error && !data,
    isError: error,
  };
};

export const useDeleteMembership = (): ApiMutationHook<SerializedMembership> =>
  useMutation<SerializedMembership>('DELETE', JsRoutesRails.v2_membership_path);
export const useUpdateMembership = (): ApiMutationHook<SerializedMembership> =>
  useMutation<SerializedMembership>('PATCH', JsRoutesRails.v2_membership_path);
