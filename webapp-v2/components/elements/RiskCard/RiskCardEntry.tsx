import { useIssue, IRisk } from 'components/elements/Issues/Index';
import { ISSUE_STATUS_STEPS } from 'components/elements/TrustChainBar/TrustChainIcon';
import * as IssuesV1 from 'generated/issuesv1';
import useTranslation from 'next-translate/useTranslation';

interface IRiskCardEntry {
  issue: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml;
}

export default function RiskCardEntry({ issue }: IRiskCardEntry): JSX.Element {
  const { t } = useTranslation();
  const { slug, description, score, group } = useIssue(issue) as IRisk;
  let cvss: string | undefined;
  const paths: Record<IssuesV1.Common['aspect'], string> = ISSUE_STATUS_STEPS;
  const path: string = paths[issue.aspect];

  if (score >= 9.0) {
    cvss = 'Critical';
  } else if (score >= 7.0) {
    cvss = 'High';
  } else if (score >= 4.0) {
    cvss = 'Medium';
  } else if (score >= 0.1) {
    cvss = 'Low';
  }

  return (
    <div className="bg-risk-yellow box-border p-8">
      <span className="flex gap-4 items-center">
        <svg
          width="30"
          height="30"
          viewBox="0 0 30 30"
          fill="none"
          xmlns="http://www.w3.org/2000/svg">
          <path d={path} fill="#805008" />
        </svg>
        <h4 className="font-bold text-lg">{slug}</h4>
      </span>
      <div className="mt-2">
        <span className="font-bold">
          Component:
          <span className="font-light ml-2">{t(`common:${issue.aspect}`)}</span>
        </span>
        <span className="font-bold mx-4">
          Type:
          <span className="font-light ml-2">{t(`common:${group}`)}</span>
        </span>
        {cvss && (
          <span className="font-bold">
            CVSS:
            <span className="font-light ml-2">
              {score} {cvss}
            </span>
          </span>
        )}
      </div>
      <div className="grid grid-cols-2 gap-2 items-start mt-4">
        <div className="text-lg col-span-2" role="contentinfo">
          {description}
        </div>
      </div>
    </div>
  );
}
