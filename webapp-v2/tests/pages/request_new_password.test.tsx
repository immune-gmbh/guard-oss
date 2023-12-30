/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test, beforeEach } from '@jest/globals';
import '@testing-library/jest-dom';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import DashboardIndex from 'pages/request_new_password';
import { toast } from 'react-toastify';

const mockFetch = jest.spyOn(global, 'fetch');

function mockFetchImplementation(response: Record<string, unknown> | string, status = 200) {
  return mockFetch.mockImplementation((input: unknown, init) => {
    expect(input).toMatch(/.+\/v2\/password_reset/);
    expect(init).toMatchSnapshot({
      body: JSON.stringify({ email: 'email@user.com' }),
      headers: {
        Accept: 'application/vnd.api+json',
        'Content-Type': 'application/vnd.api+json',
        Authorization: expect.stringMatching(/^Bearer .+$/),
      },
    });

    return Promise.resolve({
      json: () => Promise.resolve(response),
      ok: status == 200,
      status,
    } as Response);
  });
}

describe('send form to request password', () => {
  const mockToastSuccess = jest.spyOn(toast, 'success');
  const mockToastError = jest.spyOn(toast, 'error');

  beforeEach(() => {
    mockFetch.mockReset();
  });

  test('it sends form with successful response', async () => {
    mockFetchImplementation({ ok: true });

    render(<DashboardIndex />);

    const input = screen.getByPlaceholderText(/Email Address/i);
    const form = screen.getByTestId('form');

    expect(form).toBeInstanceOf(HTMLFormElement);

    fireEvent.change(input, { target: { value: 'email@user.com' } });

    await act(async () => {
      fireEvent.submit(form);
    });

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
      expect(mockToastSuccess).toHaveBeenCalled();
      expect(mockToastError).not.toHaveBeenCalled();
    });
  });
});
