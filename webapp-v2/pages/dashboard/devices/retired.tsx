import DeviceBox from 'components/elements/DeviceBox/DeviceBox';
import Headline from 'components/elements/Headlines/Headline';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import { useDevices } from 'hooks/devices';
import Head from 'next/head';
import { handleLoadError } from 'utils/errorHandling';

export default function DevicesRetired(): JSX.Element {
  const {
    meta: { retiredDevices },
    isLoading,
    isError,
  } = useDevices();
  if (isLoading) return <Spinner />;
  if (isError) return handleLoadError('Retired Devices');

  return (
    <>
      <Head>
        <title>Immune Guard | Retired Devices</title>
      </Head>
      <div className="grid gap-x-8 transition-all overflow-hidden">
        <div>
          <Headline as="h1">Retired Devices</Headline>
          <div className="mt-12 space-y-4">
            {retiredDevices.length === 0 ? (
              <p className="text-3xl">Currently there is no device retired</p>
            ) : (
              retiredDevices?.map((device) => <DeviceBox key={device.id} device={device} />)
            )}
          </div>
        </div>
      </div>
    </>
  );
}

DevicesRetired.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout className="relative">{page}</DashboardLayout>;
};
