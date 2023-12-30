/**
 * @jest-environment jsdom
 */
import { describe, expect, jest, test } from '@jest/globals';
import * as stripe from '@stripe/react-stripe-js';
import { render, screen } from '@testing-library/react';
import CardSetupForm from 'components/elements/OrganisationPayment/PaymentForm';
import { mockStripeFn } from 'tests/mocks';

jest.mock('@stripe/react-stripe-js');
const { Elements } = jest.requireActual('@stripe/react-stripe-js') as any;
const mockedStripeReact = jest.mocked(stripe);


describe('card setup form', () => {
  let mockStripe: any;

  beforeEach(() => {
    mockStripe = mockStripeFn();

    mockedStripeReact.CardElement.mockReturnValue(
      <div>CardElement Component</div>
    )
  });

  test('renders correctly', () => {

    render(
      <Elements stripe={mockStripe}>
        <CardSetupForm />
      </Elements>
    );

    expect(mockedStripeReact.CardElement).toHaveBeenCalled();
    expect(screen.queryByText("CardElement Component")).toBeInTheDocument();
  });
});
