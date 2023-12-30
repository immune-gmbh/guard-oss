import AuthenticationGuard from 'components/elements/AuthenticationGuard/AuthenticationGuard';
import Edgebanner from 'components/elements/Edgebanner/Edgebanner';
import NextJsRoutes from 'generated/NextJsRoutes';
import { NextPage } from 'next';
import App, { AppContext, AppProps } from 'next/app';
import getConfig from 'next/config';
import Head from 'next/head';
import SessionProvider from 'provider/SessionProvider';
import React, { ReactNode } from 'react';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import 'styles/global.css';
import { SWRConfig } from 'swr';
import { authenticatedJsonapiRequest, retryPromiseFn } from 'utils/fetcher';
import { deserialiseWithoutMeta, fetchGoAuthToken, initializeStorage } from 'utils/session';

if (process.env.NEXT_PUBLIC_MOCK) {
  import('../mocks');
}

const { publicRuntimeConfig } = getConfig();

if (typeof window !== 'undefined') {
  initializeStorage();
}

type NextPageWithLayout = NextPage & {
  getLayout?: (page: React.ReactElement) => ReactNode;
};

type AppPropsWithLayout = AppProps & {
  Component: NextPageWithLayout;
};

function ImmuneGuard({ Component, pageProps }: AppPropsWithLayout): JSX.Element {
  const getLayout = Component.getLayout ?? ((page) => page);

  return (
    <SWRConfig
      value={{
        refreshInterval: 60000,
        dedupingInterval: 50000,
        onErrorRetry: (error, key, _config, revalidate, { retryCount }) => {
          // Never retry for a specific key.
          if (key === '/v2/session') return;

          // Never retry on 404.
          if (error.status === 404 || retryCount >= 3) {
            // @TODO: inform sentry about this issue
            if (process.env.NODE_ENV != 'development') {
              return (window.location.href = NextJsRoutes.loginPath);
            }
          }

          // Retry after 5 seconds.
          setTimeout(() => revalidate({ retryCount }), 5000);
        },
        fetcher: (resource) =>
          retryPromiseFn(async () => {
            const data = await authenticatedJsonapiRequest(
              publicRuntimeConfig.hosts.apiSrv + resource,
              {
                headers: {
                  Authorization: `Bearer ${fetchGoAuthToken()}`,
                },
              },
            );
            return deserialiseWithoutMeta(data);
          }, 5),
      }}>
      <Head>
        <link rel="shortcut icon" href="/favicon.png" type="image/png" />
        <title>immune Guard{publicRuntimeConfig.isEdge ? ' (Edge)' : ''}</title>
      </Head>
      <SessionProvider>
        <AuthenticationGuard>
          {publicRuntimeConfig.isEdge && <Edgebanner />}
          {getLayout(<Component {...pageProps} />)}
        </AuthenticationGuard>
      </SessionProvider>
      <ToastContainer />
    </SWRConfig>
  );
}

ImmuneGuard.getInitialProps = async (appContext: AppContext) => {
  App.getInitialProps(appContext);
  return {};
};

export default ImmuneGuard;
