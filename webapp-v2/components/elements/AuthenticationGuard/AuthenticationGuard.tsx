import Spinner from 'components/elements/Spinner/Spinner';
import { useSession } from 'hooks/useSession';
import { useRouter } from 'next/router';
import React, { useEffect, useMemo } from 'react';
import { toast } from 'react-toastify';
import { ISession } from 'utils/session';

const AUTHENTICATED_DIRS = [
  {
    path: '/dashboard',
    guard: (session: ISession) => !!session,
  },
  {
    path: '/admin',
    guard: (session: ISession) => session && session?.user?.role == 'admin',
  },
];

function AuthenticationGuard({ children }): JSX.Element {
  const { session, isInitialized } = useSession();
  const { pathname, push } = useRouter();

  const isUnauthorized = useMemo(
    () =>
      isInitialized &&
      AUTHENTICATED_DIRS.some((dir) => pathname.indexOf(dir.path) !== -1 && !dir.guard(session)),

    [isInitialized, session, pathname],
  );

  useEffect(() => {
    if (isUnauthorized) {
      if (session) {
        push('/logout');
        toast.error('You are currently not authorized to access this feature.');
      } else {
        push('/login');
      }
    }
  }, [isUnauthorized]);

  if (!isInitialized || isUnauthorized) {
    return <Spinner />;
  }

  return children;
}
export default AuthenticationGuard;
