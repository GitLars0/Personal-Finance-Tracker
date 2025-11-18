import React from 'react';
import { Navigate } from 'react-router-dom';

// Admin Route Component - requires both authentication and admin role
function AdminRoute({ children }) {
  const token = localStorage.getItem('token');
  const user = JSON.parse(localStorage.getItem('user') || '{}');
  
  if (!token) {
    return <Navigate to="/login" />;
  }
  
  if (user.role !== 'admin') {
    return <Navigate to="/dashboard" />;
  }
  
  return children;
}

export default AdminRoute;