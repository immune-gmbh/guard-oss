import RiskCardEntry from 'components/elements/RiskCard/RiskCardEntry';
import DashboardLayout from 'components/layouts/dashboard';
import * as IssuesV1 from 'generated/issuesv1';
import { examples } from 'generated/issuesv1Examples';
import useTranslation from 'next-translate/useTranslation';
import Link from 'next/link';

const ASPECT_ORDER = [
  'supply-chain',
  'configuration',
  'firmware',
  'bootloader',
  'operating-system',
  'endpoint-protection',
];

function DocsRisks(): JSX.Element {
  const { t } = useTranslation('common');
  const risks = examples.filter(({ incident }) => !incident);
  const byaspect: Array<[string, Array<IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml>]> =
    Object.entries(
      risks.reduce(
        (acc, issue) => ({ ...acc, [issue.aspect]: [...(acc[issue.aspect] || []), issue] }),
        {},
      ),
    );
  byaspect.sort(([a], [b]) => ASPECT_ORDER.indexOf(a) - ASPECT_ORDER.indexOf(b));
  const nav = byaspect.map(([k, v], idx, ary) => (
    <span className="mr-2" key={k}>
      <Link href={`#${k}`} passHref={true}>
        <a className="underline">{t(k)}</a>
      </Link>{' '}
      ({v.length}){idx < ary.length - 1 ? ',' : ''}
    </span>
  ));
  const cards = byaspect.map(([k, v]) => (
    <>
      <h3 className="font-bold text-xl text-purple-500 mt-8 mb-4" id={k}>
        {t(k)}
      </h3>
      {v.map((incident) => (
        <RiskCardEntry key={incident.id} issue={incident} />
      ))}
    </>
  ));

  return (
    <section className="text-risk-dark-yellow bg-white">
      <h2 className="font-bold text-2xl text-purple-500 mb-4">RISKS</h2>
      <div className="text-lg mb-3 text-purple-500 ">{nav}</div>
      <div className="space-y-[30px]">{cards}</div>
    </section>
  );
}

DocsRisks.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};

export default DocsRisks;
