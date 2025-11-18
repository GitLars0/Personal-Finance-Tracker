import React from 'react';
import CategoryTree from './CategoryTree';

function CategoryList({ 
  filteredCategories, 
  filter, 
  resetForm, 
  setFormData, 
  setShowForm,
  groupedCategories,
  handleEdit,
  handleDelete 
}) {
  if (filteredCategories.length === 0) {
    return (
      <div className="no-categories">
        <p>No {filter === 'all' ? '' : filter} categories found.</p>
        <button 
          className="action-button"
          onClick={() => {
            resetForm();
            if (filter !== 'all') {
              setFormData(prev => ({ ...prev, kind: filter }));
            }
            setShowForm(true);
          }}
        >
          Create {filter === 'all' ? '' : filter} Category
        </button>
      </div>
    );
  }

  return (
    <CategoryTree 
      groupedCategories={groupedCategories}
      handleEdit={handleEdit}
      handleDelete={handleDelete}
    />
  );
}

export default CategoryList;