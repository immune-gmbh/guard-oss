import { ApiSrv } from 'types/apiSrv';

export enum DeviceActionType {
  TOGGLE = 'TOGGLE',
  CLEAR = 'CLEAR',
  INIT = 'INIT',
  SELECT_ALL = 'SELECT_ALL'
}

export interface DeviceAction {
  type: DeviceActionType;
  device?: ApiSrv.Device;
  devices?: Array<ApiSrv.Device>;
}

export function SelectedDevicesReducer(
  state: Array<ApiSrv.Device>,
  action: DeviceAction,
): Array<ApiSrv.Device> {
  switch (action.type) {
    case DeviceActionType.TOGGLE:
      if (!state.find((item) => item.id === action.device.id)) {
        return [...state, action.device];
      }
      return state.filter((device) => device.id !== action.device.id);
    case DeviceActionType.CLEAR:
      return [];
    case DeviceActionType.SELECT_ALL:
      return action.devices;
    case DeviceActionType.INIT:
      return (action.devices || []);
    default:
      return state;
  }
}
