import React from 'react';

function TransactionSummary({ filteredTransactions, formatCurrency }) {
  if (filteredTransactions.length === 0) {
    return null;
  }

  const totalIncome = filteredTransactions
    .filter(t => t.amount_cents > 0)
    .reduce((sum, t) => sum + t.amount_cents, 0);
  
  const totalExpenses = Math.abs(filteredTransactions
    .filter(t => t.amount_cents < 0)
    .reduce((sum, t) => sum + t.amount_cents, 0));
  
  const netAmount = filteredTransactions.reduce((sum, t) => sum + t.amount_cents, 0);

  return (
    <div className="transactions-summary">
      <div className="summary-item">
        <span>Total Transactions:</span>
        <strong>{filteredTransactions.length}</strong>
      </div>
      <div className="summary-item">
        <span>Total Income:</span>
        <strong className="income">
          {formatCurrency(totalIncome)}
        </strong>
      </div>
      <div className="summary-item">
        <span>Total Expenses:</span>
        <strong className="expense">
          {formatCurrency(totalExpenses)}
        </strong>
      </div>
      <div className="summary-item">
        <span>Net Amount:</span>
        <strong className={netAmount >= 0 ? 'income' : 'expense'}>
          {formatCurrency(netAmount)}
        </strong>
      </div>
    </div>
  );
}

export default TransactionSummary;