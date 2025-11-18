import React from 'react';
import CategorySelector from '../categories/CategorySelector';
import AccountSelector from '../accounts/AccountSelector';

function TransactionForm({ 
  showForm, 
  editingTransaction, 
  formData, 
  setFormData, 
  handleFormSubmit, 
  resetForm, 
  setShowForm, 
  setEditingTransaction,
  accounts, 
  categories, 
  onCategoryCreated,
  onAccountCreated,
  error 
}) {
  if (!showForm) return null;

  return (
    <div className="form-overlay">
      <div className="transaction-form">
        <h2>{editingTransaction ? 'Edit Transaction' : 'Add New Transaction'}</h2>
        
        <form onSubmit={handleFormSubmit}>
          <div className="form-row">
            <AccountSelector
              value={formData.account_id}
              onChange={(value) => setFormData({...formData, account_id: value})}
              accounts={accounts}
              onAccountCreated={onAccountCreated}
              required={true}
              label="Account"
            />

            <CategorySelector
              value={formData.category_id}
              onChange={(value) => setFormData({...formData, category_id: value})}
              categories={categories}
              onCategoryCreated={onCategoryCreated}
              label="Category"
              defaultKind={parseFloat(formData.amount_cents) >= 0 ? 'income' : 'expense'}
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Amount (USD) *</label>
              <input
                type="number"
                step="0.01"
                placeholder="0.00"
                value={formData.amount_cents}
                onChange={(e) => setFormData({...formData, amount_cents: e.target.value})}
                required
              />
              <small className="form-help">Positive for income, negative for expenses</small>
            </div>

            <div className="form-group">
              <label>Date *</label>
              <input
                type="date"
                value={formData.txn_date}
                onChange={(e) => setFormData({...formData, txn_date: e.target.value})}
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label>Description</label>
            <input
              type="text"
              placeholder="Transaction description"
              value={formData.description}
              onChange={(e) => setFormData({...formData, description: e.target.value})}
            />
          </div>

          <div className="form-group">
            <label>Notes</label>
            <textarea
              placeholder="Additional notes (optional)"
              value={formData.notes}
              onChange={(e) => setFormData({...formData, notes: e.target.value})}
              rows={3}
            />
          </div>

          <div className="form-actions">
            <button type="button" onClick={() => {
              setShowForm(false);
              setEditingTransaction(null);
              resetForm();
            }}>
              Cancel
            </button>
            <button type="submit" className="primary">
              {editingTransaction ? 'Update Transaction' : 'Add Transaction'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default TransactionForm;