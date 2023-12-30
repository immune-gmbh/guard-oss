import NextJsRoutes from 'generated/NextJsRoutes';
import getConfig from 'next/config';
import { toast } from 'react-toastify';
import { processRequestError } from 'utils/api';
import { convertToCamelRecursive } from 'utils/case';
import { deserialiseWithoutMeta, fetchAuthToken, fetchSession } from 'utils/session';

function exponentialDelay(retryNumber = 0): number {
  const delay = Math.pow(2, retryNumber) * 100;
  const randomSum = delay * 0.2 * Math.random(); // 0-20% of the delay
  return delay + randomSum;
}

export async function authenticatedJsonapiRequest<TResponse>(
  url: string,
  config: RequestInit = {},
): Promise<TResponse> {
  config.headers = Object.assign(
    {},
    {
      Authorization: `Bearer ${fetchAuthToken()}`,
      Accept: 'application/vnd.api+json',
      'Content-Type': 'application/vnd.api+json',
    },
    config.headers,
  );

  const response = await fetch(url, config);
  if (!response.ok) {
    throw response;
  }

  const data = await response.json();
  return data as TResponse;
}

export const fetcher = async (url: string, withRetries = false): Promise<unknown> => {
  const { publicRuntimeConfig } = getConfig();

  if (url.indexOf('http') == -1) {
    const host = publicRuntimeConfig.hosts.authSrv;
    url = `${host}${url}`;
  }
  let response: any;

  try {
    response = await authenticatedJsonapiRequest<any>(url);
  } catch (error) {
    if (error.status == '401') {
      let refreshedToken: any;

      try {
        refreshedToken = await processRequestError(url, error.status, await fetchSession());
      } catch (error) {
        window.location.href = NextJsRoutes.logoutPath;
        toast.error('You were logged out.');
      }
      try {
        response = await authenticatedJsonapiRequest<any>(url, {
          headers: { Authorization: `Bearer ${refreshedToken}` },
        });

        toast.success('Logged back in');
      } catch (error) {
        toast.error('You were logged out.');
      }
    } else if (withRetries) throw new Error(error.status);
    else return await error.text();
  }
  return deserialiseWithoutMeta(convertToCamelRecursive(response));
};

export const retryPromiseFn = async (fn: Function, retries = 1): Promise<unknown> => {
  let retryCount = 0;
  let currentError: unknown;

  while (retries > 0) {
    try {
      return await fn();
    } catch (error) {
      currentError = error;

      retryCount += 1;
      retries -= 1;

      await new Promise((resolve) => setTimeout(resolve, exponentialDelay(retryCount)));
    }
  }
  return currentError;
};

export const retryFetcher = async (url: string, retries = 1): Promise<unknown> => {
  return await retryPromiseFn(() => fetcher(url, true), retries);
};

export const swrFetcher = async (url: string) => retryFetcher(url, 5);

export default fetcher;
