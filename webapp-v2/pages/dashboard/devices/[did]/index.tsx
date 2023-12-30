import BreadcrumbNavigation from 'components/elements/BreadcrumbNavigation/BreadcrumbNavigation';
import DeviceName from 'components/elements/DeviceName/DeviceName';
import DeviceStatus from 'components/elements/DeviceStatus/DeviceStatus';
import DeviceNoAppraisal from 'components/elements/DeviceStatus/DeviceStatusNoAppraisal';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useDevice } from 'hooks/devices';
import { useSession } from 'hooks/useSession';
import { useRouter } from 'next/router';
import { handleLoadError } from 'utils/errorHandling';

export default function Device(): JSX.Element {
  const router = useRouter();
  const { did, organisation } = router.query;
  const {
    session: { memberships, currentMembership },
    setCurrentMembership,
  } = useSession();

  if (organisation) {
    const membership = memberships.find((membership) => membership.id === organisation);
    if (membership && membership !== currentMembership) {
      const href = NextJsRoutes.dashboardDevicesDidIndexPath.replace('[did]', did as string);
      setCurrentMembership(membership, href);
    }
  }

  const { device, isError, isLoading } = useDevice(did as string);
  const appraisal = device?.appraisals?.[0];

  if (isLoading) return <Spinner />;
  if (isError || !device) {
    return handleLoadError(`Device ${did}`);
  }

  return (
    <>
      <BreadcrumbNavigation
        crumbs={[
          {
            title: device.name,
            path: {
              pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
              query: { did: device?.id },
            },
            component: <DeviceName device={device} />,
          },
        ]}
      />
      {device && appraisal ? <DeviceStatus device={device} /> : <DeviceNoAppraisal />}
    </>
  );
}

Device.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout className="relative">{page}</DashboardLayout>;
};
