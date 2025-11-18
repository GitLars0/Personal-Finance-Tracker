import React from 'react';

function TransactionItem({ 
  transaction, 
  formatDate, 
  formatCurrency, 
  handleEdit, 
  handleDelete 
}) {
  return (
    <div className="table-row">
      <div className="transaction-date">
        {formatDate(transaction.txn_date)}
      </div>
      
      <div className="transaction-description">
        <strong>{transaction.description || 'No description'}</strong>
        {transaction.notes && (
          <small className="transaction-notes">{transaction.notes}</small>
        )}
      </div>
      
      <div className="transaction-category">
        {transaction.category?.name || 'Uncategorized'}
        {transaction.category && (
          <small className={`category-kind ${transaction.category.kind}`}>
            {transaction.category.kind}
          </small>
        )}
      </div>
      
      <div className="transaction-account">
        {transaction.account?.name || 'Unknown Account'}
        {transaction.account && (
          <small>{transaction.account.account_type}</small>
        )}
      </div>
      
      <div className={`transaction-amount ${transaction.amount_cents >= 0 ? 'income' : 'expense'}`}>
        {transaction.amount_cents >= 0 ? '+' : ''}{formatCurrency(transaction.amount_cents)}
      </div>
      
      <div className="transaction-actions">
        <button 
          onClick={() => handleEdit(transaction)}
          className="edit-btn"
          title="Edit transaction"
        >
          ✏️
        </button>
        <button 
          onClick={() => handleDelete(transaction)}
          className="delete-btn"
          title="Delete transaction"
        >
          🗑️
        </button>
      </div>
    </div>
  );
}

export default TransactionItem;