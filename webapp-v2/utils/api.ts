import NextJsRoutes from 'generated/NextJsRoutes';
import getConfig from 'next/config';
import { useState } from 'react';
import { toast } from 'react-toastify';
import { convertToSnakeRecursive, convertToCamelRecursive } from 'utils/case';
import {
  deserialiseWithoutMeta,
  fetchSession,
  setAuthToken,
  ISession,
  storeSession,
  fetchGoAuthToken,
} from 'utils/session';

import { authenticatedJsonapiRequest } from './fetcher';
import { SerializedMembership, SerializedSession } from './types';

const { publicRuntimeConfig } = getConfig();

const HOST_URLS = {
  API: publicRuntimeConfig.hosts.apiSrv,
  AUTH: publicRuntimeConfig.hosts.authSrv,
} as const;
type HOST_TYPES = keyof typeof HOST_URLS;

export function deserialiseIfSet<T>(obj: T, names: string[]): T {
  names.forEach((name) => {
    if (obj[name] !== undefined && obj[name]) {
      obj[name] = deserialiseWithoutMeta(obj[name]);
    }
  });
  return obj;
}

export function pathWithQuery(path: string, query: Record<string, string>): string {
  const stringifiedQuery = Object.entries(query)
    .map(([key, val]) => `${key}=${val}`)
    .join('&');
  return stringifiedQuery ? `${path}?${stringifiedQuery}` : path;
}

export interface ApiResponse<T> {
  data: T | null;
  isLoading: boolean;
  isError: boolean;
}

export interface ApiMutationHook<T> {
  mutate: (body: Record<string, unknown>, urlFn?: (params: any) => string) => Promise<unknown>;
  data: T;
  isLoading: boolean;
  isError: boolean;
}

function authorizedAuthsrvFetch(
  url: string,
  options: Record<string, unknown>,
  host: HOST_TYPES = 'AUTH',
): Promise<Response> {
  const token = fetchGoAuthToken();
  const hostUrl = HOST_URLS[host];
  url = `${hostUrl}${url}`;
  return fetch(url, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/vnd.api+json',
      'Content-Type': 'application/vnd.api+json',
    },
  });
}

function RequestException(statusCode: number): void {
  this.statusCode = statusCode;
  this.name = 'RequestException';
}

export function useMutation<T>(
  method: 'DELETE' | 'PATCH' | 'POST',
  hookUrlFn: (params: any) => string,
  host: HOST_TYPES = 'AUTH',
  bodyFn = (body: any) => body,
): ApiMutationHook<T> {
  const [data, setData] = useState();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  return {
    mutate: async (body: Record<string, unknown>, urlFn: (params: any) => string) => {
      const url = (urlFn && urlFn(body)) || (hookUrlFn && hookUrlFn(body));
      if (!url) {
        throw new Error('Mutation URL not defined');
      }
      setLoading(true);
      body = bodyFn(body);
      return authorizedAuthsrvFetch(
        url,
        {
          method,
          body:
            Object.keys(body).length == 0
              ? undefined
              : JSON.stringify(convertToSnakeRecursive(body)),
        },
        host,
      )
        .then((res) => {
          if (!res.ok) {
            switch (res.status) {
              case 400:
              case 422:
                return res;
              default:
                throw new RequestException(res.status);
            }
          }
          return res;
        })
        .then((res) => Promise.all([res.status, res.json()]))
        .then(([code, json]) => {
          const data = deserialiseWithoutMeta(convertToCamelRecursive(json));
          data['code'] = code;
          setData(data as any);
          setLoading(false);
          setError(false);
          return data;
        })
        .catch((err) => {
          setError(true);

          if (err instanceof RequestException) {
            const session = fetchSession();
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-ignore
            processRequestError(url, err.statusCode, session);
          } else {
            return { errors: err };
          }
        });
    },
    data: data as T,
    isLoading: loading,
    isError: error,
  };
}

// To not spam the interface if one url constantly fails, only display toast once
const urlsWithError: Record<string, number> = {};

export async function refreshSession(
  setSession: (session: any) => void,
  membership: SerializedMembership,
): Promise<unknown> {
  const data = await authenticatedJsonapiRequest(
    `${publicRuntimeConfig.hosts.authSrv}/v2/session/refresh`,
  );
  const deserializedSession = deserialiseWithoutMeta(
    convertToCamelRecursive(data),
  ) as SerializedSession;
  const newSession = {
    ...deserializedSession,
    currentMembership: membership,
  } as ISession;
  await storeSession(newSession);
  setSession(newSession);
  setAuthToken(newSession.currentMembership?.token);
  return;
}

export const processRequestError = (
  url: string,
  statusCode: number,
  session = undefined,
): unknown => {
  if (statusCode == 401) {
    if (session && session?.currentMembership) {
      const body = JSON.parse(atob(session.currentMembership?.token.split('.')[1]));
      const exp = new Date(body.exp * 1000);
      if (exp < new Date(Date.now())) {
        toast.info('Your session expired.');
        return refreshSession(session, session.currentMembership);
      }
    }
    toast.error('Please log in again.');
    window.location.href = NextJsRoutes.logoutPath;
  }
  if (process.env.NODE_ENV == 'development') {
    const message = `Request to ${url} failed with code ${statusCode}`;
    if (urlsWithError[url] == undefined) {
      toast.error(message);
      urlsWithError[url] = 1;
    }
    console.error(message);
  } else {
    // TODO: Trigger Sentry
    if (window.location.pathname !== NextJsRoutes.logoutPath) {
      window.location.href = NextJsRoutes.logoutPath;
    }
  }
};
