import TabNavigation from 'components/containers/TabNavigation/TabNavigation';
import Headline from 'components/elements/Headlines/Headline';
import UserEditNotifications from 'components/elements/UserEditNotifications/UserEditNotifications';
import UserEditProfile from 'components/elements/UserEditProfile/UserEditProfile';
// import UserEditTwoFactor from 'components/elements/UserEditTwoFactor/UserEditTwoFactor';
import UserOrganisationsList from 'components/elements/UserOrganisationsList/UserOrganisationsList';
import DashboardLayout from 'components/layouts/dashboard';

export default function DashboardUsersAccount(): JSX.Element {
  return (
    <>
      <Headline className="mb-6">Settings</Headline>
      <TabNavigation
        bodyClassNames="bg-gray-100 p-6"
        role="tabpanel"
        navigationPoints={[
          {
            title: 'Profile information',
            component: <UserEditProfile />,
          },
          {
            title: 'Organisations',
            component: <UserOrganisationsList />,
          },
          {
            title: 'Notifications',
            component: <UserEditNotifications />,
          },
          // {
          //   title: '2FA',
          //   component: <UserEditTwoFactor />,
          // },
        ]}
      />
    </>
  );
}

DashboardUsersAccount.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
