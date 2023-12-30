/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import '@testing-library/jest-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import * as UsersHook from 'hooks/users';
import { useRouter } from 'next/router';
import SignIn from 'pages/register';
import { toast } from 'react-toastify';

jest.mock('next/dist/client/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseCreateUser = jest.spyOn(UsersHook, 'useCreateUser') as jest.Mock;

const getElements = () => {
  return {
    emailInput: screen.queryByPlaceholderText('Email Address'),
    passwordInput: screen.queryByPlaceholderText('Password'),
    nameInput: screen.queryByPlaceholderText('Name'),
    signUpButton: screen.queryByText('Sign up with email'),
  };
};

describe('register', () => {
  const mockRouterPush = jest.fn(() => Promise.resolve(true));
  const mockToastError = jest.spyOn(toast, 'error');

  mockUseCreateUser.mockImplementation(() => ({
    mutate: () => Promise.resolve({}),
  }));

  beforeAll(() => {
    (useRouter as jest.Mock).mockReturnValue({
      query: { did: '123', token: null },
      push: mockRouterPush,
      prefetch: () => Promise.resolve(true),
    });
  });

  afterEach(() => {
    mockRouterPush.mockReset();
  });

  test('allows registering if api response is successful', async () => {
    render(<SignIn />);

    const { emailInput, passwordInput, nameInput, signUpButton } = getElements();

    fireEvent.change(nameInput, { target: { value: 'Example' } });
    fireEvent.change(emailInput, { target: { value: 'mister@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });
    fireEvent.click(signUpButton);

    await waitFor(() => {
      expect(mockRouterPush).toHaveBeenCalledWith(
        '/registration/activate_email?email=mister@example.com',
      );
    });
  });

  test('throws error if api response fails', async () => {
    mockUseCreateUser.mockImplementationOnce(() => ({
      mutate: () => Promise.reject({}),
      data: {},
      isError: true,
    }));

    render(<SignIn />);

    const { emailInput, passwordInput, nameInput, signUpButton } = getElements();

    fireEvent.change(nameInput, { target: { value: 'Example' } });
    fireEvent.change(emailInput, { target: { value: 'mister@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });
    fireEvent.click(signUpButton);

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalled();
      expect(mockRouterPush).not.toHaveBeenCalled();
    });
  });
});
