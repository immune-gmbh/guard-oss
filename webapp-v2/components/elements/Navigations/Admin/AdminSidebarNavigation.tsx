import { ChevronLeftIcon, HomeIcon, PlusIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import NavigationSection from 'components/elements/Navigations/NavigationSection';
import NextJsRoutes from 'generated/NextJsRoutes';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/router';
import React from 'react';

const isActive = (currentPath: string, href: string): boolean => currentPath.includes(href);

const AdminUsersNavigationItems = [
  {
    href: '/admin/users',
    label: 'Manage users',
  },
  {
    href: '/admin/invite',
    label: 'Add user',
    Icon: PlusIcon,
  },
];
const AdminOrganisationsNavigationItems = [
  {
    href: '/admin/organisations',
    label: 'Manage organisations',
  },
  {
    href: '/admin/organisations/new',
    label: 'Add organisation',
    Icon: PlusIcon,
  },
];

export default function AdminSidebarNavigation(): JSX.Element {
  const router = useRouter();

  return (
    <aside className="h-screen sticky py-8 px-6 top-0 bg-purple-500 text-white min-w-[250px] max-w-[350px]">
      <Link href={NextJsRoutes.dashboardIndexPath} passHref>
        <a className="mb-4 leading-[0px] flex flex-col items-center space-y-2 cursor-pointer">
          <Image
            alt="Immune Logo"
            src="/immune.svg"
            quality="100"
            width={175}
            height={53}
            className="w-full"
          />
        </a>
      </Link>
      <div className="flex">
        <Link href={NextJsRoutes.dashboardIndexPath} passHref>
          <a>
            <span className="p-2 rounded font-bold text-sm text-right flex text-gray-400 cursor-pointer hover:text-white">
              <ChevronLeftIcon className="w-5 h-5" /> Back
            </span>
          </a>
        </Link>
        <span className="bg-red-cta p-2 rounded font-bold text-sm text-right">ADMIN</span>
      </div>
      <nav>
        <Link href={NextJsRoutes.dashboardIndexPath} passHref>
          <a
            className={classNames('flex my-8 items-center', {
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
          <NavigationSection headline="Users" menuItems={AdminUsersNavigationItems} />
          <NavigationSection
            headline="Organisations"
            menuItems={AdminOrganisationsNavigationItems}
          />
        </section>
      </nav>
    </aside>
  );
}
