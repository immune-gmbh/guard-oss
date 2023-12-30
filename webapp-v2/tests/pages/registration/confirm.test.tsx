/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { render, waitFor } from '@testing-library/react';
import { useRouter } from 'next/dist/client/router';
import { contextDefaultValues, SessionContext } from 'provider/SessionProvider';
import RegisterConfirm from 'pages/registration/confirm';
import * as UtilsApi from 'utils/api';
import { toast } from 'react-toastify';
import { serializedSession } from 'tests/mocks';


jest.mock('next/dist/client/router', () => ({
  useRouter: jest.fn(),
}));

const mockMutateUseActivate = jest.spyOn(UtilsApi, 'useMutation') as jest.Mock;

const useActivateResponse = {
  mutate: () => Promise.resolve(serializedSession),
  data: serializedSession,
  isLoading: false,
  isError: false,
}

describe('register confirm', () => {
  const mockRouterPush = jest.fn(() => Promise.resolve(true));
  const mockToastError = jest.spyOn(toast, 'error');

  mockMutateUseActivate.mockImplementation(() => (useActivateResponse));

  afterEach(() => {
    mockRouterPush.mockReset();
  });

  test('redirects to register view if no token available', async () => {
    (useRouter as jest.Mock).mockReturnValueOnce({
      query: { did: '123', token: null },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });

    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <RegisterConfirm />
      </SessionContext.Provider>,
    );

    expect(mockRouterPush).toHaveBeenCalledWith('/register');
  });

  test('logs out user if api activate response has errors', async () => {
    const logout = jest.fn();

    (useRouter as jest.Mock).mockReturnValueOnce({
      query: { did: '123', activationToken: 'token' },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });

    mockMutateUseActivate.mockImplementationOnce(() => ({
      ...useActivateResponse,
      isError: true
    }));
  
    render(
      <SessionContext.Provider value={{...contextDefaultValues, logout }}>
        <RegisterConfirm />
      </SessionContext.Provider>,
    );

    expect(mockToastError).toHaveBeenCalled();
    expect(logout).toHaveBeenCalled();
  });

  test('redirects to welcome page', async () => {
    (useRouter as jest.Mock).mockReturnValueOnce({
      query: { did: '123', activationToken: 'token' },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });

    mockMutateUseActivate.mockImplementationOnce(() => (useActivateResponse));

    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <RegisterConfirm />
      </SessionContext.Provider>,
    );

    await waitFor(() => {
      expect(mockRouterPush).toHaveBeenCalledWith('/dashboard/welcome');
    });
  });
});
