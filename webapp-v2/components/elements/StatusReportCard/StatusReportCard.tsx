import Button from 'components/elements/Button/Button';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import NextJsRoutes from 'generated/NextJsRoutes';
import * as IssuesV1 from 'generated/issuesv1';
import { useAppraisalsForDeviceConfig } from 'hooks/appraisals';
import { IGNORED_DEVICE_STATES, useDevicesConfig } from 'hooks/devices';
import { useAcceptAllChanges } from 'hooks/policies';
import { useRouter } from 'next/router';
import { useState } from 'react';
import { toast } from 'react-toastify';
import { ApiSrv } from 'types/apiSrv';

import StatusReportCardEntry from './StatusReportCardEntry';

interface IStatusReportCard {
  incidents: Array<IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml>;
  device: ApiSrv.Device;
}
export default function StatusReportCard({ incidents, device }: IStatusReportCard): JSX.Element {
  const { mutate: mutateDevices } = useDevicesConfig();
  const acceptAllChanges = useAcceptAllChanges();
  const { mutate: mutateDeviceAppraisals } = useAppraisalsForDeviceConfig();
  const [submitted, setSubmitted] = useState(false);
  const router = useRouter();

  const deviceIsMutable = !IGNORED_DEVICE_STATES.includes(device.state);

  if (incidents && incidents.length > 0)
    return (
      <section role="group">
        <h3 className="font-bold text-2xl text-purple-500 mb-4">INCIDENTS</h3>
        {incidents.map((incident) => (
          <StatusReportCardEntry key={incident.id} incident={incident} />
        ))}
        <div className="flex justify-end mt-8 border-purple-500 border-opacity-20 border-t pt-4">
          {deviceIsMutable && (
            <ConfirmModal
              headline="Do you really want to accept incidents as legitimate actions?"
              confirmLabel="Yes"
              onConfirm={() => {
                setSubmitted(true);

                acceptAllChanges
                  .mutate({
                    fwOverrides: incidents.map((incident) => incident.id),
                    device,
                  })
                  .then((response) => {
                    setSubmitted(false);

                    if ((response as any).errors) {
                      toast.error('Something went wrong while updating');
                    } else {
                      const dev = response as ApiSrv.Device;
                      mutateDevices();
                      mutateDeviceAppraisals(device.id);
                      toast.success(`Device "${dev.name}" was updated"`);
                    }

                    router.push({
                      pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
                      query: { did: device.id },
                    });
                  });
              }}
              TriggerComponent={(props) => (
                <Button data-qa="accept-changes" theme="SECONDARY" {...props}>
                  {submitted && <LoadingSpinner />}
                  Accept All
                </Button>
              )}></ConfirmModal>
          )}
        </div>
      </section>
    );
  else return <></>;
}
