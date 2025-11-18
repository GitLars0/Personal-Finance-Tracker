import React from 'react';

function CategoryItem({ category, level, handleEdit, handleDelete }) {
  return (
    <div className={`category-item level-${level}`}>
      <div className="category-info">
        <div className="category-main">
          <h4 style={{ marginLeft: level * 20 }}>
            {level > 0 && '└─ '}
            {category.name}
          </h4>
          <span className={`category-kind ${category.kind}`}>
            {category.kind}
          </span>
        </div>
        {category.description && (
          <p className="category-description">{category.description}</p>
        )}
      </div>
      
      <div className="category-actions">
        <button 
          onClick={() => handleEdit(category)}
          className="edit-btn"
          title="Edit category"
        >
          ✏️
        </button>
        <button 
          onClick={() => handleDelete(category)}
          className="delete-btn"
          title="Delete category"
        >
          🗑️
        </button>
      </div>
    </div>
  );
}

export default CategoryItem;