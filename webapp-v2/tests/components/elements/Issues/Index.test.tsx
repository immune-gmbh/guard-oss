/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import { render, screen } from '@testing-library/react';
import { useIssueFromId } from 'components/elements/Issues/Index';
import binarly from 'locales/en/binarly.json';
import common from 'locales/en/common.json';
import incidents from 'locales/en/incidents.json';
import risks from 'locales/en/risks.json';
import I18nProvider from 'next-translate/I18nProvider';
import { issuesAspects } from 'utils/issues';

jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

const AuxComponent = ({ issueId }: { issueId: string }): JSX.Element => {
  return <div data-testid={issueId}>{useIssueFromId(issueId).description}</div>;
};

describe('useIssueFromId works with every issue', () => {
  test('it renders description with empty args', () => {
    Object.keys(issuesAspects).forEach((issueKey) => {

      render(
        <I18nProvider lang="en" namespaces={{ risks, incidents, common, binarly }}>
          <AuxComponent issueId={issueKey} />
        </I18nProvider>,
      );

      const content = screen.queryByTestId(issueKey);
      expect(content).toBeInTheDocument();
    });
  });
});
