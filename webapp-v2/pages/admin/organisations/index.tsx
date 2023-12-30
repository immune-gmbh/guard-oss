import { SearchIcon } from '@heroicons/react/outline';
import OrganisationsTable from 'components/elements/DomainTables/OrganisationsTable';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import Spinner from 'components/elements/Spinner/Spinner';
import AdminLayout from 'components/layouts/admin';
import { useOrganisations } from 'hooks/organisations';
import React, { useState } from 'react';

function AdminOrganisations(): JSX.Element {
  const { data, isLoading } = useOrganisations();
  const [searchText, setSearchText] = useState('');

  return (
    <AdminLayout className="space-y-4">
      {isLoading && <Spinner />}
      <Headline>Organisation Management</Headline>
      <div className="flex justify-between">
        <Input
          placeholder="User Name, Email Address"
          onChangeValue={setSearchText}
          wrapperClassName="w-96"
          icon={<SearchIcon />}
        />
      </div>
      <div>{data && <OrganisationsTable organisations={data} searchText={searchText} />}</div>
    </AdminLayout>
  );
}

export default AdminOrganisations;
