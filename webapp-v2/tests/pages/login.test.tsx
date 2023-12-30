/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { useRouter } from 'next/router';
import Login from 'pages/login';
import { contextDefaultValues, SessionContext } from 'provider/SessionProvider';
import { serializedSession } from 'tests/mocks';
import * as UtilsApi from 'utils/api';

jest.mock('next/dist/client/router', () => ({
  useRouter: jest.fn(),
}));

const mockMutateUseLogin = jest.spyOn(UtilsApi, 'useMutation') as jest.Mock;

const getElements = () => {
  return {
    emailInput: screen.queryByPlaceholderText('Email Address'),
    passwordInput: screen.queryByPlaceholderText('Password'),
    registerButton: screen.queryByText('Register'),
    signInButton: screen.queryByText('Sign in with email'),
  };
};

describe('login', () => {
  const { location } = window;
  const mockLocationAssign = jest.fn();
  const mockRouterPush = jest.fn(() => Promise.resolve(true));

  mockMutateUseLogin.mockImplementation(() => ({
    mutate: () => serializedSession,
    data: serializedSession,
    isLoading: false,
    isError: false,
  }));

  beforeAll(() => {
    (useRouter as jest.Mock).mockReturnValue({
      query: { did: '123', token: null },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });

    delete window.location;
    window.location = { ...location, assign: mockLocationAssign };
  });

  afterAll(() => {
    window.location = location;
  });

  test('logs in with username and password', async () => {
    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <Login />
      </SessionContext.Provider>,
    );

    const { emailInput, passwordInput, signInButton } = getElements();

    fireEvent.change(emailInput, { target: { value: 'kai@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });
    fireEvent.click(signInButton);

    await waitFor(() => {
      expect(mockMutateUseLogin).toHaveBeenCalled();
      expect(mockLocationAssign).toHaveBeenCalledWith('https://xxxx.xxxx/dashboard/welcome');
    });
  });

  test('logs in with token', async () => {
    (useRouter as jest.Mock).mockReturnValueOnce({
      query: { did: '123', token: 'token' },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });

    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <Login />
      </SessionContext.Provider>,
    );

    expect(mockMutateUseLogin).toHaveBeenCalled();
    expect(mockLocationAssign).toHaveBeenCalledWith('https://xxxx.xxxx/dashboard/welcome');
  });

  test('redirects to register page', async () => {
    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <Login />
      </SessionContext.Provider>,
    );

    const { registerButton } = getElements();
    expect(registerButton).toBeInTheDocument();

    fireEvent.click(registerButton);
    expect(mockRouterPush).toHaveBeenCalledWith('/register', expect.anything(), expect.anything());
  });

  test('redirects to request password page', async () => {
    render(
      <SessionContext.Provider value={contextDefaultValues}>
        <Login />
      </SessionContext.Provider>,
    );

    const forgotPasswordLink = screen.queryByText(/Forgot password\? Reset it here.../i);
    expect(forgotPasswordLink).toBeInTheDocument();

    fireEvent.click(forgotPasswordLink);
    expect(mockRouterPush).toHaveBeenCalledWith(
      '/request_new_password',
      expect.anything(),
      expect.anything(),
    );
  });
});
