import { deserialise } from 'kitsu-core';
import localforage from 'localforage';
import _ from 'lodash';
import getConfig from 'next/config';
import Cookies from 'universal-cookie';
import { convertToCamelRecursive } from 'utils/case';
import { SerializedMembership, SerializedSession } from 'utils/types';

const cookies = new Cookies();
const COOKIE_MAX_AGE = 30 * 24 * 60 * 60;
const AUTH_TOKEN_COOKIE_KEY = 'auth-token';
const SESSION_STORAGE_KEY = 'session';

export function fetchGoAuthToken(): Cookies | string {
  const { publicRuntimeConfig } = getConfig();

  if (publicRuntimeConfig.developmentBearerToken) {
    return publicRuntimeConfig.developmentBearerToken;
  }
  return cookies.get(AUTH_TOKEN_COOKIE_KEY);
}

export function fetchAuthToken(): Cookies {
  return cookies.get(AUTH_TOKEN_COOKIE_KEY);
}
export function setAuthToken(token: string): void {
  return cookies.set(AUTH_TOKEN_COOKIE_KEY, token, { path: '/', maxAge: COOKIE_MAX_AGE });
}

type tempKeys = 'membershipId';
export function storeTempValue(key: tempKeys, value: string): void {
  return cookies.set(key, value, { path: '/', maxAge: COOKIE_MAX_AGE });
}
export function fetchTempValue(key: tempKeys): string {
  return cookies.get(key);
}
export function clearTempValue(key: tempKeys): void {
  return cookies.remove(key);
}

// shortcut deserialization to data, ignoring meta
function shortcutToDataAttribute(deserialisedData: any): any {
  const errors = deserialisedData?.errors || deserialisedData?.error;
  if (deserialisedData?.data) {
    deserialisedData = deserialisedData.data;
  }
  if (errors) {
    deserialisedData['errors'] = errors;
  }
  if (!_.isPlainObject(deserialisedData) && !_.isArrayLikeObject(deserialisedData)) {
    return deserialisedData;
  }
  Object.keys(deserialisedData).forEach((key) => {
    deserialisedData[key] = shortcutToDataAttribute(deserialisedData[key]);
  });
  return deserialisedData;
}
export function deserialiseWithoutMeta(response: unknown): unknown {
  const deserialisedData: unknown = deserialise(response);

  return shortcutToDataAttribute(deserialisedData);
}

export interface ISession extends SerializedSession {
  currentMembership?: SerializedMembership;
}

export async function programaticLogin(response: Record<string, unknown>): Promise<void> {
  try {
    // Cypress does something wierd
    const a = JSON.parse(JSON.stringify(response));
    const session = deserialiseWithoutMeta(convertToCamelRecursive(a)) as ISession;
    session.currentMembership = session.memberships[0];
    await storeSession(session);
  } catch (e) {
    console.error(e);
  }
}

export async function storeSession(session: ISession): Promise<ISession> {
  if (typeof window !== 'undefined') {
    try {
      return await localforage.setItem(SESSION_STORAGE_KEY, session);
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(err);
    }
  }
}
export async function fetchSession(): Promise<ISession> {
  if (typeof window !== 'undefined') {
    try {
      return await localforage.getItem(SESSION_STORAGE_KEY);
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(err);
    }
  }
}

export function initializeStorage(): void {
  localforage.config({
    driver: [localforage.INDEXEDDB, localforage.WEBSQL, localforage.LOCALSTORAGE],
    name: 'immuneGuard',
    version: 1.0,
    storeName: 'immune_guard_v2', // Should be alphanumeric, with underscores.
  });
}
