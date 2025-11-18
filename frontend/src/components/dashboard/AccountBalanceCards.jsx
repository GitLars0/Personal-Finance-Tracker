import React from 'react';

function AccountBalanceCards({ accountBalances, formatCurrency }) {
  if (!accountBalances?.accounts?.length) {
    return null;
  }

  return (
    <section className="dashboard-section">
      <h2>Account Balances</h2>
      <div className="cards-grid">
        {accountBalances.accounts.map((account) => (
          <div key={account.account_id} className="info-card">
            <div className="card-header">
              <span className="account-type">{account.account_type}</span>
            </div>
            <h3>{account.account_name}</h3>
            <p className="amount">{formatCurrency(account.balance_cents)}</p>
            <small>{account.transaction_count} transactions</small>
          </div>
        ))}
        <div className="info-card total-card">
          <h3>Total Balance</h3>
          <p className="amount total-amount">
            {formatCurrency(accountBalances.total_balance_cents || 0)}
          </p>
        </div>
      </div>
    </section>
  );
}

export default AccountBalanceCards;