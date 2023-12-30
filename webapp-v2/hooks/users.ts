import { JsRoutesRails } from 'generated/authsrvRoutes';
import useSWR from 'swr';
import { ApiMutationHook, ApiResponse, deserialiseIfSet, useMutation } from 'utils/api';
import { swrFetcher } from 'utils/fetcher';
import { SerializedMembership, SerializedUser } from 'utils/types';

export const useUsers = (withOrganisations?: boolean): ApiResponse<SerializedUser[]> => {
  const { data, error } = useSWR(
    `${JsRoutesRails.v2_users_path()}${withOrganisations ? '?include=organisations' : ''}`,
    swrFetcher,
  );

  return {
    data: data as SerializedUser[],
    isLoading: !error && !data,
    isError: error,
  };
};

export const useCreateUser = (): ApiMutationHook<SerializedUser> =>
  useMutation<SerializedUser>('POST', JsRoutesRails.v2_users_path);
export const useDeleteUser = (): ApiMutationHook<SerializedUser> =>
  useMutation<SerializedUser>('DELETE', JsRoutesRails.v2_user_path);
export const useUpdateUser = (): ApiMutationHook<SerializedUser> =>
  useMutation<SerializedUser>('PATCH', JsRoutesRails.v2_user_path);
export const useChangePasswordUser = (): ApiMutationHook<{
  success: boolean;
  message: string | undefined;
}> =>
  useMutation<{ success: boolean; message: string | undefined }>(
    'PATCH',
    JsRoutesRails.change_password_v2_users_path,
  );
export const useResendActivationEmailUser = (): ApiMutationHook<{}> =>
  useMutation('POST', JsRoutesRails.resend_v2_user_path);

interface IInviteUserResponse {
  message: string;
  success: boolean;
  user: SerializedUser;
  membership: SerializedMembership;
}

export const useInviteUser = (): ApiMutationHook<IInviteUserResponse> => {
  const inviteUserMutation = useMutation<IInviteUserResponse>(
    'POST',
    JsRoutesRails.v2_memberships_path,
  );

  return {
    ...inviteUserMutation,
    mutate: (...args) =>
      inviteUserMutation
        .mutate(...args)
        .then((responseJson) => deserialiseIfSet(responseJson, ['user', 'membership'])),
  };
};
