import React, { useState } from 'react';
import InlineCategoryForm from './InlineCategoryForm';

function CategorySelector({ 
  value, 
  onChange, 
  categories, 
  onCategoryCreated, 
  required = false, 
  label = "Category",
  defaultKind = 'expense',
  allowedKinds = ['expense', 'income']
}) {
  const [showCreateForm, setShowCreateForm] = useState(false);

  const handleCategoryCreated = (newCategory) => {
    console.log('CategorySelector: handling new category:', newCategory);
    onCategoryCreated(newCategory);
    onChange(newCategory.id.toString()); // Auto-select the new category
    setShowCreateForm(false);
  };

  if (showCreateForm) {
    console.log('CategorySelector: showing create form');
    return (
      <div className="form-group">
        <label>{label} {required && '*'}</label>
        <InlineCategoryForm 
          onCategoryCreated={handleCategoryCreated}
          onCancel={() => setShowCreateForm(false)}
          defaultKind={defaultKind}
          allowedKinds={allowedKinds}
        />
      </div>
    );
  }

  return (
    <div className="form-group">
      <label>{label} {required && '*'}</label>
      <div className="category-selector-wrapper">
        <select
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={required}
        >
          <option value="">Select Category</option>
          {categories.map(category => (
            <option key={category.id} value={category.id}>
              {category.name} ({category.kind})
            </option>
          ))}
        </select>
        
        <button 
          type="button" 
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            console.log('+ New button clicked!');
            setShowCreateForm(true);
          }}
          className="create-category-btn"
          title="Create new category"
        >
          + New
        </button>
      </div>
      
      {categories.length === 0 && (
        <small className="form-help">
          No categories found. <button type="button" onClick={() => setShowCreateForm(true)} className="link-btn">Create your first category</button>
        </small>
      )}
    </div>
  );
}

export default CategorySelector;