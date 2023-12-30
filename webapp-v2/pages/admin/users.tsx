import { FilterIcon, SearchIcon } from '@heroicons/react/outline';
import UsersTable from 'components/elements/DomainTables/UsersTable';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import Spinner from 'components/elements/Spinner/Spinner';
import TagSelect from 'components/elements/TagSelect/TagSelect';
import AdminLayout from 'components/layouts/admin';
import { useUsers } from 'hooks/users';
import React, { useState } from 'react';

const ROLE_OPTIONS = { all: 'All', admin: 'Admin', user: 'User' } as const;

function AdminUsers(): JSX.Element {
  const { data, isLoading } = useUsers(true);
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<keyof typeof ROLE_OPTIONS>('all');

  return (
    <AdminLayout className="space-y-4">
      {isLoading && <Spinner />}
      <Headline>User Management</Headline>
      <div className="flex justify-between">
        <Input
          placeholder="User Name, Email Address"
          onChangeValue={setSearchText}
          wrapperClassName="w-96"
          icon={<SearchIcon />}
        />
      </div>
      <TagSelect
        selectedKey={roleFilter}
        options={ROLE_OPTIONS}
        onSelect={(s) => setRoleFilter(s as keyof typeof ROLE_OPTIONS)}
        icon={<FilterIcon />}
      />
      <div>
        {data && <UsersTable roleFilter={roleFilter} searchText={searchText} users={data} />}
      </div>
    </AdminLayout>
  );
}
export default AdminUsers;
