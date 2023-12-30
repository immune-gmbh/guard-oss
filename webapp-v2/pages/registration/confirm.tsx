import Spinner from 'components/elements/Spinner/Spinner';
import NextJsRoutes from 'generated/NextJsRoutes';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useSession } from 'hooks/useSession';
import { useRouter } from 'next/dist/client/router';
import { useEffect } from 'react';
import { toast } from 'react-toastify';
import { useMutation } from 'utils/api';
import { setAuthToken } from 'utils/session';
import { SerializedSession, Response } from 'utils/types';

export default function RegisterConfirm(): JSX.Element {
  const router = useRouter();
  const { setSession, logout } = useSession();
  // info: get query: { activationToken: string; membershipToken: string }
  const {
    isLoading: activateIsLoading,
    isError: activateIsError,
    mutate: mutateUseActivate,
  } = useMutation(
    'POST',
    router.query.activationToken
      ? JsRoutesRails.activate_v2_user_path
      : JsRoutesRails.activate_v2_membership_path,
  );

  useEffect(() => {
    if ((router.query.activationToken || router.query.membershipToken) === undefined) {
      router.push(NextJsRoutes.registerPath);
    }
    if (activateIsLoading) {
      return;
    }
    if (activateIsError) {
      toast.error('Failed to activate user');
      logout();
    }

    (async () => {
      try {
        const token = router.query.activationToken || router.query.membershipToken;
        const sessionData = (await mutateUseActivate({ id: token })) as Response<SerializedSession>;

        if (!sessionData || sessionData.errors?.length > 0) {
          console.log('An error occured.');
          router.push(NextJsRoutes.registerPath);
        } else {
          setSession({ ...sessionData, currentMembership: sessionData?.memberships?.[0] });
          setAuthToken(sessionData.memberships?.[0].token);

          if (sessionData.user.hasSeenIntro) {
            router.push(NextJsRoutes.dashboardIndexPath);
          } else if (
            !sessionData.user.hasPassword &&
            !sessionData.user.authenticatedGithub &&
            !sessionData.user.authenticatedGoogle
          ) {
            router.push(NextJsRoutes.registrationSetPasswordPath);
          } else {
            router.push(NextJsRoutes.dashboardWelcomePath);
          }
        }
      } catch (e) {
        console.error(e);
      }
    })();
  }, []);

  return <Spinner />;
}
