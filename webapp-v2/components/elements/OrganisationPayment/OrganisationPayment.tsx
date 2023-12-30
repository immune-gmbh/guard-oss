import { Elements } from '@stripe/react-stripe-js';
import { loadStripe } from '@stripe/stripe-js';
import Headline from 'components/elements/Headlines/Headline';
import SubscriptionInfo from 'components/elements/OrganisationPayment/SubscriptionInfo';
import { useOrganisation } from 'hooks/organisations';
import { useSubscription } from 'hooks/subscriptions';
import getConfig from 'next/config';
import React from 'react';

import PaymentForm from './PaymentForm';

interface IOrganisationPaymentProps {
  organisationId: string;
}

const OrganisationPayment: React.FC<IOrganisationPaymentProps> = ({ organisationId }) => {
  const { publicRuntimeConfig } = getConfig();

  const { data: organisation } = useOrganisation({ id: organisationId });
  const { data: subscription } = useSubscription(organisation?.subscription.id);

  if (organisation?.freeloader) {
    return (
      <div className="space-y-3 ">
        <p>This organisation is free.</p>
      </div>
    );
  }

  const stripePromise = loadStripe(publicRuntimeConfig.stripeApiKey);

  return (
    <div className="space-y-3 ">
      <Headline size={4}>Payment Details</Headline>
      <Elements stripe={stripePromise}>
        {subscription?.billingDetails.last4 ? (
          <>
            <p>
              Card in use: {subscription.billingDetails.last4}{' '}
              {subscription.billingDetails.expiryDate}
            </p>
          </>
        ) : (
          <>
            <p>No credit card has been provided yet.</p>
          </>
        )}
        <PaymentForm />
      </Elements>
      {organisation.subscription.id && (
        <SubscriptionInfo
          showInvoices={organisation.canEdit}
          subscriptionId={organisation.subscription.id}
        />
      )}
    </div>
  );
};
export default OrganisationPayment;
