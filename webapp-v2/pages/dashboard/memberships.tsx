import { FilterIcon, SearchIcon } from '@heroicons/react/solid';
import ImmuneLink from 'components/elements/Button/Link';
import MembershipsTable from 'components/elements/DomainTables/MembershipsTable';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import TagSelect from 'components/elements/TagSelect/TagSelect';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useOrganisation } from 'hooks/organisations';
import { useSession } from 'hooks/useSession';
import Head from 'next/head';
import Link from 'next/link';
import React, { useState } from 'react';

const ROLE_OPTIONS = { all: 'All', owner: 'Owner', user: 'User' } as const;

export default function DashboardUsers(): JSX.Element {
  const {
    session: {
      currentMembership: {
        organisation: { id: organisationId },
      },
    },
  } = useSession();
  const { data } = useOrganisation({ id: organisationId, 'include[]': 'users' });
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<keyof typeof ROLE_OPTIONS>('all');

  return (
    <>
      <Head>
        <title>immune Guard | User Management</title>
      </Head>
      <Headline>User Management</Headline>
      <div className="flex justify-between">
        <Input
          placeholder="User Name, Email Address"
          onChangeValue={setSearchText}
          wrapperClassName="w-96"
          icon={<SearchIcon />}
        />
        <ImmuneLink theme="MAIN" href={NextJsRoutes.dashboardUsersInvitePath}>
          Invite User
        </ImmuneLink>
      </div>
      <TagSelect
        selectedKey={roleFilter}
        options={ROLE_OPTIONS}
        onSelect={(s) => setRoleFilter(s as keyof typeof ROLE_OPTIONS)}
        icon={<FilterIcon />}
      />
      <div>
        {data && (
          <MembershipsTable
            editable={data.canEdit}
            roleFilter={roleFilter}
            searchText={searchText}
            memberships={data.memberships}
            adminView={false}
          />
        )}
        <div className=" px-4 py-2 bg-gray-50 border-t-2">
          <Link href={NextJsRoutes.dashboardUsersInvitePath} passHref>
            <a>
              <span className="underline font-bold cursor-pointer">Invite User</span>
            </a>
          </Link>
        </div>
      </div>
    </>
  );
}

DashboardUsers.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout className="space-y-4">{page}</DashboardLayout>;
};
