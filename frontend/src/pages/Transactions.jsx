import React, { useEffect, useState, useCallback } from 'react';
import '../styles/Transactions.css';
import { TransactionForm, TransactionList, TransactionFilters, TransactionSummary } from '../components/transactions';
import DeleteTransactionModal from '../components/transactions/DeleteTransactionModal';

function Transactions() {
  const [transactions, setTransactions] = useState([]);
  const [categories, setCategories] = useState([]);
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingTransaction, setEditingTransaction] = useState(null);
  const [deleteTarget, setDeleteTarget] = useState(null);

  // Form state
  const [formData, setFormData] = useState({
    account_id: '',
    category_id: '',
    amount_cents: '',
    description: '',
    txn_date: new Date().toISOString().split('T')[0],
    notes: ''
  });

  // Filters
  const [filters, setFilters] = useState({
    search: '',
    category_id: '',
    account_id: '',
    from: '',
    to: ''
  });

  const token = localStorage.getItem('token');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const headers = { 'Authorization': `Bearer ${token}` };

      // Fetch all required data
      const [transactionsRes, categoriesRes, accountsRes] = await Promise.all([
        fetch('/api/transactions', { headers }),
        fetch('/api/categories', { headers }),
        fetch('/api/accounts', { headers }).catch(() => ({ ok: false })) // accounts might not exist yet
      ]);

      if (!transactionsRes.ok && transactionsRes.status !== 404) {
        throw new Error('Failed to fetch transactions');
      }

      if (!categoriesRes.ok && categoriesRes.status !== 404) {
        throw new Error('Failed to fetch categories');
      }

      const transactions = transactionsRes.ok ? await transactionsRes.json() : [];
      const categories = categoriesRes.ok ? await categoriesRes.json() : [];
      const accounts = accountsRes.ok ? await accountsRes.json() : [];

      setTransactions(transactions);
      setCategories(categories);
      setAccounts(accounts);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) {
      setError('Not authenticated');
      setLoading(false);
      return;
    }
    fetchData();
  }, [token, fetchData]);

  const handleFormSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.account_id || !formData.amount_cents || !formData.txn_date) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      const submitData = {
        ...formData,
        amount_cents: parseInt(formData.amount_cents * 100), // Convert to cents
        category_id: formData.category_id ? parseInt(formData.category_id) : null,
        account_id: parseInt(formData.account_id)
      };

      const url = editingTransaction 
        ? `/api/transactions/${editingTransaction.id}`
        : '/api/transactions';
      
      const method = editingTransaction ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers,
        body: JSON.stringify(submitData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save transaction');
      }

      // Success - refresh data and close form
      await fetchData();
      resetForm();
      setShowForm(false);
      setEditingTransaction(null);
      
      // Dispatch event to notify other components (like Budget page)
      const eventType = editingTransaction ? 'transactionUpdated' : 'transactionCreated';
      console.log(`Transactions page: Dispatching ${eventType} event`);
      window.dispatchEvent(new Event(eventType));
      localStorage.setItem('lastTransactionUpdate', Date.now().toString());
    } catch (err) {
      setError(err.message);
    }
  };

  const handleEdit = (transaction) => {
    setEditingTransaction(transaction);
    setFormData({
      account_id: transaction.account_id?.toString() || '',
      category_id: transaction.category_id?.toString() || '',
      amount_cents: (Math.abs(transaction.amount_cents) / 100).toString(),
      description: transaction.description || '',
      txn_date: transaction.txn_date?.split('T')[0] || '',
      notes: transaction.notes || ''
    });
    setShowForm(true);
  };

  const handleDelete = (transaction) => {
    setDeleteTarget(transaction);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;

    try {
      const response = await fetch(`/api/transactions/${deleteTarget.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete transaction');
      }

      await fetchData();
      setDeleteTarget(null);
      
      // Dispatch event to notify other components (like Budget page)
      console.log('Transactions page: Dispatching transactionDeleted event');
      window.dispatchEvent(new Event('transactionDeleted'));
      localStorage.setItem('lastTransactionUpdate', Date.now().toString());
    } catch (err) {
      setError(err.message);
    }
  };

  const cancelDelete = () => {
    setDeleteTarget(null);
  };

  const resetForm = () => {
    setFormData({
      account_id: '',
      category_id: '',
      amount_cents: '',
      description: '',
      txn_date: new Date().toISOString().split('T')[0],
      notes: ''
    });
    setEditingTransaction(null);
    setError(null);
  };

  const handleCategoryCreated = (newCategory) => {
    console.log('New category created:', newCategory);
    setCategories(prev => {
      console.log('Previous categories:', prev);
      const updated = [...prev, newCategory];
      console.log('Updated categories:', updated);
      return updated;
    });
  };

  const handleAccountCreated = (newAccount) => {
    console.log('New account created:', newAccount);
    setAccounts(prev => {
      console.log('Previous accounts:', prev);
      const updated = [...prev, newAccount];
      console.log('Updated accounts:', updated);
      return updated;
    });
  };

  const formatCurrency = (cents) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(cents / 100);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  // Apply filters
  const filteredTransactions = transactions.filter(txn => {
    if (filters.search && !txn.description?.toLowerCase().includes(filters.search.toLowerCase())) {
      return false;
    }
    if (filters.category_id && txn.category_id?.toString() !== filters.category_id) {
      return false;
    }
    if (filters.account_id && txn.account_id?.toString() !== filters.account_id) {
      return false;
    }
    if (filters.from && txn.txn_date < filters.from) {
      return false;
    }
    if (filters.to && txn.txn_date > filters.to) {
      return false;
    }
    return true;
  });

  if (loading) {
    return (
      <div className="transactions-container">
        <div className="loading">Loading transactions...</div>
      </div>
    );
  }

  if (error && !showForm) {
    return (
      <div className="transactions-container">
        <div className="error-message">{error}</div>
        <button onClick={fetchData}>Retry</button>
      </div>
    );
  }

  return (
    <div className="transactions-container">
      <header className="transactions-header">
        <h1>Transactions</h1>
        <button 
          className="add-transaction-btn"
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
        >
          ➕ Add Transaction
        </button>
      </header>

      {error && <div className="error-message">{error}</div>}

      <TransactionForm 
        showForm={showForm}
        editingTransaction={editingTransaction}
        formData={formData}
        setFormData={setFormData}
        handleFormSubmit={handleFormSubmit}
        resetForm={resetForm}
        setShowForm={setShowForm}
        setEditingTransaction={setEditingTransaction}
        accounts={accounts}
        categories={categories}
        onCategoryCreated={handleCategoryCreated}
        onAccountCreated={handleAccountCreated}
        error={error}
      />

      <TransactionFilters 
        filters={filters}
        setFilters={setFilters}
        categories={categories}
        accounts={accounts}
      />

      <TransactionList 
        filteredTransactions={filteredTransactions}
        transactions={transactions}
        resetForm={resetForm}
        setShowForm={setShowForm}
        formatDate={formatDate}
        formatCurrency={formatCurrency}
        handleEdit={handleEdit}
        handleDelete={handleDelete}
      />

      <TransactionSummary 
        filteredTransactions={filteredTransactions}
        formatCurrency={formatCurrency}
      />

      {deleteTarget && (
        <DeleteTransactionModal
          transaction={deleteTarget}
          onConfirm={confirmDelete}
          onCancel={cancelDelete}
          formatCurrency={formatCurrency}
          formatDate={formatDate}
        />
      )}
    </div>
  );
}

export default Transactions;