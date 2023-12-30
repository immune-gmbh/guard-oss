import Actionbar from 'components/elements/Actionbar/Actionbar';
import Button from 'components/elements/Button/Button';
import Checkbox from 'components/elements/Checkbox/Checkbox';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import ControlledToggle from 'components/elements/Toggle/ControlledToggle';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useOrganisation, useUpdateOrganisation } from 'hooks/organisations';
import { useRouter } from 'next/router';
import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { mutate } from 'swr';
import { SerializedOrganisation, Response, Err } from 'utils/types';

type FormInputs = {
  splunkEnabled: boolean;
  splunkEventCollectorUrl: string;
  splunkAuthenticationToken: string;
  splunkAcceptAllServerCertificates: boolean;
  syslogEnabled: boolean;
  syslogHostnameOrAddress: string;
  syslogUdpPort: string;
};

function OrganisationNotifications(): JSX.Element {
  const updateOrganisation = useUpdateOrganisation();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const { data: organisation } = useOrganisation({ id: router.query.id as string });

  const {
    register,
    formState: { errors, isDirty },
    setError,
    control,
    watch,
    handleSubmit,
    reset,
  } = useForm<FormInputs>({
    defaultValues: {
      splunkEnabled: organisation.splunkEnabled,
      splunkEventCollectorUrl: organisation.splunkEventCollectorUrl,
      splunkAuthenticationToken: organisation.splunkAuthenticationToken,
      splunkAcceptAllServerCertificates: organisation.splunkAcceptAllServerCertificates,
      syslogEnabled: organisation.syslogEnabled,
      syslogHostnameOrAddress: organisation.syslogHostnameOrAddress,
      syslogUdpPort: organisation.syslogUdpPort,
    },
  });
  const watchSplunkEnabled = watch('splunkEnabled', organisation.splunkEnabled);
  const watchSyslogEnabled = watch('syslogEnabled', organisation.syslogEnabled);

  const editableOrganisationAttributes = [
    'splunkEnabled',
    'splunkEventCollectorUrl',
    'splunkAuthenticationToken',
    'splunkAcceptAllServerCertificates',
    'syslogEnabled',
    'syslogHostnameOrAddress',
    'syslogUdpPort',
  ] as const;

  const onSubmit = (data): void => {
    setIsLoading(true);
    updateOrganisation
      .mutate({ id: router.query.id, ...data })
      .then((organisation: Response<SerializedOrganisation>) => {
        let hasErrors = false;
        organisation?.errors?.map &&
          organisation.errors.map((err: Err) => {
            hasErrors = true;
            setError(err.id as typeof editableOrganisationAttributes[number], {
              message: err.title,
            });
          });
        if (!hasErrors) {
          mutate(JsRoutesRails.v2_organisation_path({ id: organisation.id }), organisation, false);
          toast.success('Organisation saved.');
        }
        setIsLoading(false);
      });
  };

  // reload the form when default values change
  useEffect(() => {
    reset(organisation, { keepIsSubmitted: true, keepSubmitCount: true });
  }, [organisation, reset]);

  return (
    <>
      <form id="OrganisationNotifications" onSubmit={handleSubmit(onSubmit)}>
        <div role="group" className="space-y-4 mb-8">
          <Headline size={4}>Splunk</Headline>
          <ControlledToggle control={control as any} name="splunkEnabled" qaLabel="splunk" />
          <p className="max-w-[70ch]">
            Send alerts about untrusted devices to your Splunk server. To enable Splunk alerts
            enable Splunk’s HTTP Event Collector (HEC). Then configure the URL of your Splunk
            instance and the HEC authentication token below.
          </p>
          <div className="border p-4 space-y-2">
            <Input
              disabled={!watchSplunkEnabled}
              label="Server"
              placeholder="URL of the HTTP Event Collector of your Splunk server"
              {...register(`splunkEventCollectorUrl`, {
                required: { value: watchSplunkEnabled, message: 'Field is required' },
              })}
              errors={[errors?.splunkEventCollectorUrl?.message]}
            />
            <Checkbox
              label="Accept all server certificates"
              {...register(`splunkAcceptAllServerCertificates`)}
              disabled={!watchSplunkEnabled}
            />
            <Input
              disabled={!watchSplunkEnabled}
              label="Token"
              placeholder="Authentication token for Splunk’s HTTP Event Collector."
              {...register(`splunkAuthenticationToken`, {
                required: { value: watchSplunkEnabled, message: 'Field is required' },
              })}
              errors={[errors?.splunkAuthenticationToken?.message]}
            />
          </div>
        </div>
        <div role="group" className="space-y-4">
          <Headline size={4}>Syslog</Headline>
          <ControlledToggle control={control as any} name="syslogEnabled" qaLabel="syslog" />
          <p>Send CEF formatted logs about untrusted devices to your Syslog server via UDP.</p>
          <div className="border p-4 space-y-2">
            <Input
              disabled={!watchSyslogEnabled}
              label="Server"
              placeholder="Hostname or IP address of the Syslog server"
              {...register(`syslogHostnameOrAddress`, {
                required: { value: watchSyslogEnabled, message: 'Field is required' },
              })}
              errors={[errors?.syslogHostnameOrAddress?.message]}
            />
            <Input
              disabled={!watchSyslogEnabled}
              label="UDP Port"
              placeholder="UDP Port"
              {...register(`syslogUdpPort`, {
                required: { value: watchSyslogEnabled, message: 'Field is required' },
              })}
              errors={[errors?.syslogUdpPort?.message]}
            />
          </div>
        </div>
        <Actionbar>
          <div />
          {isDirty && (
            <Button
              theme="SUCCESS"
              type="submit"
              form="OrganisationNotifications"
              disabled={isLoading}>
              {isLoading && <LoadingSpinner />}
              Save Organisation
            </Button>
          )}
        </Actionbar>
      </form>
    </>
  );
}
export default OrganisationNotifications;
