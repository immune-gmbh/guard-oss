interface ISimpleRequestCache {
  requestFingerprint: string;
  data: unknown;
}
let simpleCache: Record<string, ISimpleRequestCache> = {};

// Caches the result of the last calculation based for the specified key. Uses the Request fingerprint.
export function memoizeLastRequest<T>(key: string, fingerprint: string, calculateData: () => T): T {
  if (simpleCache[key] && simpleCache[key].requestFingerprint == fingerprint) {
    return simpleCache[key].data as T;
  } else {
    simpleCache[key] = {
      requestFingerprint: fingerprint,
      data: calculateData(),
    };
  }
  return simpleCache[key].data as T;
}

// @ts-ignore
export const addFingerprintMiddleware = (data: any) => ({
  data,
  // Random 7 char string
  requestFingerprint: (Math.random() + 1).toString(36).substring(7),
});
