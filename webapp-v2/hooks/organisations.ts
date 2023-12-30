import { JsRoutesRails } from 'generated/authsrvRoutes';
import _ from 'lodash';
import useSWR from 'swr';
import { ApiMutationHook, ApiResponse, pathWithQuery, useMutation } from 'utils/api';
import { swrFetcher } from 'utils/fetcher';
import { SerializedOrganisation } from 'utils/types';

export const useOrganisations = (): ApiResponse<SerializedOrganisation[]> => {
  const { data, error } = useSWR(JsRoutesRails.v2_organisations_path(), swrFetcher);

  return {
    data: data as SerializedOrganisation[],
    isLoading: !error && !data,
    isError: error,
  };
};

interface IRouteParams {
  id: string;
  'include[]'?: 'users';
}

export const useOrganisation = (routeParams: IRouteParams): ApiResponse<SerializedOrganisation> => {
  const { data, error } = useSWR(
    routeParams['id']
      ? pathWithQuery(
          JsRoutesRails.v2_organisation_path({ id: routeParams['id'] }),
          _.pick(routeParams, 'include[]'),
        )
      : null,
    swrFetcher,
  );

  return {
    data: data as SerializedOrganisation,
    isLoading: !error && !data,
    isError: error,
  };
};

export const useDeleteOrganisation = (): ApiMutationHook<SerializedOrganisation> =>
  useMutation<SerializedOrganisation>('DELETE', JsRoutesRails.v2_organisation_path);
export const useUpdateOrganisation = (): ApiMutationHook<SerializedOrganisation> =>
  useMutation<SerializedOrganisation>('PATCH', JsRoutesRails.v2_organisation_path);
export const useCreateOrganisation = (): ApiMutationHook<SerializedOrganisation> =>
  useMutation<SerializedOrganisation>('POST', JsRoutesRails.v2_organisations_path);
