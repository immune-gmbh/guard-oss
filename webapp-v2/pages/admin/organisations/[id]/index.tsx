import TabNavigation from 'components/containers/TabNavigation/TabNavigation';
import AlertBox from 'components/elements/AlertBox/AlertBox';
import Headline from 'components/elements/Headlines/Headline';
import EditOrganisationInfo from 'components/elements/OrganisationInfo/EditOrganisationInfo';
import OrganisationNotifications from 'components/elements/OrganisationNotifcations/OrganisationNotifcations';
import OrganisationPayment from 'components/elements/OrganisationPayment/OrganisationPayment';
import SubscriptionSummary from 'components/elements/OrganisationPayment/SubscriptionSummary';
import Spinner from 'components/elements/Spinner/Spinner';
import AdminLayout from 'components/layouts/admin';
import { useOrganisation } from 'hooks/organisations';
import { useSubscription } from 'hooks/subscriptions';
import { useRouter } from 'next/router';
import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { handleLoadError } from 'utils/errorHandling';

const editableOrganisationAttributes = [
  'name',
  'invoiceName',
  'vatNumber',
  'splunkEnabled',
  'splunkEventCollectorUrl',
  'splunkAuthenticationToken',
  'splunkAcceptAllServerCertificates',
  'syslogEnabled',
  'syslogHostnameOrAddress',
  'syslogUdpPort',
] as const;

const editableAddressAttributes = ['streetAndNumber', 'postalCode', 'city', 'country'] as const;

function AdminOrganisation(): JSX.Element {
  const router = useRouter();
  let { id } = router.query;
  const { edit } = router.query;
  id = typeof id == 'object' ? id[0] : (id as string);
  const { data: organisation, isLoading, isError } = useOrganisation({ id });
  const { data: subscription } = useSubscription(organisation?.subscription.id);
  const [isEditable, setIsEditable] = useState(edit == '1');
  const [isSubmitting] = useState(false);
  const methods = useForm({});

  useEffect(() => {
    if (organisation) {
      editableOrganisationAttributes.map((attr) => methods.setValue(attr, organisation[attr]));
      if (isEditable && !organisation.canEdit) {
        setIsEditable(false);
      }
      if (organisation.address) {
        editableAddressAttributes.map((attr) =>
          methods.setValue(`address.${attr}`, organisation.address[attr]),
        );
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [organisation]);

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
    <AdminLayout>
      {alerts}
      {isSubmitting && <Spinner />}
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
            navigationPoints={[
              {
                title: 'Regular Information',
                component: <EditOrganisationInfo admin={true} />,
              },
              {
                title: 'Billing',
                component: <OrganisationPayment organisationId={organisation.id} />,
              },
              {
                title: 'Notifications',
                component: <OrganisationNotifications />,
              },
            ]}
          />
        )}
      </div>
    </AdminLayout>
  );
}
export default AdminOrganisation;
