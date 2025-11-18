import React from 'react';

function AccountSummary({ accounts, formatCurrency }) {
  if (accounts.length === 0) {
    return null;
  }

  const totalBalance = accounts.reduce((sum, acc) => 
    sum + (acc.current_balance_cents || acc.initial_balance_cents || 0), 0);
  
  const assets = accounts
    .filter(acc => (acc.current_balance_cents || acc.initial_balance_cents || 0) > 0)
    .reduce((sum, acc) => sum + (acc.current_balance_cents || acc.initial_balance_cents || 0), 0);
  
  const liabilities = Math.abs(accounts
    .filter(acc => (acc.current_balance_cents || acc.initial_balance_cents || 0) < 0)
    .reduce((sum, acc) => sum + (acc.current_balance_cents || acc.initial_balance_cents || 0), 0));

  return (
    <div className="accounts-summary">
      <h3>Account Summary</h3>
      <div className="summary-grid">
        <div className="summary-item">
          <span>Total Accounts:</span>
          <strong>{accounts.length}</strong>
        </div>
        <div className="summary-item">
          <span>Total Balance:</span>
          <strong className={totalBalance >= 0 ? 'positive' : 'negative'}>
            {formatCurrency(totalBalance)}
          </strong>
        </div>
        <div className="summary-item">
          <span>Assets:</span>
          <strong className="positive">
            {formatCurrency(assets)}
          </strong>
        </div>
        <div className="summary-item">
          <span>Liabilities:</span>
          <strong className="negative">
            {formatCurrency(liabilities)}
          </strong>
        </div>
      </div>
    </div>
  );
}

export default AccountSummary;