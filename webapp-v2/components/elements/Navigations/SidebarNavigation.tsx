import { HomeIcon, KeyIcon, LogoutIcon, PlusIcon, UserIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import Headline from 'components/elements/Headlines/Headline';
import SubHeadline from 'components/elements/Headlines/SubHeadline';
import NavigationSection from 'components/elements/Navigations/NavigationSection';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useConfig } from 'hooks/config';
import { useSession } from 'hooks/useSession';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { SerializedMembership, SerializedUser } from 'utils/types';

import { INavigationLinkProps } from './NavigationLink';
import OrganisationSelect from './OrganisationSelect';

const DeviceNavigationItems = [
  {
    href: NextJsRoutes.dashboardDevicesIndexPath,
    label: 'Devices',
  },
  {
    href: NextJsRoutes.dashboardDevicesAddPath,
    label: 'Registration',
    Icon: PlusIcon,
  },
];
const ManagementNavigationItems = (membership: SerializedMembership): INavigationLinkProps[] => [
  {
    href: {
      pathname: NextJsRoutes.dashboardOrganisationsIdPath,
      query: { id: membership.organisation.id },
    },
    label: 'Organisation',
  },
  {
    href: NextJsRoutes.dashboardMembershipsPath,
    label: 'User',
  },
];

const DocsNavigationItems = [
  {
    href: NextJsRoutes.docsIncidentsPath,
    label: 'All Incidents',
  },
  {
    href: NextJsRoutes.docsRisksPath,
    label: 'All Risks',
  },
];

const SecurityScanItems = [
  {
    href: NextJsRoutes.dashboardIncidentsPath,
    label: 'Incidents',
  },
  {
    href: NextJsRoutes.dashboardRisksPath,
    label: 'Risks',
  },
];

const AccountItems = (role: SerializedUser['role']): INavigationLinkProps[] => [
  {
    href: NextJsRoutes.dashboardUsersAccountPath,
    label: 'Profile',
    Icon: UserIcon,
  },
  {
    href: NextJsRoutes.logoutPath,
    label: 'Logout',
    Icon: LogoutIcon,
  },
  {
    href: NextJsRoutes.adminUsersPath,
    label: 'Admin',
    Icon: KeyIcon,
    hide: role !== 'admin',
  },
];

const isActive = (currentPath: string, href: string): boolean => currentPath.includes(href);

export default function SidebarNavigation(): JSX.Element {
  const router = useRouter();
  const {
    session: {
      currentMembership,
      user: { role },
    },
  } = useSession();
  const { config } = useConfig();

  const minHeight = role == 'admin' ? 'min-h-[64rem]' : 'min-h-[62rem]';

  return (
    <aside
      className={`h-screen sticky flex flex-col justify-between py-8 px-6 top-0 bg-purple-500 text-white ${minHeight} w-[15.5rem]`}>
      <div>
        <Headline as="h1">
          <Image
            alt="Immune Logo"
            src="/immune.svg"
            quality="100"
            width={175}
            height={53}
            className="w-full"
          />
        </Headline>
        <nav className="mb-8">
          <Link href={NextJsRoutes.dashboardIndexPath} passHref>
            <a
              className={classNames('flex my-5 items-center', {
                'text-white text-opacity-60 ': !isActive(
                  router.pathname,
                  NextJsRoutes.dashboardIndexPath,
                ),
              })}>
              <HomeIcon height="15" className="pr-4" />
              <span className="text-xl font-bold antialiased transition-opacity hover:opacity-60">
                Dashboard
              </span>
            </a>
          </Link>
          <section className="grid gap-y-8">
            <NavigationSection headline="Fleet Management" menuItems={DeviceNavigationItems} />
            <NavigationSection headline="Security Analysis" menuItems={SecurityScanItems} />
            <NavigationSection
              headline="Settings"
              menuItems={ManagementNavigationItems(currentMembership)}
            />
            <NavigationSection headline="Documentation" menuItems={DocsNavigationItems} />
          </section>
        </nav>
      </div>
      <nav className="my-8">
        <SubHeadline>Organisation</SubHeadline>
        <OrganisationSelect />
        <NavigationSection menuItems={AccountItems(role)} className="mt-4" />
      </nav>
      <span className="absolute bottom-4 opacity-60 text-xs">Release: {config?.release}</span>
    </aside>
  );
}
