/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, within } from '@testing-library/react';
import * as DevicesHook from 'hooks/devices';
import * as SessionHook from 'hooks/useSession';
import { useRouter } from 'next/router';
import Device from 'pages/dashboard/devices/[did]';
import { mockSessionContext, serializedDeviceWithAppraisals } from 'tests/mocks';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const mockUseRouter = useRouter as jest.Mock;
const mockUseDevice = jest.spyOn(DevicesHook, 'useDevice');
const mockUseSession = jest.spyOn(SessionHook, 'useSession') as jest.Mock;

describe('devices/[did]/index', () => {
  const deviceDid = serializedDeviceWithAppraisals.id;

  beforeEach(() => {
    mockUseRouter.mockReturnValue({
      pathname: `/dashboard/devices/${deviceDid}`,
      query: { did: deviceDid },
    });

    mockUseSession.mockImplementation(() => mockSessionContext);

    mockUseDevice.mockImplementation(() => ({
      device: serializedDeviceWithAppraisals,
      isError: false,
      isLoading: false,
    }));
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('renders component correctly', () => {
    render(<Device />);

    const [incidentsSection, risksSection] = screen.queryAllByRole('group');

    expect(screen.queryByText('fedora')).toBeInTheDocument();
    expect(screen.queryByText(/Device Integrity/i)).toBeInTheDocument();
    expect(within(incidentsSection).queryByText('INCIDENTS')).toBeInTheDocument();
    expect(within(risksSection).queryByText('RISKS')).toBeInTheDocument();
  });
});
