import TabNavigation from 'components/containers/TabNavigation/TabNavigation';
import AlertBox from 'components/elements/AlertBox/AlertBox';
import Headline from 'components/elements/Headlines/Headline';
import EditOrganisationInfo from 'components/elements/OrganisationInfo/EditOrganisationInfo';
import OrganisationNotifications from 'components/elements/OrganisationNotifcations/OrganisationNotifcations';
import OrganisationPayment from 'components/elements/OrganisationPayment/OrganisationPayment';
import SubscriptionSummary from 'components/elements/OrganisationPayment/SubscriptionSummary';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import { useOrganisation } from 'hooks/organisations';
import { useSubscription } from 'hooks/subscriptions';
import { useSession } from 'hooks/useSession';
import Head from 'next/head';
import { useRouter } from 'next/router';
import React, { useEffect } from 'react';
import { handleLoadError } from 'utils/errorHandling';

export default function DashboardUsersOrganisation(): JSX.Element {
  const router = useRouter();
  let { id } = router.query;
  id = typeof id == 'object' ? id[0] : (id as string);
  const {
    session: { currentMembership },
  } = useSession();
  const { data: organisation, isLoading, isError } = useOrganisation({ id });
  const { data: subscription } = useSubscription(organisation?.subscription?.id);

  useEffect(() => {
    router.replace({ query: { ...router.query, id: currentMembership.organisation.id } });
  }, [currentMembership]);

  if (isError) return handleLoadError('Organisation');
  if (isLoading) return <Spinner />;

  const alerts = [
    !organisation?.subscription.id && (
      <AlertBox
        key="no-subscription"
        headline="No subscription found!"
        text="If you want to continue using our services, please update your credit card information"
      />
    ),
  ];

  return (
    <>
      <Head>
        <title>immune Guard | Organisation</title>
      </Head>
      {alerts}
      <div className="shadow-xl p-6 mb-6 bg-white">
        <div className="flex justify-between items-center mb-8">
          <Headline className="">{organisation && organisation.name}</Headline>
          {subscription && (
            <div className="float-right text-right">
              <SubscriptionSummary
                subscription={subscription}
                freeloader={organisation.freeloader}
              />
            </div>
          )}
        </div>
        {organisation && (
          <TabNavigation
            className="shadow-lg"
            bodyClassNames="bg-gray-100 p-6"
            activeTabIndex={router.query.tab === 'billing' ? 1 : 0}
            role="tabpanel"
            navigationPoints={[
              {
                title: 'Regular Information',
                component: <EditOrganisationInfo />,
              },
              {
                title: 'Billing',
                component: <OrganisationPayment organisationId={id} />,
              },
              {
                title: 'Notifications',
                component: <OrganisationNotifications />,
              },
            ]}
          />
        )}
      </div>
    </>
  );
}

DashboardUsersOrganisation.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
