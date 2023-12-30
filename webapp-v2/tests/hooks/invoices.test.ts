import { describe, expect, jest, test, beforeEach } from '@jest/globals';
import { useInvoices } from 'hooks/invoices';
import * as SWR from 'swr';
import { serializedSubscriptionInvoices } from 'tests/mocks';
import { fetcher, swrFetcher } from 'utils/fetcher';

const mockFetch = jest.spyOn(global, 'fetch');
const mockSWR = jest.spyOn(SWR, 'default');

function mockFetchImplementation(response: Array<any> | string, status = 200) {
  return mockFetch.mockImplementation(() =>
    Promise.resolve({
      json: () => Promise.resolve(response),
      text: () => Promise.resolve('some error message'),
      ok: status == 200,
      status,
    } as Response),
  );
}

describe('useInvoices hook', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  test('returns response in the correct format', async () => {
    mockFetchImplementation(serializedSubscriptionInvoices);
    const response = await swrFetcher('/invoices');
    mockSWR.mockImplementation(() => ({ data: response } as any));

    const { data } = useInvoices({ subscription_id: '123' });

    expect(data).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          id: expect.any(String),
        }),
      ]),
    );
  });

  test('sends an error, only fetches once', async () => {
    mockFetchImplementation('there was an error', 404);
    const response = await fetcher('/invoices');
    mockSWR.mockImplementation(() => ({ error: response } as any));

    const { isError } = useInvoices({ subscription_id: '123' });

    expect(mockFetch).toHaveBeenCalledTimes(1);
    expect(isError).toEqual('some error message');
  });
});
