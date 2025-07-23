import React, { useState } from 'react';
import { loadStripe, PaymentRequest, StripePaymentRequestButtonElementClickEvent } from '@stripe/stripe-js';
import { Elements, CardElement, useStripe, useElements, PaymentRequestButtonElement } from '@stripe/react-stripe-js';
import api from '../utils/api';

const stripePromise = loadStripe(process.env.REACT_APP_STRIPE_PUBLISHABLE_KEY || '');

const BuyTokensForm: React.FC = () => {
  const stripe = useStripe();
  const elements = useElements();
  const [amount, setAmount] = useState(10);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState('');
  const [error, setError] = useState('');
  const [paymentRequest, setPaymentRequest] = useState<PaymentRequest | null>(null);
  const [prButtonReady, setPrButtonReady] = useState(false);

  React.useEffect(() => {
    if (stripe) {
      const pr = stripe.paymentRequest({
        country: 'US',
        currency: 'usd',
        total: {
          label: 'Buy Tokens',
          amount: amount * 100,
        },
        requestPayerName: true,
        requestPayerEmail: true,
      });
      pr.on('paymentmethod', handlePaymentRequest);
      pr.canMakePayment().then(result => {
        if (result) setPaymentRequest(pr);
      });
    }
  }, [stripe, amount]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');
    if (!stripe || !elements) {
      setError('Stripe not loaded');
      setLoading(false);
      return;
    }
    try {
      // 1. Create PaymentIntent on backend
      const response = await api.post('/tokens/purchase', { amount });
      const clientSecret = response.data.client_secret;
      // 2. Confirm card payment
      const result = await stripe.confirmCardPayment(clientSecret, {
        payment_method: {
          card: elements.getElement(CardElement)!,
        },
      });
      if (result.error) {
        setError(result.error.message || 'Payment failed');
      } else if (result.paymentIntent && result.paymentIntent.status === 'succeeded') {
        setSuccess('Payment successful! Tokens will be credited to your account.');
      } else {
        setError('Payment failed');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Payment failed');
    } finally {
      setLoading(false);
    }
  };

  const handlePaymentRequest = (event: any) => {
    event.complete('success');
    setSuccess('Payment successful! Tokens will be credited to your account.');
  };

  return (
    <div className="buy-tokens-page">
      <div className="container">
        <h1>Buy Tokens</h1>
        <p>Purchase tokens to support authors and unlock content. You can pay with card, Apple Pay, or Google Pay.</p>
        {error && <div className="alert alert-error">{error}</div>}
        {success && <div className="alert alert-success">{success}</div>}
        <form onSubmit={handleSubmit} className="buy-tokens-form">
          <label htmlFor="amount" className="form-label">Amount ($):</label>
          <input
            type="number"
            id="amount"
            value={amount}
            onChange={e => setAmount(Number(e.target.value))}
            min={1}
            step={1}
            className="form-input"
            required
          />
          <p>You will receive <b>{amount * 10}</b> tokens.</p>

          <div className="form-group">
            <label className="form-label">Card Details</label>
            <CardElement options={{ hidePostalCode: true }} className="card-element" />
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading || !stripe}>
            {loading ? 'Processing...' : `Buy ${amount * 10} tokens for $${amount}`}
          </button>
        </form>

        {/* Payment Request Button (Apple Pay, Google Pay) */}
        {paymentRequest && (
          <div className="payment-request-section">
            <p>Or pay with Apple Pay / Google Pay:</p>
            <PaymentRequestButtonElement
              options={{ paymentRequest }}
              onReady={() => setPrButtonReady(true)}
              onClick={(e: StripePaymentRequestButtonElementClickEvent) => {
                if (!prButtonReady) e.preventDefault();
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
};

const BuyTokens: React.FC = () => (
  <Elements stripe={stripePromise}>
    <BuyTokensForm />
  </Elements>
);

export default BuyTokens; 