import Headline from 'components/elements/Headlines/Headline';
import EditOrganisationInfo from 'components/elements/OrganisationInfo/EditOrganisationInfo';
import AdminLayout from 'components/layouts/admin';

function AdminNewOrganisation(): JSX.Element {
  return (
    <AdminLayout>
      <div className="shadow-xl bg-white p-6 mb-6">
        <div className="max-w-prose">
          <Headline className="mb-6">New Organisation</Headline>
          <EditOrganisationInfo create={true} admin={true} />
        </div>
      </div>
    </AdminLayout>
  );
}
export default AdminNewOrganisation;
