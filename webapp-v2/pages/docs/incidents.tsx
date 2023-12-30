import StatusReportCardEntry from 'components/elements/StatusReportCard/StatusReportCardEntry';
import DashboardLayout from 'components/layouts/dashboard';
import * as IssuesV1 from 'generated/issuesv1';
import { examples } from 'generated/issuesv1Examples';
import useTranslation from 'next-translate/useTranslation';
import Link from 'next/link';

//function unreachable(x: never): never;
//
//function unreachable(x: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml): never {
//  throw new Error(`unreachable: ${x.id}`);
//}
//
//interface IssueCardProps {
//  issue: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml;
//}
//
//function IssueCard({ issue }: IssueCardProps): JSX.Element {
//  switch (issue.id) {
//    case 'csme/no-update':
//      return <div>test</div>;
//  }
//  return unreachable(issue);
//}

const ASPECT_ORDER = [
  'supply-chain',
  'configuration',
  'firmware',
  'bootloader',
  'operating-system',
  'endpoint-protection',
];

function DocsIncidents(): JSX.Element {
  const { t } = useTranslation('common');
  const incidents = examples.filter(({ incident }) => incident);
  const byaspect: Array<[string, Array<IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml>]> =
    Object.entries(
      incidents.reduce(
        (acc, issue) => ({ ...acc, [issue.aspect]: [...(acc[issue.aspect] || []), issue] }),
        {},
      ),
    );
  byaspect.sort(([a], [b]) => ASPECT_ORDER.indexOf(a) - ASPECT_ORDER.indexOf(b));
  const nav = byaspect.map(([k, v], idx, ary) => (
    <span className="mr-2" key={`nav-${k}-${idx}`}>
      <Link href={`#${k}`} passHref={true}>
        <a className="underline">{t(k)}</a>
      </Link>{' '}
      ({v.length}){idx < ary.length - 1 ? ',' : ''}
    </span>
  ));
  const cards = byaspect.map(([k, v], idx) => (
    <div key={`card-${k}-${idx}`}>
      <h3 className="font-bold text-xl text-purple-500 mt-8 mb-4" id={k}>
        {t(k)}
      </h3>
      {v.map((incident, idx) => (
        <StatusReportCardEntry key={`${incident.id}-${idx}`} incident={incident} />
      ))}
    </div>
  ));

  return (
    <div>
      <h2 className="font-bold text-2xl text-purple-500 mb-4">INCIDENTS</h2>
      <div className="text-lg mb-3 text-purple-500 ">{nav}</div>
      {cards}
    </div>
  );
}

DocsIncidents.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};

export default DocsIncidents;
