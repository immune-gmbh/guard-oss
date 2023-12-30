import { ChevronRightIcon } from '@heroicons/react/solid';
import DeviceStatus from 'components/elements/Dashboard/DeviceStatus';
import RisksRanking from 'components/elements/Dashboard/RisksRanking';
import Headline from 'components/elements/Headlines/Headline';
import NoDevices from 'components/elements/NoDevices/NoDevices';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useDashboard } from 'hooks/dashboard';
import { useSession } from 'hooks/useSession';
import useTranslation from 'next-translate/useTranslation';
import Link from 'next/link';
import { useEffect } from 'react';

import { InstallDeviceComponent } from './devices/add';

export default function DashboardIndex(): JSX.Element {
  const { t } = useTranslation();
  const {
    session: { currentMembership },
    isInitialized,
    logout,
  } = useSession();
  const { data, isLoading, isError, loadData } = useDashboard();

  useEffect(() => {
    if (currentMembership) {
      loadData();
    }
  }, [currentMembership]);

  if (isLoading || !isInitialized) {
    return <Spinner />;
  }

  if (isError) {
    logout();
    return <Spinner />;
  }

  const hasActiveDevices =
    Object.values(data?.deviceStats || {}).reduce((prev, curr) => prev + curr, 0) > 0;

  const IncidentsBanner = (
    <SnippetBox
      className="group bg-slight-red p-6 pr-1.5 flex-row justify-between cursor-pointer hover:bg-red-critical"
      role="banner">
      <div>
        <Headline size={5} className="text-red-cta font-semibold">
          {t('dashboard:incidents.title')}
        </Headline>
        <p className="font-semibold text-purple-500 mt-1">
          {t('dashboard:incidents.description', {
            count: data?.incidents?.count,
            devices: data?.incidents?.devices,
          })}
        </p>
      </div>
      <ChevronRightIcon className="h-14 text-red-cta stroke-slight-red stroke-1 group-hover:stroke-red-critical" />
    </SnippetBox>
  );

  return (
    <>
      {!hasActiveDevices ? (
        <div className="space-y-16">
          <NoDevices />
          <InstallDeviceComponent currentMembership={currentMembership} />
        </div>
      ) : (
        <div className="grid grid-flow-row gap-4 h-full auto-rows-max mt-4">
          {data?.incidents?.count ? (
            <Link href={NextJsRoutes.dashboardIncidentsPath} passHref>
              <a className="h-fit">{IncidentsBanner}</a>
            </Link>
          ) : null}
          <div className="flex flex-wrap gap-4">
            <RisksRanking risks={data?.risks || []} />
            <DeviceStatus
              organisationId={currentMembership?.organisation?.id}
              deviceStats={data?.deviceStats}
            />
          </div>
        </div>
      )}
    </>
  );
}

DashboardIndex.getLayout = function GetLayout(page: React.ReactElement) {
  return <DashboardLayout className="max-w-full min-w-[28rem]">{page}</DashboardLayout>;
};
