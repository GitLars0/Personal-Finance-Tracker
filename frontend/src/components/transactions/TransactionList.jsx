import React from 'react';
import TransactionItem from './TransactionItem';

function TransactionList({ 
  filteredTransactions, 
  transactions, 
  resetForm, 
  setShowForm, 
  formatDate, 
  formatCurrency, 
  handleEdit, 
  handleDelete 
}) {
  if (filteredTransactions.length === 0) {
    return (
      <div className="empty-state">
        <h3>No transactions found</h3>
        <p>
          {transactions.length === 0 
            ? "You haven't added any transactions yet."
            : "No transactions match your current filters."
          }
        </p>
        <button 
          className="action-button"
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
        >
          Add Your First Transaction
        </button>
      </div>
    );
  }

  return (
    <div className="transactions-list">
      <div className="transactions-table">
        <div className="table-header">
          <div>Date</div>
          <div>Description</div>
          <div>Category</div>
          <div>Account</div>
          <div>Amount</div>
          <div>Actions</div>
        </div>
        
        {filteredTransactions.map(transaction => (
          <TransactionItem
            key={transaction.id}
            transaction={transaction}
            formatDate={formatDate}
            formatCurrency={formatCurrency}
            handleEdit={handleEdit}
            handleDelete={handleDelete}
          />
        ))}
      </div>
    </div>
  );
}

export default TransactionList;