import React from 'react';
import { Link } from 'react-router-dom';

function RecentTransactionsSection({ recentTransactions, formatCurrency, formatDate }) {
  return (
    <section className="dashboard-section">
      <h2>Recent Transactions</h2>
      {recentTransactions.length === 0 ? (
        <div className="no-data-message">
          <p>No transactions yet</p>
          <Link to="/transactions" className="action-link">
            Add your first transaction →
          </Link>
        </div>
      ) : (
        <div className="transactions-list">
          {recentTransactions.map((txn) => (
            <div key={txn.id} className="transaction-item">
              <div className="transaction-details">
                <strong>{txn.description || 'No description'}</strong>
                <small>{formatDate(txn.txn_date)}</small>
              </div>
              <div className={`transaction-amount ${txn.amount_cents >= 0 ? 'income' : 'expense'}`}>
                {txn.amount_cents >= 0 ? '+' : ''}{formatCurrency(txn.amount_cents)}
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

export default RecentTransactionsSection;