import { Collapsible } from 'components/elements/Collapsible/Collapsible';
import { IIncident, useIssue } from 'components/elements/Issues/Index';
import { ISSUE_STATUS_STEPS } from 'components/elements/TrustChainBar/TrustChainIcon';
import * as IssuesV1 from 'generated/issuesv1';

interface IStatusReportCardEntry {
  incident: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml;
}

export default function StatusReportCardEntry({ incident }: IStatusReportCardEntry): JSX.Element {
  const incidentObj = useIssue(incident) as IIncident;
  const paths: Record<IssuesV1.Common['aspect'], string> = ISSUE_STATUS_STEPS;
  const path: string = paths[incident.aspect];

  return (
    <section className="p-8 bg-slight-red text-purple-500 mb-6" key={incident.id}>
      <div className="grid grid-cols-[30px,1fr] grid-rows-2 gap-x-4">
        <div className="col-start-1">
          <svg
            width="30"
            height="30"
            viewBox="0 0 30 30"
            fill="none"
            xmlns="http://www.w3.org/2000/svg">
            <path d={path} fill="#EB3A45" />
          </svg>
        </div>
        <h3 className="col-start-2 font-bold text-2xl text-red-cta">{incidentObj.slug}</h3>
      </div>
      <span className="col-start-2 row-start-2" role="contentinfo">
        {incidentObj.description}
      </span>
      {incidentObj?.forensicsPost && (
        <div className="col-start-2 mt-4">
          <h3 className="font-bold underline">Forensic analysis</h3>
          {incidentObj?.collapsible && (
            <>
              <Collapsible
                pre={incidentObj?.forensicsPre}
                enabled={incidentObj?.collapsible}
                indent={false}>
                {incidentObj?.forensicsPost}
              </Collapsible>
            </>
          )}
          {!incidentObj?.collapsible && incidentObj?.forensicsPost}
        </div>
      )}
      <div className="col-start-2 mt-4" role="complementary">
        <h3 className="font-bold underline">How do I resolve this incident?</h3>
        {incidentObj.cta}
      </div>
      <p className="font-bold mt-4">
        If this incident is a legitimate action, you can choose to accept it.
      </p>
    </section>
  );
}
