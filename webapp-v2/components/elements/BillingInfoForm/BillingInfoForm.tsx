import { CardElement, useElements, useStripe } from '@stripe/react-stripe-js';
import Spinner from 'components/elements/Spinner/Spinner';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useRouter } from 'next/dist/client/router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { clearTempValue } from 'utils/session';

import { setupSubscription, createSubscription, updateUser } from './BillingUtil';
import { EuropeanCountries, Countries } from './Countries';

interface IBillingInfoForm {
  invited: boolean;
  owner: boolean;
  userId: string;
  membershipId: string;
}

const BillingInfoForm = ({
  invited,
  owner,
  userId,
  membershipId,
}: IBillingInfoForm): JSX.Element => {
  const router = useRouter();
  const stripe = useStripe();
  const elements = useElements();

  const {
    register,
    formState: { errors },
    handleSubmit,
    setValue,
    getValues,
  } = useForm();

  const [loading, setLoading] = useState(false);
  const [cardError, setCardError] = useState(undefined);
  const [showVatIdField, setShowVatIdField] = useState(false);
  //const country = register('country', { required: 'required' });

  const handleSubscriptionSetup = (paymentMethodId): void => {
    setupSubscription({
      membershipId,
      name: getValues('name'),
      password: getValues('password'),
      paymentMethodId,
      street_and_number: getValues('street_and_number'),
      city: getValues('city'),
      postal_code: getValues('postal_code'),
      country: getValues('country'),
      notifyInvoice: getValues('notifyInvoice'),
      notifyDeviceUpdate: getValues('notifyDeviceUpdate'),
      vatNumber: getValues('vatNumber'),
    })
      .then(async ({ error, membershipId, setup_intent }) => {
        if (error) {
          setLoading(false);
          console.log('error:', error);
          setCardError(error);
        } else {
          if (setup_intent.status == 'requires_action') {
            const response = await stripe.confirmCardSetup(setup_intent.client_secret);

            if (response.error) {
              console.log('error:', response.error.message);
              setCardError(response.error.message);
              setLoading(false);
            } else {
              setCardError(undefined);
              handleSubcriptionCreation(membershipId, paymentMethodId);
            }
          } else if (setup_intent.status == 'succeeded') {
            setCardError(undefined);
            handleSubcriptionCreation(membershipId, paymentMethodId);
          } else {
            setLoading(false);
            setCardError('An error occured. Please contact the support');
          }
        }
      })
      .catch((errors) => {
        setLoading(false);
        console.log(errors);
      });
  };

  const handleSubcriptionCreation = (membershipId, paymentMethodId): void => {
    createSubscription({ membershipId, paymentMethodId })
      .then(async () => {
        clearTempValue('membershipId');
        router.push(NextJsRoutes.dashboardWelcomePath);
      })
      .catch((errors) => {
        setLoading(false);
        console.log(errors);
      });
  };

  const handleAdressRegistration = (): void => {
    updateUser({
      street_and_number: getValues('street_and_number'),
      city: getValues('city'),
      postal_code: getValues('postal_code'),
      country: getValues('country'),
      name: getValues('name'),
      userId,
      password: getValues('password'),
    })
      .then(async () => {
        router.push(NextJsRoutes.dashboardWelcomePath);
      })
      .catch((errors) => {
        console.log(errors);
      });
  };

  const submit = async (): Promise<unknown> => {
    setLoading(true);
    // Block native form submission.
    if (invited && !owner) {
      // just post to user
      handleAdressRegistration();
      return;
    }

    if (!stripe || !elements) {
      // Stripe.js has not loaded yet. Make sure to disable
      // form submission until Stripe.js has loaded.
      return;
    }

    // Get a reference to a mounted CardElement. Elements knows how
    // to find your CardElement because there can only ever be one of
    // each type of element.
    const cardElement = elements.getElement(CardElement);

    // Use your card Element with other Stripe.js APIs
    const { error, paymentMethod } = await stripe.createPaymentMethod({
      type: 'card',
      card: cardElement,
    });

    if (error) {
      console.log('[error]', error);
      setCardError(error.message);
      setLoading(false);
    } else {
      console.log('[PaymentMethod]', paymentMethod);
      handleSubscriptionSetup(paymentMethod.id);
    }
  };

  return (
    <form onSubmit={handleSubmit(submit)} className="grid grid-flow-row gap-7 w-1/3">
      {invited && (
        <>
          <div className="grid grid-flow-row">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              className="border rounded p-2 text-primary"
              {...register('password', {
                required: 'required',
              })}
              type="password"
            />
            {errors.password && <span role="alert">{errors.password.message}</span>}
          </div>
          <div className="grid grid-flow-row">
            <label htmlFor="name">Name</label>
            <input
              id="name"
              className="border rounded p-2 text-primary"
              {...register('name', {
                required: 'required',
              })}
              type="text"
            />
            {errors.name && <span role="alert">{errors.name.message}</span>}
          </div>
        </>
      )}
      <h2 className="text-3xl">Address</h2>
      <div className="grid grid-flow-row">
        <label htmlFor="street_and_number">Street & Number</label>
        <input
          id="street_and_number"
          className="border rounded p-2 text-primary"
          {...register('street_and_number', {
            required: 'required',
          })}
          type="text"
        />
        {errors.street_and_number && <span role="alert">{errors.street_and_number.message}</span>}
      </div>
      <div className="grid grid-flow-row">
        <label htmlFor="city">City</label>
        <input
          id="city"
          className="border rounded p-2 text-primary"
          {...register('city', {
            required: 'required',
          })}
          type="text"
        />
        {errors.city && <span role="alert">{errors.city.message}</span>}
      </div>
      <div className="grid grid-flow-row">
        <label htmlFor="postal_code">Postal Code</label>
        <input
          id="postal_code"
          className="border rounded p-2 text-primary"
          {...register('postal_code', {
            required: 'required',
          })}
          type="text"
        />
        {errors.postal_code && <span role="alert">{errors.postal_code.message}</span>}
      </div>
      <div className="grid grid-flow-row">
        <label htmlFor="country">Country</label>
        <select
          id="country"
          className="border rounded p-2 text-primary"
          onChange={(e) => {
            setValue('country', e.target.selectedOptions[0].value, { shouldValidate: true });
            setShowVatIdField(EuropeanCountries.includes(getValues('country')));
          }}>
          {Countries.map((country) => (
            <option key={country} value={country}>
              {country}
            </option>
          ))}
        </select>
        {errors.country && <span role="alert">{errors.country.message}</span>}
      </div>
      {(!invited || (invited && owner)) && (
        <>
          {showVatIdField && (
            <div className="grid grid-flow-row">
              <label htmlFor="vatNumber">European VAT-ID</label>
              <input
                id="vatNumber"
                className="border rounded p-2 text-primary"
                {...register('vatNumber', {})}
                type="text"
              />
              {errors.vatNumber && <span role="alert">{errors.vatNumber.message}</span>}
            </div>
          )}
          <h2 className="text-3xl">Billing</h2>
          <div className="grid grid-flow-row">
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
          <div className="border-b border-white" />
          <div className="grid grid-flow-row">
            <label htmlFor="notifyInvoice">Notify invoices</label>
            <input
              id="notifyInvoice"
              className="border rounded p-2 text-primary"
              {...register('notifyInvoice', {})}
              type="checkbox"
            />
            {errors.notifyInvoice && <span role="alert">{errors.notifyInvoice.message}</span>}
          </div>
        </>
      )}
      <div className="grid grid-flow-row">
        <label htmlFor="notifyDeviceUpdate">Notify device updates</label>
        <input
          id="notifyDeviceUpdate"
          className="border rounded p-2 text-primary"
          {...register('notifyDeviceUpdate', {})}
          type="checkbox"
        />
        {errors.notifyDeviceUpdate && <span role="alert">{errors.notifyDeviceUpdate.message}</span>}
      </div>
      {loading ? <Spinner></Spinner> : <button type="submit">Register</button>}
    </form>
  );
};

export default BillingInfoForm;
