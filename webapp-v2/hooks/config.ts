import localforage from 'localforage';
import { useState, useEffect } from 'react';
import useSWR from 'swr';
import { swrFetcher } from 'utils/fetcher';

export const CONFIGURATION_STORAGE_KEY = 'configuration';

export interface IConfiguration {
  release?: string;
  agentUrls?: {
    ubuntu?: string;
    fedora?: string;
    generic?: string;
    windows?: string;
  };
}
interface IConfigurationResponse {
  data: {
    config: IConfiguration;
  };
  error: Error;
}

const INIT_CONFIG = {};

const useConfig = () => {
  const { data, error } = useSWR('/v2/appconfig', swrFetcher) as IConfigurationResponse;

  const [config, setConfig] = useState(data?.config);

  useEffect(() => {
    (async () => {
      const config =
        ((await localforage.getItem(CONFIGURATION_STORAGE_KEY)) as IConfiguration) || INIT_CONFIG;
      setConfig(config)
    })();
  }, []);

  useEffect(() => {
    if (typeof window !== 'undefined' && data) {
      (async function () {
        try {
          await localforage.setItem(CONFIGURATION_STORAGE_KEY, data?.config);
          setConfig(data?.config)
        } catch (err) {
          console.log('ConfigurationProvider', err);
        }
      })();
    }
  }, [data]);

  return {
    config,
    isLoading: !error && !data,
    isError: error
  }
}

export { useConfig };
