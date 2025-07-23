import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import api from '../utils/api';

interface TokenBalance {
  balance: number;
  total_earned: number;
  total_spent: number;
}

interface TokenTransaction {
  id: number;
  transaction_type: 'purchase' | 'tip' | 'cashout' | 'refund';
  amount: number;
  status: 'pending' | 'completed' | 'failed' | 'cancelled';
  created_at: string;
  recipient?: {
    display_name: string;
  };
  chapter?: {
    title: string;
  };
}

const Tokens: React.FC = () => {
  const { user, updateProfile } = useAuth();
  const [balance, setBalance] = useState<TokenBalance | null>(null);
  const [transactions, setTransactions] = useState<TokenTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [purchaseAmount, setPurchaseAmount] = useState(10);
  const [cashoutAmount, setCashoutAmount] = useState(10);
  const [purchasing, setPurchasing] = useState(false);
  const [cashingOut, setCashingOut] = useState(false);

  useEffect(() => {
    fetchTokenData();
  }, []);

  const fetchTokenData = async () => {
    try {
      setLoading(true);
      const [balanceResponse, transactionsResponse] = await Promise.all([
        api.get('/tokens/balance'),
        api.get('/tokens/transactions')
      ]);
      
      setBalance(balanceResponse.data);
      setTransactions(transactionsResponse.data.transactions || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch token data');
    } finally {
      setLoading(false);
    }
  };

  const handlePurchase = async () => {
    if (purchaseAmount < 1) {
      setError('Minimum purchase amount is $1');
      return;
    }

    setPurchasing(true);
    setError('');

    try {
      const response = await api.post('/tokens/purchase', {
        amount: purchaseAmount
      });

      // In a real implementation, you would integrate with Stripe here
      // For now, we'll simulate a successful purchase
      alert(`Purchase successful! You will receive ${response.data.tokens_to_award} tokens.`);
      
      // Refresh token data
      await fetchTokenData();
      
      // Update user profile to reflect new balance
      if (user) {
        updateProfile({ ...user, token_balance: (user.token_balance || 0) + response.data.tokens_to_award });
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Purchase failed');
    } finally {
      setPurchasing(false);
    }
  };

  const handleCashout = async () => {
    if (cashoutAmount < 10) {
      setError('Minimum cashout amount is 10 tokens');
      return;
    }

    if (!balance || balance.balance < cashoutAmount) {
      setError('Insufficient token balance');
      return;
    }

    setCashingOut(true);
    setError('');

    try {
      const response = await api.post('/tokens/cashout', {
        amount: cashoutAmount
      });

      alert(`Cashout request submitted! You will receive $${response.data.payout_amount_usd.toFixed(2)}.`);
      
      // Refresh token data
      await fetchTokenData();
      
      // Update user profile to reflect new balance
      if (user) {
        updateProfile({ ...user, token_balance: (user.token_balance || 0) - cashoutAmount });
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Cashout failed');
    } finally {
      setCashingOut(false);
    }
  };

  const getTransactionDescription = (transaction: TokenTransaction) => {
    switch (transaction.transaction_type) {
      case 'purchase':
        return `Purchased ${transaction.amount} tokens`;
      case 'tip':
        if (transaction.amount > 0) {
          return `Received tip of ${transaction.amount} tokens${transaction.chapter ? ` for "${transaction.chapter.title}"` : ''}`;
        } else {
          return `Tipped ${Math.abs(transaction.amount)} tokens${transaction.recipient ? ` to ${transaction.recipient.display_name}` : ''}`;
        }
      case 'cashout':
        return `Cashed out ${Math.abs(transaction.amount)} tokens`;
      case 'refund':
        return `Refund of ${transaction.amount} tokens`;
      default:
        return 'Transaction';
    }
  };

  const getTransactionStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'text-green-600';
      case 'pending':
        return 'text-yellow-600';
      case 'failed':
        return 'text-red-600';
      case 'cancelled':
        return 'text-gray-600';
      default:
        return 'text-gray-600';
    }
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="tokens-page">
      <div className="container">
        <div className="page-header">
          <h1>💰 Token Management</h1>
          <p>Buy tokens to support authors and cash out your earnings</p>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        {/* Token Balance Overview */}
        {balance && (
          <div className="balance-overview">
            <div className="stats-grid">
              <div className="card">
                <h3>Current Balance</h3>
                <p className="stat-number">💰 {balance.balance} tokens</p>
                <p className="stat-value">≈ ${(balance.balance / 10).toFixed(2)} USD</p>
              </div>
              <div className="card">
                <h3>Total Earned</h3>
                <p className="stat-number">📈 {balance.total_earned} tokens</p>
                <p className="stat-value">≈ ${(balance.total_earned / 10).toFixed(2)} USD</p>
              </div>
              <div className="card">
                <h3>Total Spent</h3>
                <p className="stat-number">💸 {balance.total_spent} tokens</p>
                <p className="stat-value">≈ ${(balance.total_spent / 10).toFixed(2)} USD</p>
              </div>
            </div>
          </div>
        )}

        {/* Purchase Tokens */}
        <div className="section">
          <h2>Buy Tokens</h2>
          <div className="card">
            <p className="mb-4">Tokens cost $1 for 10 tokens. Purchase tokens to support your favorite authors!</p>
            
            <div className="purchase-options">
              <button
                className={`purchase-option ${purchaseAmount === 5 ? 'selected' : ''}`}
                onClick={() => setPurchaseAmount(5)}
              >
                $5 = 50 tokens
              </button>
              <button
                className={`purchase-option ${purchaseAmount === 10 ? 'selected' : ''}`}
                onClick={() => setPurchaseAmount(10)}
              >
                $10 = 100 tokens
              </button>
              <button
                className={`purchase-option ${purchaseAmount === 25 ? 'selected' : ''}`}
                onClick={() => setPurchaseAmount(25)}
              >
                $25 = 250 tokens
              </button>
              <button
                className={`purchase-option ${purchaseAmount === 50 ? 'selected' : ''}`}
                onClick={() => setPurchaseAmount(50)}
              >
                $50 = 500 tokens
              </button>
            </div>

            <div className="custom-amount">
              <label htmlFor="custom-purchase" className="form-label">Custom amount ($):</label>
              <input
                type="number"
                id="custom-purchase"
                value={purchaseAmount}
                onChange={(e) => setPurchaseAmount(parseFloat(e.target.value) || 0)}
                className="form-input"
                min="1"
                step="0.01"
              />
              <p className="token-preview">
                You'll receive: {Math.floor(purchaseAmount * 10)} tokens
              </p>
            </div>

            <button
              onClick={handlePurchase}
              disabled={purchasing || purchaseAmount < 1}
              className="btn btn-primary"
            >
              {purchasing ? 'Processing...' : `Buy ${Math.floor(purchaseAmount * 10)} tokens for $${purchaseAmount.toFixed(2)}`}
            </button>
          </div>
        </div>

        {/* Cash Out Tokens (Authors Only) */}
        {(user?.role === 'author' || user?.role === 'admin') && balance && balance.total_earned > 0 && (
          <div className="section">
            <h2>Cash Out Earnings</h2>
            <div className="card">
              <p className="mb-4">
                Cash out your earned tokens to USD. Minimum cashout is 10 tokens. 
                Payout rates vary based on your performance.
              </p>
              
              <div className="cashout-options">
                <button
                  className={`cashout-option ${cashoutAmount === 10 ? 'selected' : ''}`}
                  onClick={() => setCashoutAmount(10)}
                >
                  10 tokens
                </button>
                <button
                  className={`cashout-option ${cashoutAmount === 25 ? 'selected' : ''}`}
                  onClick={() => setCashoutAmount(25)}
                >
                  25 tokens
                </button>
                <button
                  className={`cashout-option ${cashoutAmount === 50 ? 'selected' : ''}`}
                  onClick={() => setCashoutAmount(50)}
                >
                  50 tokens
                </button>
                <button
                  className={`cashout-option ${cashoutAmount === 100 ? 'selected' : ''}`}
                  onClick={() => setCashoutAmount(100)}
                >
                  100 tokens
                </button>
              </div>

              <div className="custom-amount">
                <label htmlFor="custom-cashout" className="form-label">Custom amount (tokens):</label>
                <input
                  type="number"
                  id="custom-cashout"
                  value={cashoutAmount}
                  onChange={(e) => setCashoutAmount(parseInt(e.target.value) || 0)}
                  className="form-input"
                  min="10"
                  max={balance.balance}
                />
                <p className="cashout-preview">
                  Estimated payout: ${(cashoutAmount * 0.75 / 10).toFixed(2)} USD
                </p>
              </div>

              <button
                onClick={handleCashout}
                disabled={cashingOut || cashoutAmount < 10 || cashoutAmount > balance.balance}
                className="btn btn-outline"
              >
                {cashingOut ? 'Processing...' : `Cash out ${cashoutAmount} tokens`}
              </button>
            </div>
          </div>
        )}

        {/* Transaction History */}
        <div className="section">
          <h2>Transaction History</h2>
          <div className="card">
            {transactions.length === 0 ? (
              <p className="empty-state">No transactions yet.</p>
            ) : (
              <div className="transactions-list">
                {transactions.map((transaction) => (
                  <div key={transaction.id} className="transaction-item">
                    <div className="transaction-info">
                      <div className="transaction-description">
                        {getTransactionDescription(transaction)}
                      </div>
                      <div className="transaction-meta">
                        <span className="transaction-date">
                          {new Date(transaction.created_at).toLocaleDateString()}
                        </span>
                        <span className={`transaction-status ${getTransactionStatusColor(transaction.status)}`}>
                          {transaction.status}
                        </span>
                      </div>
                    </div>
                    <div className={`transaction-amount ${transaction.amount > 0 ? 'positive' : 'negative'}`}>
                      {transaction.amount > 0 ? '+' : ''}{transaction.amount} tokens
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Tokens; 