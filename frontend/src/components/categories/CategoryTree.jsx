import React from 'react';
import CategoryItem from './CategoryItem';

function CategoryTree({ groupedCategories, handleEdit, handleDelete }) {
  const renderCategoryTree = (parentId = 'root', level = 0) => {
    const categoriesAtLevel = groupedCategories[parentId] || [];
    
    return categoriesAtLevel.map(category => (
      <div key={category.id}>
        <CategoryItem 
          category={category}
          level={level}
          handleEdit={handleEdit}
          handleDelete={handleDelete}
        />
        
        {/* Render subcategories */}
        {renderCategoryTree(category.id, level + 1)}
      </div>
    ));
  };

  return (
    <div className="category-tree">
      {renderCategoryTree()}
    </div>
  );
}

export default CategoryTree;