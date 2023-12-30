import { describe, expect, jest, test, beforeEach, afterEach } from '@jest/globals';
import { exportForTesting, useDevice, useDevices, USE_DEVICES_URL } from 'hooks/devices';
import React from 'react';
import * as SWR from 'swr';
import { serializedDeviceWithAppraisals, serializedDevices } from 'tests/mocks';

// to test with axios instead
// import axios from 'axios';
// jest.mock('axios')
// const axiosGet = axios.get as jest.MockedFunction<typeof axios.get>

const mockFetch = jest.spyOn(global, 'fetch');
const mockSWR = jest.spyOn(SWR, 'default');

function mockFetchImplementation(response: any, status = 200) {
  return mockFetch.mockImplementation(() =>
    Promise.resolve({
      json: () => Promise.resolve(response),
      ok: status == 200,
      status,
    } as Response),
  );
}

describe('fetcher', () => {
  const { fetcher } = exportForTesting;

  beforeEach(() => {
    mockFetch.mockReset();
    jest.spyOn(global.Math, 'random').mockReturnValue(0.123456789);
  });

  afterEach(() => {
    jest.spyOn(global.Math, 'random').mockRestore();
  });

  test('returns response in the correct format', async () => {
    mockFetchImplementation(serializedDevices);

    const response = await fetcher(USE_DEVICES_URL);

    expect(response).toEqual(
      expect.objectContaining({
        data: expect.objectContaining({
          data: expect.any(Array),
          requestFingerprint: 'xjyls',
        }),
      }),
    );
  });
});

describe('calculateDevicesMeta', () => {
  const { fetcher } = exportForTesting;

  beforeEach(() => {
    mockFetch.mockReset();
    jest.resetModules();
    jest.spyOn(React, 'useMemo').mockImplementation((calcValue) => calcValue());
  });

  test('get devices list from useDevices hook', async () => {
    mockFetchImplementation(serializedDevices);
    const response = await fetcher(USE_DEVICES_URL);
    mockSWR.mockImplementation(() => ({ data: response } as any));

    const { devices, meta } = useDevices();

    expect(devices).toEqual(response.data.data);
    expect(meta).toEqual(
      expect.objectContaining({
        retiredDevices: [],
      }),
    );
  });

  test('get one device by id from useDevice hook', async () => {
    mockFetchImplementation(serializedDeviceWithAppraisals);
    const response = await fetcher(USE_DEVICES_URL);
    mockSWR.mockImplementation(() => ({ data: response } as any));

    const { device } = useDevice('1770');

    expect(device).toEqual(
      expect.objectContaining({
        id: expect.any(String),
        hwid: expect.any(String),
        name: expect.any(String),
        policy: {
          endpointProtection: expect.any(String),
          intelTsc: expect.any(String),
        },
        state: expect.any(String),
        appraisals: expect.any(Array),
      }),
    );
  });
});
