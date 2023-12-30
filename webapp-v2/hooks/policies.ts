import { ApiSrv } from 'types/apiSrv';
import { ApiMutationHook, useMutation } from 'utils/api';

interface IAcceptAllChanges {
  fwOverrides: Array<string>;
  device: ApiSrv.Device;
}
const useAcceptAllChanges = (): ApiMutationHook<ApiSrv.Device> =>
  useMutation<ApiSrv.Device>(
    'POST',
    ({ device }) => `/v2/devices/${device.id}/override`,
    'API',
    ({ fwOverrides }: IAcceptAllChanges) => fwOverrides,
  );

export { useAcceptAllChanges };
