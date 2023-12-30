import * as IssuesV1 from 'generated/issuesv1';

import { INCIDENTS } from './Incidents';
import { RISKS } from './Risks';

const EMPTY_ARGS = {
  files: [],
  components: [],
  processes: [],
  before: {},
  after: {},
  apps: [],
  devices: [],
  fprs: [],
  updates: [],
  values: [],
  variables: [],
  pcr: [],
  partitions: [],
};

export const KV: ({
  k,
  v,
}: {
  k: string | JSX.Element;
  v: string | JSX.Element;
}) => JSX.Element = ({ k, v }) => (
  <tr>
    <td>{k}:</td>
    <td className="px-4">{v || <i>Missing</i>}</td>
  </tr>
);

export const KVTable: React.FC<{ className?: string }> = ({ className, children }) => (
  <table className={className}>
    <tbody>{children}</tbody>
  </table>
);

export interface IRisk {
  group: 'vulnerability' | 'apt' | 'mitigation-failure' | 'update' | 'configuration-failure';
  score?: number;
  slug: React.ReactElement;
  description: React.ReactElement;
}

export interface IIncident {
  forensicsPre?: React.ReactElement;
  forensicsPost?: React.ReactElement;
  collapsible?: boolean;
  cta: React.ReactElement;
  slug: React.ReactElement;
  description: React.ReactElement;
}

type IIssue = IIncident | IRisk;

export const useIssue = (issue?: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml): IIssue => {
  if (!issue.id) return {} as IIssue;

  const issueId = issue.id.startsWith('brly/') ? 'brly/' : issue.id;

  if (issueId in RISKS) return RISKS[issueId](issue);
  if (issueId in INCIDENTS) return INCIDENTS[issueId](issue);
  return {} as IIssue;
};

export const useIssueFromId = (issueId: string): IIssue => {
  const issue = {
    id: issueId,
    args: EMPTY_ARGS,
  };

  return useIssue(issue as unknown as IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml);
};

export const getIssueType = (issueId: string): string => {
  if (!issueId) return '';
  const issueKey = issueId.startsWith('brly/') ? 'brly/' : issueId;
  return issueKey in RISKS ? 'risks' : 'incidents';
};
