import React, { useState, useEffect } from 'react';
import Modal from '../Modal';
import useAuthFetch from '../../hooks/useAuthFetch';
import CategorySelector from '../categories/CategorySelector';
import AiBudgetSuggestions from './AiBudgetSuggestions';

export default function BudgetForm({ onClose, onSaved, editingBudget, categories, onCategoryCreated }) {
  const authFetch = useAuthFetch();
  const [formData, setFormData] = useState({ start_date: new Date().toISOString().split('T')[0], end_date: '', items: [{ category_id: '', budget_cents: '' }] });
  const [error, setError] = useState(null);
  const [currentCategories, setCurrentCategories] = useState(categories);

  useEffect(() => {
    setCurrentCategories(categories);
  }, [categories]);

  const handleCategoryCreated = (newCategory) => {
    setCurrentCategories(prev => [...prev, newCategory]);
    if (onCategoryCreated) {
      onCategoryCreated(newCategory);
    }
  };

  useEffect(() => {
    if (editingBudget) {
      setFormData({
        start_date: editingBudget.period_start?.split('T')[0] || '',
        end_date: editingBudget.period_end?.split('T')[0] || '',
        items: editingBudget.items?.map(i => ({ category_id: i.category_id?.toString() || '', budget_cents: (i.planned_cents / 100).toString() })) || [{ category_id: '', budget_cents: '' }]
      });
    }
  }, [editingBudget]);

  // helper to reset form when needed (inline usage preferred to avoid linter unused warning)
  const handleApplyAiSuggestion = (suggestion) => {
    setFormData(prevData => {
      // Find if category already exists in form items
      const existingItemIndex = prevData.items.findIndex(
        item => item.category_id === suggestion.category_id.toString()
      );

      if (existingItemIndex !== -1) {
        // Update existing item
        const updatedItems = [...prevData.items];
        updatedItems[existingItemIndex].budget_cents = suggestion.suggested_amount.toString();
        return { ...prevData, items: updatedItems };
      } else {
        // Add new item (but first check if we have an empty item to replace)
        const emptyItemIndex = prevData.items.findIndex(
          item => !item.category_id && !item.budget_cents
        );
        
        const newItem = {
          category_id: suggestion.category_id.toString(),
          budget_cents: suggestion.suggested_amount.toString()
        };

        if (emptyItemIndex !== -1) {
          // Replace the first empty item
          const updatedItems = [...prevData.items];
          updatedItems[emptyItemIndex] = newItem;
          return { ...prevData, items: updatedItems };
        } else {
          // Add as new item
          return { ...prevData, items: [...prevData.items, newItem] };
        }
      }
    });
  };

  const getTargetMonth = () => {
    if (formData.start_date) {
      const startDate = new Date(formData.start_date);
      return startDate.getMonth() + 1; // JS months are 0-indexed
    }
    return new Date().getMonth() + 1;
  };

  const getTargetYear = () => {
    if (formData.start_date) {
      const startDate = new Date(formData.start_date);
      return startDate.getFullYear();
    }
    return new Date().getFullYear();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    try {
      const validItems = formData.items.filter(it => {
        // Must have a category selected
        if (!it.category_id) return false;
        
        // Must have a valid amount (can be 0 or positive, empty defaults to 0)
        const budgetCents = it.budget_cents;
        const amount = budgetCents === '' || budgetCents == null ? 0 : parseFloat(budgetCents);
        return !isNaN(amount) && amount >= 0;
      });
      
      if (validItems.length === 0) throw new Error('Please add at least one budget item with a category and amount');

      const payload = {
        period_start: formData.start_date,
        period_end: formData.end_date,
        currency: 'USD',
        items: validItems.map(item => {
          const budgetCents = item.budget_cents;
          const amount = budgetCents === '' || budgetCents == null ? 0 : parseFloat(budgetCents);
          return {
            category_id: parseInt(item.category_id), 
            planned_cents: Math.round(amount * 100)
          };
        })
      };

      if (editingBudget) {
        await authFetch(`/api/budgets/${editingBudget.id}`, { method: 'PUT', body: JSON.stringify(payload) });
      } else {
        await authFetch('/api/budgets', { method: 'POST', body: JSON.stringify(payload) });
      }

      onSaved && onSaved();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to save budget');
    }
  };

  return (
    <Modal title={editingBudget ? 'Edit Budget' : 'Create New Budget'} onClose={onClose}>
      {error && <div className="error-message">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label>Start Date *</label>
            <input type="date" value={formData.start_date} onChange={e => setFormData({...formData, start_date: e.target.value})} required />
          </div>
          <div className="form-group">
            <label>End Date *</label>
            <input type="date" value={formData.end_date} onChange={e => setFormData({...formData, end_date: e.target.value})} required />
          </div>
        </div>

        {/* AI Budget Suggestions */}
        <AiBudgetSuggestions
          onApplySuggestion={handleApplyAiSuggestion}
          targetMonth={getTargetMonth()}
          targetYear={getTargetYear()}
          disabled={editingBudget} // Only show for new budgets
        />

        <div className="budget-items-section">
          <h4>Budget Items</h4>
          {formData.items.map((item, idx) => (
            <div key={idx} className="budget-item-form">
              <div className="budget-category-selector">
                <CategorySelector
                  value={item.category_id}
                  onChange={(value) => {
                    const copy = [...formData.items]; 
                    copy[idx].category_id = value; 
                    setFormData({...formData, items: copy});
                  }}
                  categories={currentCategories.filter(cat => cat.kind === 'expense')}
                  onCategoryCreated={handleCategoryCreated}
                  label={`Category ${idx + 1}`}
                  defaultKind="expense"
                  allowedKinds={['expense']}
                  required
                />
              </div>
              <input type="number" step="0.01" placeholder="0.00" value={item.budget_cents} onChange={e => {
                const copy = [...formData.items]; copy[idx].budget_cents = e.target.value; setFormData({...formData, items: copy});
              }} />
              {formData.items.length > 1 && <button type="button" className="remove-item-btn" onClick={() => { setFormData({...formData, items: formData.items.filter((_, i) => i !== idx)}) }}>Remove</button>}
            </div>
          ))}
          <button type="button" className="add-item-btn" onClick={() => setFormData({...formData, items: [...formData.items, { category_id: '', budget_cents: '' }]})}>Add Category</button>
        </div>

        <div className="form-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button type="submit" className="primary">{editingBudget ? 'Update Budget' : 'Create Budget'}</button>
        </div>
      </form>
    </Modal>
  );
}
