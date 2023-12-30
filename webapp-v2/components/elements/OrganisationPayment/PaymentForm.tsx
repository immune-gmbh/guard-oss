import { useStripe, useElements, CardElement } from '@stripe/react-stripe-js';
import Button from 'components/elements/Button/Button';
import Spinner from 'components/elements/Spinner/Spinner';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import getConfig from 'next/config';
import { useRouter } from 'next/router';
import { useState } from 'react';
import { toast } from 'react-toastify';
import { useSWRConfig } from 'swr';
import { authenticatedJsonapiRequest } from 'utils/fetcher';
import { deserialiseWithoutMeta } from 'utils/session';
import { SerializedSubscription } from 'utils/types';

export default function CardSetupForm(): JSX.Element {
  const { publicRuntimeConfig } = getConfig();

  const router = useRouter();
  const stripe = useStripe();
  const elements = useElements();
  const [isLoading, setIsLoading] = useState(false);
  const [cardError, setCardError] = useState(undefined);
  const { mutate } = useSWRConfig();

  const handleSubmit = async (): Promise<unknown> => {
    setIsLoading(true);

    if (!stripe || !elements) {
      // Stripe.js has not yet loaded.
      // Make sure to disable form submission until Stripe.js has loaded.
      setIsLoading(false);
      setCardError(`Stripe still loading. Please wait.`);
      return;
    }

    const data = (await authenticatedJsonapiRequest(
      `${publicRuntimeConfig.hosts.authSrv}/v2/subscriptions/intent`,
      {
        method: 'POST',
        body: JSON.stringify({
          organisation_id: router.query.id,
        }),
      },
    )) as any;

    if (data.errors) {
      setIsLoading(false);
      setCardError(data.errors[0]?.title || 'Illformed response');
      return;
    }

    const clientSecret = data.data?.attributes?.setup_intent_secret;
    const { error, setupIntent } = await stripe.confirmCardSetup(clientSecret, {
      payment_method: {
        card: elements.getElement(CardElement),
      },
    });

    if (error) {
      setIsLoading(false);
      toast.error(error.message || 'Something went wrong');
    } else {
      const data = (await authenticatedJsonapiRequest(
        `${publicRuntimeConfig.hosts.authSrv}/v2/subscriptions/default_payment_method`,
        {
          method: 'POST',
          body: JSON.stringify({
            organisation_id: router.query.id,
            payment_method_id: setupIntent.payment_method,
          }),
        },
      )) as any;

      if (data.errors) {
        toast.error('Updating payment details failed!');
      } else {
        const subscription = deserialiseWithoutMeta(data) as SerializedSubscription;
        toast.success('Payment details have been updated');
        mutate(JsRoutesRails.v2_subscription_path({ id: subscription.id }));
      }
      setIsLoading(false);
    }
  };

  return (
    <section className="space-y-4">
      <div>
        {isLoading && <Spinner />}
        <label htmlFor="credit_card">Credit Card</label>
        <div className="border rounded p-2 text-primary bg-white">
          <CardElement
            options={{
              iconStyle: 'solid',
              style: {
                base: {
                  fontSize: '16px',
                  backgroundColor: '#FFF',
                  padding: '4px',
                  color: 'rgba(0, 0, 0, 0.94)',
                  '::placeholder': {
                    color: 'rgba(0, 0, 0, 0.6)',
                  },
                },
                invalid: {
                  color: '#FFD1D8',
                },
              },
            }}
          />
        </div>
        {cardError && <div id="card-errors">{cardError}</div>}
      </div>
      <Button theme="CTA" type="submit" onClick={handleSubmit}>
        Save Card
      </Button>
    </section>
  );
}
