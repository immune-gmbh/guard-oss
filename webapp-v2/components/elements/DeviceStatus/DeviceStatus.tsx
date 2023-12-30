import classNames from 'classnames';
import { Collapsible } from 'components/elements/Collapsible/Collapsible';
import ResurrectDeviceLink from 'components/elements/ResurrectDeviceLink/ResurrectDeviceLink';
import RetireDeviceButton from 'components/elements/RetireDeviceButton/RetireDeviceButton';
import RiskCard from 'components/elements/RiskCard/RiskCard';
import StatusReportCard from 'components/elements/StatusReportCard/StatusReportCard';
import Toggle from 'components/elements/Toggle/Toggle';
import TrustChainBar from 'components/elements/TrustChainBar/TrustChainBar';
import { VERDICT_STATUS_STEPS } from 'components/elements/TrustChainBar/TrustChainIcon';
import { format, formatDistance } from 'date-fns';
import { UNRETIREABLE_STATES, useDeviceUrl, useUpdateDevice } from 'hooks/devices';
import { toast } from 'react-toastify';
import { mutate } from 'swr';
import { ApiSrv } from 'types/apiSrv';
import { sortIPv4 } from 'utils/report';

interface IDeviceStatusProps {
  device: ApiSrv.Device;
}

export default function DeviceStatus({ device }: IDeviceStatusProps): JSX.Element {
  const updateDevice = useUpdateDevice();
  const deviceUrl = useDeviceUrl(device?.id);
  const appraisal = device.appraisals?.[0];

  // process verdict
  const firstErrorStep = Object.keys(VERDICT_STATUS_STEPS).find(
    (item) => appraisal.verdict[item] === 'vulnerable',
  ) as keyof ApiSrv.Verdict;
  const incidents = appraisal.issues?.issues.filter(({ incident }) => incident);
  const risks = appraisal.issues?.issues.filter(({ incident }) => !incident);

  // network addresses
  const macs: Array<string> = [];
  let ipv4s: Array<string> = [];
  let ipv6s: Array<string> = [];
  if (appraisal.report.values?.nics) {
    appraisal.report.values.nics.forEach((nic) => {
      if (nic.ipv4) ipv4s = ipv4s.concat(nic.ipv4).sort(sortIPv4);
      if (nic.ipv6) ipv6s = ipv6s.concat(nic.ipv6);
      if (nic.mac) macs.push(nic.mac);
    });
  }

  // last update
  const renderTimeValue = (): JSX.Element => {
    if (device.attestation_in_progress) {
      const date = new Date(device.attestation_in_progress);
      return (
        <div className="shadow-md col-start-1 col-span-2 p-4 mb-2">
          <span className="block uppercase font-bold text-sm tracking-widest">
            Attestation in Progress
          </span>
          <span>Last update: {format(date, '(MM/dd/yyyy, p)')}</span>
        </div>
      );
    }

    const lastReportDate = new Date(appraisal?.appraised);
    const formattedLastReportDate = formatDistance(lastReportDate, new Date(), { addSuffix: true });
    return (
      <div className="grid gap-x-4 col-start-1 col-span-2 grid-cols-[80px,1fr]">
        <label className="font-bold col-start-1">Timestamp</label>
        <span>
          {formattedLastReportDate}
          <br />
          {format(lastReportDate, '(MM/dd/yyyy, p)')}
        </span>
      </div>
    );
  };

  const renderPlatform = (show: boolean): JSX.Element => {
    if (show) {
      return (
        <>
          <label className="font-bold col-start-1">Platform</label>
          <span>{appraisal.report.values?.host?.name || '-'}</span>{' '}
        </>
      );
    }
  };

  return (
    <>
      <section className="text-purple-500 mt-8 mb-8 grid grid-cols-[auto,1fr,16em] gap-16 items-start">
        <div className="grid grid-cols-[80px,1fr] gap-4">
          {renderTimeValue()}
          <label className="font-bold grid-col col-start-1">Hostname</label>
          <span>{appraisal.report.values?.host?.hostname || '-'}</span>

          <label className="font-bold col-start-1">Serial #</label>
          <span>{appraisal.report.values?.smbios?.serial || '-'}</span>

          <label className="font-bold col-start-1">Product</label>
          <span>{appraisal.report.values?.smbios?.product || '-'}</span>

          <label className="font-bold col-start-1">Agent</label>
          <span>{appraisal.report.values?.agent_release || '-'}</span>

          {renderPlatform(!!device.attestation_in_progress)}
        </div>
        <div className="grid grid-cols-[80px,1fr] gap-4">
          <>
            {renderPlatform(!device.attestation_in_progress)}
            <label className="font-bold col-start-1">IPv4</label>
            <Collapsible enabled={ipv4s.length > 1}>
              <ul>
                {ipv4s.length
                  ? ipv4s.map((ipv4) => (
                      <li key={ipv4} className="text-ellipsis overflow-hidden">
                        {ipv4}
                      </li>
                    ))
                  : '-'}
              </ul>
            </Collapsible>
            <label className="font-bold col-start-1">IPv6</label>
            <Collapsible enabled={ipv6s.length > 1}>
              <ul>
                {ipv6s.length
                  ? ipv6s.map((ipv6) => (
                      <li key={ipv6} className="text-ellipsis overflow-hidden">
                        {ipv6}
                      </li>
                    ))
                  : '-'}
              </ul>
            </Collapsible>
            <label className="font-bold col-start-1">Ethernet</label>
            <Collapsible enabled={macs.length > 1}>
              <ul>
                {macs.length
                  ? macs.map((mac) => (
                      <li key={mac} className="text-ellipsis overflow-hidden">
                        {mac}
                      </li>
                    ))
                  : '-'}
              </ul>
            </Collapsible>
            {device.state === 'resurrectable' && (
              <>
                <label className="font-bold col-start-1">Retired</label>
                <span>
                  <ResurrectDeviceLink deviceId={device.id} />
                </span>
              </>
            )}
            {!UNRETIREABLE_STATES.includes(device.state) && (
              <>
                <label className="font-bold col-start-1">Action</label>
                <span>
                  <RetireDeviceButton deviceId={device.id} />
                </span>
              </>
            )}
          </>
        </div>
        <div className="space-y-4">
          <h3 className="font-bold text-xl text-purple-500 mb-8">Security Policy</h3>
          <div
            className={classNames({
              'opacity-20': device.policy.endpoint_protection === 'off',
            })}>
            <Toggle
              onChange={(checked) => {
                updateDevice
                  .mutate({
                    id: device.id,
                    attributes: {
                      policy: {
                        intel_tsc: device.policy.intel_tsc,
                        endpoint_protection: checked ? 'on' : 'if-present',
                      },
                    },
                  })
                  .then(() => {
                    mutate(deviceUrl).then((response: { data: any; error: any }) => {
                      if (response?.error || !response?.data) {
                        toast.error('Something went wrong');
                      } else {
                        toast.success('Device policy was successfully updated.');
                      }
                    });
                  });
              }}
              label="Require Endpoint Protection"
              labelColor="text-purple-500"
              checked={device.policy.endpoint_protection === 'on'}
              disabled={device.policy.endpoint_protection === 'off'}
            />
          </div>
          <div
            className={classNames({
              'opacity-20': device.policy.intel_tsc === 'off',
            })}>
            <Toggle
              onChange={(checked) => {
                updateDevice
                  .mutate({
                    id: device.id,
                    attributes: {
                      policy: {
                        intel_tsc: checked ? 'on' : 'if-present',
                        endpoint_protection: device.policy.endpoint_protection,
                      },
                    },
                  })
                  .then(() => {
                    mutate(deviceUrl).then((response: { data: any; error: any }) => {
                      if (response?.error || !response?.data) {
                        toast.error('Something went wrong');
                      } else {
                        toast.success('Device policy was successfully updated.');
                      }
                    });
                  });
              }}
              label="Require Intel Transparent Supply Chain"
              labelColor="text-purple-500"
              checked={device.policy.intel_tsc === 'on'}
              disabled={device.policy.intel_tsc === 'off'}
            />
          </div>
        </div>
      </section>
      <h3 className="font-bold text-2xl text-purple-500 mb-8 uppercase">Device Integrity</h3>
      <TrustChainBar verdict={appraisal.verdict} />
      <div className="flex flex-col space-y-8">
        {appraisal.verdict.result === 'vulnerable' && firstErrorStep && incidents && (
          <StatusReportCard incidents={incidents} device={device} />
        )}
        <RiskCard issues={risks} />
      </div>
    </>
  );
}
