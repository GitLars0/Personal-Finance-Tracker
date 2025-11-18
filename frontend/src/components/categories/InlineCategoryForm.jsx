import React, { useState } from 'react';
import useAuthFetch from '../../hooks/useAuthFetch';
import '../../styles/InlineCategory.css';

function InlineCategoryForm({ onCategoryCreated, onCancel, defaultKind = 'expense', allowedKinds = ['expense', 'income'] }) {
  const authFetch = useAuthFetch();
  const [formData, setFormData] = useState({
    name: '',
    kind: defaultKind,
    description: ''
  });
  const [error, setError] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    console.log('InlineCategoryForm: form submitted');
    
    if (!formData.name.trim()) {
      setError('Category name is required');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      console.log('InlineCategoryForm: creating category with data:', formData);

      const newCategory = await authFetch('/api/categories', {
        method: 'POST',
        body: JSON.stringify({
          name: formData.name.trim(),
          kind: formData.kind,
          description: formData.description.trim() || null,
          parent_id: null
        })
      });
      
      console.log('InlineCategoryForm: created category:', newCategory);
      onCategoryCreated(newCategory);
      
      // Reset form
      setFormData({
        name: '',
        kind: defaultKind,
        description: ''
      });
    } catch (err) {
      console.error('InlineCategoryForm: error creating category:', err);
      if (err && err.status === 409) {
        setError('A category with this name already exists. Please choose a different name.');
      } else if (err && err.message) {
        setError(err.message);
      } else {
        setError('Failed to create category');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="inline-category-form">
      <div className="inline-form-header">
        <h4>Create New Category</h4>
        <button type="button" onClick={onCancel} className="close-btn">×</button>
      </div>
      
      {error && <div className="error-message">{error}</div>}
      
      <div className="compact-form">
        <div className="form-row">
          <div className="form-group">
            <label>Name *</label>
            <input
              type="text"
              placeholder="e.g., Groceries, Salary"
              value={formData.name}
              onChange={(e) => setFormData({...formData, name: e.target.value})}
              required
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  handleSubmit(e);
                }
              }}
            />
          </div>
          
          <div className="form-group">
            <label>Type *</label>
            <select
              value={formData.kind}
              onChange={(e) => setFormData({...formData, kind: e.target.value})}
              required
              disabled={allowedKinds.length === 1}
            >
              {allowedKinds.includes('expense') && <option value="expense">Expense</option>}
              {allowedKinds.includes('income') && <option value="income">Income</option>}
            </select>
          </div>
        </div>
        
        <div className="form-group">
          <label>Description (optional)</label>
          <input
            type="text"
            placeholder="Brief description"
            value={formData.description}
            onChange={(e) => setFormData({...formData, description: e.target.value})}
          />
        </div>
        
        <div className="form-actions">
          <button type="button" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </button>
          <button 
            type="button" 
            className="primary" 
            disabled={isSubmitting}
            onClick={handleSubmit}
          >
            {isSubmitting ? 'Creating...' : 'Create Category'}
          </button>
        </div>
      </div>
    </div>
  );
}

export default InlineCategoryForm;