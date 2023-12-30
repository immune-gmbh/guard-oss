import Headline from 'components/elements/Headlines/Headline';
import EditOrganisationInfo from 'components/elements/OrganisationInfo/EditOrganisationInfo';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import React from 'react';

export default function DashboardUsersNewOrganisation(): JSX.Element {
  const onAfterSubmit = (info: any) => {
    if (info.errors) return;

    window.location.assign(NextJsRoutes.dashboardIndexPath);
  };

  return (
    <>
      <div className="shadow-xl bg-white p-6 mb-6">
        <div className="max-w-prose">
          <Headline className="mb-6">New Organisation</Headline>
          <EditOrganisationInfo create={true} afterSubmit={onAfterSubmit} />
        </div>
      </div>
    </>
  );
}

DashboardUsersNewOrganisation.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
