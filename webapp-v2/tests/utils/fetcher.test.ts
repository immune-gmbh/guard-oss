import {describe, expect, jest, test, beforeEach} from '@jest/globals';
import { retryFetcher } from 'utils/fetcher';

const mockFetch = jest.spyOn(global, 'fetch');

function mockFetchImplementation(response: object | string, status = 200) {
  return mockFetch.mockImplementation(() =>
    Promise.resolve({
      json: () => Promise.resolve(response),
      ok: (status == 200),
      status: status,
    } as Response),
  );
}

describe('retryFetch method', () => {
  const responseBody = { data: ['obj1', 'obj2'] };

  beforeEach(() => {
    mockFetch.mockReset();
  });

  test('does not retry if it doesnt fail', async () => {
    mockFetchImplementation(responseBody)

    const responseWithRetry = await retryFetcher('/data', 3)

    expect(responseWithRetry).toEqual(responseBody.data)
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  test('does retry when it fails, but stops when successful', async () => {
    mockFetchImplementation(responseBody)

    mockFetch.mockImplementationOnce(() =>
      Promise.resolve({
        json: () => null,
        ok: false,
        status: 401
      } as Response),
    );

    await retryFetcher('/data', 5)
    
    expect(mockFetch).toHaveBeenCalledTimes(2)
  })

  test('does retry when it fails, stops when no more retries', async () => {
    mockFetchImplementation('there was an error', 404)

    await retryFetcher('/data', 4)
    expect(mockFetch).toHaveBeenCalledTimes(4)
  })
})
