import { JsRoutesRails } from 'generated/authsrvRoutes';
import { ISessionContext, SessionContext } from 'provider/SessionProvider';
import { useContext } from 'react';
import { ApiMutationHook, useMutation } from 'utils/api';

export const useSession = (): ISessionContext => {
  const useSession = useContext(SessionContext);
  return useSession;
};

export const useDeleteSession = (): ApiMutationHook<void> =>
  useMutation<void>('DELETE', JsRoutesRails.v2_session_path);
