/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen, fireEvent } from '@testing-library/react';
import StatusReportCard from 'components/elements/StatusReportCard/StatusReportCard';
import { HttpsImmuneAppSchemasIssuesv1SchemaYaml } from 'generated/issuesv1';
import { useRouter } from 'next/router';
import { serializedDeviceWithAppraisals } from 'tests/mocks';
import { ApiSrv } from 'types/apiSrv';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

describe('send form to request password', () => {
  const device = serializedDeviceWithAppraisals;
  const incidents = device.appraisals[0].issues.issues.filter(({ incident }) => incident);

  test('it sends form with successful response', async () => {
    (useRouter as jest.Mock).mockReturnValue({
      query: { did: '123' },
    });

    render(
      <StatusReportCard
        incidents={incidents as HttpsImmuneAppSchemasIssuesv1SchemaYaml[]}
        device={device as ApiSrv.Device}
      />,
    );

    const button = screen.getByText(/Accept All/i);
    expect(button).toBeInstanceOf(HTMLButtonElement);

    fireEvent.click(button);

    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  test('it shows button only if device status is mutable', () => {
    (useRouter as jest.Mock).mockReturnValue({
      query: { did: '123' },
    });

    const nonMutableDevice = { ...(device as ApiSrv.Device), state: 'retired' };

    render(
      <StatusReportCard
        incidents={incidents as HttpsImmuneAppSchemasIssuesv1SchemaYaml[]}
        device={nonMutableDevice as ApiSrv.Device}
      />,
    );

    const button = screen.queryByText(/Accept All/i);
    expect(button).not.toBeInTheDocument();
  });
});
