import localforage from 'localforage';
import React, { Context, createContext, useReducer, useEffect } from 'react';
import {
  DeviceAction,
  DeviceActionType,
  SelectedDevicesReducer,
} from 'reducer/SelectedDevicesReducer';
import { ApiSrv } from 'types/apiSrv';

const SELECTED_DEVICES_STORAGE_KEY = 'selected_devices';

interface SelectedDevicesContext {
  items: Array<ApiSrv.Device>;
  dispatch: React.Dispatch<DeviceAction>;
}

const SelectedDevicesContext: Context<SelectedDevicesContext> = createContext(
  {} as SelectedDevicesContext,
);

const SelectedDevicesProvider: React.FC = ({ children }) => {
  const [items, dispatch] = useReducer(SelectedDevicesReducer, []);

  useEffect(() => {
    (async () => {
      const devices =
        ((await localforage.getItem(SELECTED_DEVICES_STORAGE_KEY)) as Array<ApiSrv.Device>) || [];
      dispatch({ type: DeviceActionType.INIT, devices });
    })();
  }, []);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      (async function () {
        try {
          await localforage.setItem(SELECTED_DEVICES_STORAGE_KEY, items);
        } catch (err) {
          console.log(err);
        }
      })();
    }
  }, [items]);

  return (
    <SelectedDevicesContext.Provider
      value={{
        items,
        dispatch,
      }}>
      {children}
    </SelectedDevicesContext.Provider>
  );
};

export { SelectedDevicesProvider, SelectedDevicesContext };
