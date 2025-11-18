import React from 'react';

function CategoryForm({ 
  showForm, 
  editingCategory, 
  formData, 
  setFormData, 
  handleFormSubmit, 
  resetForm, 
  setShowForm, 
  setEditingCategory,
  categories 
}) {
  if (!showForm) return null;

  return (
    <div className="form-overlay">
      <div className="category-form">
        <h2>{editingCategory ? 'Edit Category' : 'Add New Category'}</h2>
        
        <form onSubmit={handleFormSubmit}>
          <div className="form-group">
            <label>Category Name *</label>
            <input
              type="text"
              placeholder="e.g., Groceries, Salary, Entertainment"
              value={formData.name}
              onChange={(e) => setFormData({...formData, name: e.target.value})}
              required
            />
          </div>

          <div className="form-group">
            <label>Category Type *</label>
            <select
              value={formData.kind}
              onChange={(e) => setFormData({...formData, kind: e.target.value})}
              required
            >
              <option value="expense">Expense</option>
              <option value="income">Income</option>
            </select>
            <small className="form-help">
              Income: Money coming in • Expense: Money going out
            </small>
          </div>

          <div className="form-group">
            <label>Parent Category</label>
            <select
              value={formData.parent_id}
              onChange={(e) => setFormData({...formData, parent_id: e.target.value})}
            >
              <option value="">No parent (top-level category)</option>
              {categories
                .filter(cat => cat.kind === formData.kind && cat.id !== editingCategory?.id)
                .map(category => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
            </select>
            <small className="form-help">
              Optional: Create subcategories by selecting a parent
            </small>
          </div>

          <div className="form-group">
            <label>Description</label>
            <textarea
              placeholder="Optional description for this category"
              value={formData.description}
              onChange={(e) => setFormData({...formData, description: e.target.value})}
              rows={3}
            />
          </div>

          <div className="form-actions">
            <button type="button" onClick={() => {
              setShowForm(false);
              setEditingCategory(null);
              resetForm();
            }}>
              Cancel
            </button>
            <button type="submit" className="primary">
              {editingCategory ? 'Update Category' : 'Add Category'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CategoryForm;