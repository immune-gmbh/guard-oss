import { VERDICT_STATUS_STEPS } from 'components/elements/TrustChainBar/TrustChainIcon';
import * as IssuesV1 from 'generated/issuesv1';

import RiskCardEntry from './RiskCardEntry';

interface IRiskCard {
  issues: IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml[];
}

export default function RiskCard({ issues }: IRiskCard): JSX.Element {
  if (!issues || issues.length === 0)
    return (
      <section className="text-green-800 bg-white">
        <h3 className="font-bold text-2xl text-purple-500 mb-4">RISKS</h3>
        <div className="space-y-[30px]">
          <div className="bg-green-valid-light box-border p-8">
            <span className="flex gap-4 items-center">
              <svg
                width="30"
                height="30"
                viewBox="0 0 30 30"
                fill="none"
                xmlns="http://www.w3.org/2000/svg">
                <path d={VERDICT_STATUS_STEPS['firmware']} fill="#276749" />
              </svg>
              <h4 className="font-bold text-lg">Current Snapshot: No risks found</h4>
            </span>
            <div className="grid grid-cols-2 gap-2 items-start mt-4">
              <p className="text-lg col-span-2">
                This device and its components doesnt contain any risks we are aware of being used
                as part of a firmware supply-chain exploitation. The checks include detecting known
                malware inside the firmware components, vulnerabilities, mitigation, and
                configuration failures from the OEM/ODM/IBV and other suppliers.
              </p>
              <p className="text-lg col-span-2">
                This current snapshot of the system doesnt guarantee the firmware supply-chain
                integrity for its entire lifetime. Hackers can still use unknown vulnerabilities,
                malware, and attack paths to implant their code into the firmware. By firmware
                updates, new vulnerabilities, configuration and mitgation failures may also be
                introduced through OEM/ODM/IBV vendors.
              </p>
              <p className="text-lg col-span-2 font-bold">
                Our incident detection will help you to cover all other unknown attack types on this
                device.
              </p>
            </div>
          </div>
        </div>
      </section>
    );

  return (
    <section className="text-risk-dark-yellow bg-white" role="group">
      <h3 className="font-bold text-2xl text-purple-500 mb-4">RISKS</h3>
      <div className="space-y-[30px]">
        {issues.map((issue) => (
          <RiskCardEntry key={issue.id} issue={issue} />
        ))}
      </div>
    </section>
  );
}
