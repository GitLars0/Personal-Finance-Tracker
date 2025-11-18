import React, { useEffect, useState, useCallback } from 'react';
import '../styles/Budget.css';
import { BudgetForm } from '../components/budget';
import BudgetProgressCard from '../components/budget/BudgetProgressCard';
import useAuthFetch from '../hooks/useAuthFetch';
import Modal from '../components/Modal';
// format helpers are used inside child components

function Budget() {
  const authFetch = useAuthFetch();
  const [budgets, setBudgets] = useState([]);
  const [budgetProgress, setBudgetProgress] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingBudget, setEditingBudget] = useState(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      
      // Fetch basic budget data and categories
      const [budgetsData, categoriesData] = await Promise.all([
        authFetch('/api/budgets'),
        authFetch('/api/categories')
      ]);
      
      setBudgets(budgetsData || []);
      setCategories(categoriesData || []);
      
      // For each budget, fetch its progress data
      const progressPromises = (budgetsData || []).map(async (budget) => {
        try {
          const progressData = await authFetch(`/api/reports/budget-progress?budget_id=${budget.id}`);
          return progressData;
        } catch (error) {
          console.warn(`Failed to fetch progress for budget ${budget.id}:`, error);
          // Return a fallback structure if budget-progress fails
          return {
            budget: {
              id: budget.id,
              period_start: budget.period_start,
              period_end: budget.period_end,
              days_remaining: Math.max(0, Math.ceil((new Date(budget.period_end) - new Date()) / (1000 * 60 * 60 * 24)))
            },
            summary: {
              total_planned_cents: budget.items?.reduce((sum, item) => sum + item.planned_cents, 0) || 0,
              total_spent_cents: 0,
              total_remaining_cents: budget.items?.reduce((sum, item) => sum + item.planned_cents, 0) || 0
            },
            categories: budget.items?.map(item => ({
              category_id: item.category_id,
              category_name: item.category?.name || 'Unknown',
              planned_cents: item.planned_cents,
              spent_cents: 0,
              remaining_cents: item.planned_cents,
              progress_percent: 0,
              status: 'under_budget'
            })) || []
          };
        }
      });
      
      const progressResults = await Promise.all(progressPromises);
      setBudgetProgress(progressResults.filter(Boolean));
      
      setError(null);
    } catch (err) {
      console.error('Budget fetchData error:', err);
      setError(err.message || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  }, [authFetch]);  // Listen for transaction changes from other parts of the app
  useEffect(() => {
    const handleTransactionChange = () => {
      // Refresh budget data when transactions change
      console.log('Budget page: Transaction change detected, refreshing data...');
      fetchData();
    };

    // Listen for custom events
    window.addEventListener('transactionCreated', handleTransactionChange);
    window.addEventListener('transactionUpdated', handleTransactionChange);
    window.addEventListener('transactionDeleted', handleTransactionChange);

    // Listen for storage changes (in case transactions are modified in another tab)
    window.addEventListener('storage', (e) => {
      if (e.key === 'lastTransactionUpdate') {
        console.log('Budget page: Storage change detected for transactions');
        handleTransactionChange();
      }
    });

    console.log('Budget page: Event listeners registered');

    return () => {
      window.removeEventListener('transactionCreated', handleTransactionChange);
      window.removeEventListener('transactionUpdated', handleTransactionChange);
      window.removeEventListener('transactionDeleted', handleTransactionChange);
      window.removeEventListener('storage', handleTransactionChange);
      console.log('Budget page: Event listeners removed');
    };
  }, [fetchData]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleEdit = (budgetProgressData) => { 
    console.log('handleEdit called with:', budgetProgressData);
    
    // Try to find budget by ID first (most reliable)
    let editBudget = budgets.find(b => b.id === budgetProgressData.budget?.id);
    
    if (!editBudget) {
      // Fallback: find by date comparison (normalize dates to YYYY-MM-DD format)
      const normalizeDate = (dateStr) => {
        if (!dateStr) return null;
        return new Date(dateStr).toISOString().split('T')[0];
      };
      
      const targetStart = normalizeDate(budgetProgressData.budget?.period_start);
      const targetEnd = normalizeDate(budgetProgressData.budget?.period_end);
      
      editBudget = budgets.find(b => {
        const budgetStart = normalizeDate(b.period_start);
        const budgetEnd = normalizeDate(b.period_end);
        return budgetStart === targetStart && budgetEnd === targetEnd;
      });
    }
    
    if (!editBudget) {
      console.error('Could not find budget for editing. Target dates:', {
        start: budgetProgressData.budget?.period_start,
        end: budgetProgressData.budget?.period_end,
        id: budgetProgressData.budget?.id
      });
      setError('Could not find budget for editing. Please refresh the page and try again.');
      return;
    }
    
    console.log('Found budget for editing:', editBudget);
    setEditingBudget(editBudget); 
    setShowForm(true); 
  };

  const [deleteTarget, setDeleteTarget] = useState(null);

  const openDeleteConfirm = (budgetProgressData) => {
    console.log('openDeleteConfirm called with:', budgetProgressData);
    console.log('Available budgets:', budgets);
    
    // Try to find budget by ID first (most reliable)
    let deleteBudget = budgets.find(b => b.id === budgetProgressData.budget?.id);
    
    if (!deleteBudget) {
      // Fallback: find by date comparison (normalize dates to YYYY-MM-DD format)
      const normalizeDate = (dateStr) => {
        if (!dateStr) return null;
        return new Date(dateStr).toISOString().split('T')[0];
      };
      
      const targetStart = normalizeDate(budgetProgressData.budget?.period_start);
      const targetEnd = normalizeDate(budgetProgressData.budget?.period_end);
      
      deleteBudget = budgets.find(b => {
        const budgetStart = normalizeDate(b.period_start);
        const budgetEnd = normalizeDate(b.period_end);
        return budgetStart === targetStart && budgetEnd === targetEnd;
      });
    }
    
    if (!deleteBudget) {
      console.error('Could not find budget for deletion. Target dates:', {
        start: budgetProgressData.budget?.period_start,
        end: budgetProgressData.budget?.period_end,
        id: budgetProgressData.budget?.id
      });
      setError('Could not find budget for deletion. Please refresh the page and try again.');
      return;
    }
    
    console.log('Found budget for deletion:', deleteBudget);
    setDeleteTarget(deleteBudget);
  };
  const cancelDelete = () => setDeleteTarget(null);

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    
    // Use the ID from the budgets array
    const budgetId = deleteTarget.id;
    
    if (!budgetId) {
      setError('Invalid budget ID');
      return;
    }

    try {
      await authFetch(`/api/budgets/${budgetId}`, { method: 'DELETE' });
      await fetchData();
      setDeleteTarget(null);
    } catch (err) {
      setError(err.message || 'Failed to delete');
    }
  };

    const onSaved = () => {
    fetchData();
    setShowForm(false);
    setEditingBudget(null);
  };

  const handleCategoryCreated = (newCategory) => {
    setCategories(prev => [...prev, newCategory]);
  };

  if (loading) return <div className="budget-container"><div className="loading">Loading budgets...</div></div>;

  return (
    <div className="budget-container">
      <header className="budget-header">
        <h1>Budgets</h1>
        <div className="header-actions">
          <button 
            className="refresh-btn" 
            onClick={fetchData}
            disabled={loading}
            title="Refresh budget data"
          >
            {loading ? '🔄' : '↻'} Refresh
          </button>
          <button className="add-budget-btn" onClick={() => { setEditingBudget(null); setShowForm(true); }}>
             Create Budget
          </button>
        </div>
      </header>

      {error && <div className="error-message">{error}</div>}

      {showForm && (
        <BudgetForm
          onClose={() => { setShowForm(false); setEditingBudget(null); }}
          onSaved={onSaved}
          editingBudget={editingBudget}
          categories={categories}
          onCategoryCreated={handleCategoryCreated}
        />
      )}

      {budgetProgress.length === 0 ? (
        <div className="empty-state">
          <h3>No budgets created yet</h3>
          <p>Create your first budget to start tracking your spending goals.</p>
          <button className="action-button" onClick={() => setShowForm(true)}>Create Your First Budget</button>
        </div>
      ) : (
        <div className="budgets-grid">
          {budgetProgress.map((budget, index) => (
            <BudgetProgressCard 
              key={budget.budget?.id || index} 
              budgetData={budget} 
              onEdit={handleEdit} 
              onDelete={openDeleteConfirm} 
            />
          ))}
        </div>
      )}

      {deleteTarget && (
        <Modal title="Confirm Delete" onClose={cancelDelete}>
          <p>Are you sure you want to delete this budget for {' '} 
            <strong>
              {(() => {
                // Handle both budget-progress and regular budget structures
                const startDate = deleteTarget.budget?.period_start || deleteTarget.period_start;
                const endDate = deleteTarget.budget?.period_end || deleteTarget.period_end;
                
                if (startDate && endDate) {
                  const start = new Date(startDate).toLocaleDateString();
                  const end = new Date(endDate).toLocaleDateString();
                  return `${start} - ${end}`;
                } else if (startDate) {
                  return new Date(startDate).toLocaleDateString();
                } else if (deleteTarget.id) {
                  return `ID: ${deleteTarget.id}`;
                } else {
                  return 'unknown period';
                }
              })()}
            </strong>?
          </p>
          <div className="form-actions">
            <button type="button" onClick={cancelDelete}>Cancel</button>
            <button type="button" className="danger" onClick={confirmDelete}>Delete</button>
          </div>
        </Modal>
      )}
    </div>
  );
}

export default Budget;