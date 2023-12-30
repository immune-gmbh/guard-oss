/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen } from '@testing-library/react';
import TrustChainBar from 'components/elements/TrustChainBar/TrustChainBar';
import common from 'locales/en/common.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { ApiSrv } from 'types/apiSrv';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

describe('TrustChainBar', () => {
  test('handles multiple incidents correctly', () => {
    const verdict: ApiSrv.Verdict = {
      type: 'verdict/2',
      result: 'vulnerable',
      supply_chain: 'vulnerable',
      configuration: 'trusted',
      firmware: 'vulnerable',
      bootloader: 'trusted',
      operating_system: 'trusted',
      endpoint_protection: 'unsupported',
    };

    render(
      <I18nProvider lang="en" namespaces={{ risks, incidents, common }}>
        <TrustChainBar verdict={verdict} unknown={false} />
      </I18nProvider>,
    );

    const supplyChain = screen.getByText('Supply Chain');
    expect(supplyChain).toHaveClass('text-red-cta');

    const firmware = screen.getByText('Host Firmware');
    expect(firmware).not.toHaveClass('text-red-cta');
  });
});
