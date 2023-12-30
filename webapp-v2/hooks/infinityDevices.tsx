/**
 * useInfiniteLoading
 *
 * @author Luke Denton <luke@iamlukedenton.com>
 * @license MIT
 */
import { useCallback, useRef, useState } from 'react';
import { ApiSrv } from 'types/apiSrv';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { deserialiseWithoutMeta, fetchGoAuthToken } from 'utils/session';

import { USE_DEVICES_URL } from './devices';

const getDevices = ({ page }): Promise<any> => {
  return authenticatedJsonapiRequest(page, {
    headers: { Authorization: `Bearer ${fetchGoAuthToken()}` },
  });
};

export interface IInfinityDevices {
  devices: ApiSrv.Device[];
  hasNext: boolean;
  loadNext: () => void;
  isLoading: boolean;
  loadItems: (page: string) => Promise<any>;
}

const getIParamFromUrl = (url: string) => {
  const urlObj = new URL(url);
  return urlObj.searchParams.get('i');
};

/**
 * Handle infinite loading a list of items
 */
export const useInfiniteDeviceLoading = (): IInfinityDevices => {
  const [items, setItems] = useState<ApiSrv.Device[]>([]);
  const [hasNext, setHasNext] = useState(true);
  const isInFlight = useRef(false);

  const nextUrl = useRef('');

  const loadItems = useCallback(async (page: string): Promise<any> => {
    isInFlight.current = true;
    const data = await getDevices({ page });

    isInFlight.current = false;
    const devices = deserialiseWithoutMeta(data) as ApiSrv.Device[];
    if (!data.data || !data.links?.next) {
      setHasNext(false);
    } else {
      const iID = getIParamFromUrl(data?.links?.next);
      const urlNoI = new URL(page);
      urlNoI.searchParams.set('i', iID);

      nextUrl.current = urlNoI.href || USE_DEVICES_URL;
      setHasNext(true);
    }

    if (devices) {
      if (page.includes('i=')) setItems((prevDevices) => [...prevDevices, ...devices]);
      else setItems(devices);
    }
  }, []);

  const loadNext = (): void => {
    if (hasNext) {
      loadItems(nextUrl.current);
    }
  };

  return {
    devices: items,
    hasNext,
    loadNext,
    isLoading: isInFlight.current,
    loadItems,
  };
};
