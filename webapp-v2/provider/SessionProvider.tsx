import Spinner from 'components/elements/Spinner/Spinner';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useDeleteSession } from 'hooks/useSession';
import localforage from 'localforage';
import React, { createContext, useEffect, useState } from 'react';
import { refreshSession } from 'utils/api';
import { fetchSession, ISession, storeSession } from 'utils/session';
import { SerializedMembership } from 'utils/types';

export interface ISessionContext {
  session: ISession | null;
  isInitialized: boolean;
  setSession: (session: ISession) => void;
  logout: () => void;
  setCurrentMembership: (membership: SerializedMembership, href?: string) => void;
}

export const contextDefaultValues: ISessionContext = {
  session: null,
  isInitialized: false,
  setSession: () => {},
  logout: () => {},
  setCurrentMembership: () => {},
};

export const SessionContext = createContext<ISessionContext>(contextDefaultValues);

function SessionProvider({ children }: { children: React.ReactNode }): JSX.Element {
  const [session, setSession] = useState(contextDefaultValues.session);
  const [isInitialized, setIsInitialized] = useState(contextDefaultValues.isInitialized);
  const deleteSession = useDeleteSession();
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    async function initializeSession(): Promise<void> {
      if (!session) {
        const initSession = await fetchSession();
        setSession(initSession);
        setIsInitialized(true);
      }
    }
    initializeSession();
  }, []);

  const updateSession = async (session: ISession) => {
    await storeSession(session);
    setSession(session);
  };

  const logout = async () => {
    if (!session) {
      if (window.location.pathname === NextJsRoutes.logoutPath) {
        window.location.href = NextJsRoutes.loginPath;
      } else {
        return;
      }
    }
    setSession(null);

    await deleteSession.mutate({});
    await storeSession(null);
    localforage.clear();
    if (process.browser) {
      window.location.href = NextJsRoutes.loginPath;
    }
  };

  const setCurrentMembership = async (
    membership: SerializedMembership,
    href?: string,
  ): Promise<void> => {
    setLoading(true);
    await refreshSession(setSession, membership);
    setLoading(false);
  };

  return (
    <SessionContext.Provider
      value={{ session, logout, setCurrentMembership, isInitialized, setSession: updateSession }}>
      {loading && <Spinner />}
      {children}
    </SessionContext.Provider>
  );
}
export default SessionProvider;
