import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Transactions from './pages/Transactions';
import Budget from './pages/Budget';
import Accounts from './pages/Accounts';
import Categories from './pages/Categories';
import Profile from './pages/Profile';
import Login from './pages/Login';
import Register from './pages/Register';
import Home from './pages/Home';
import Admin from './pages/Admin';
import BankConnections from './pages/BankConnections';
import BankCallback from './pages/BankCallback';
import { UserList } from './components/admin/users';
import { AllTransactions, AllAccounts, AllCategories, AllBudgets } from './components/admin/data';
import './styles/App.css';
import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import Footer from './components/Footer';
import AdminRoute from './AdminRoute';

// Protected Route Component
function ProtectedRoute({ children }) {
  const token = localStorage.getItem('token');
  return token ? children : <Navigate to="/login" />;
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    // Check if user is logged in
    const token = localStorage.getItem('token');
    setIsAuthenticated(!!token);
  }, []);

  // Listen for storage changes (login/logout events)
  useEffect(() => {
    const handleStorageChange = () => {
      const token = localStorage.getItem('token');
      setIsAuthenticated(!!token);
    };

    window.addEventListener('storage', handleStorageChange);
    
    // Custom event for same-tab login/logout
    window.addEventListener('authChange', handleStorageChange);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      window.removeEventListener('authChange', handleStorageChange);
    };
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setIsAuthenticated(false);
    // Dispatch custom event to notify other components
    window.dispatchEvent(new Event('authChange'));
    window.location.href = '/';
  };

  return (
    <Router>
      <div className="app-wrapper">
        <Navbar isAuthenticated={isAuthenticated} onLogout={handleLogout} />
        <Sidebar isAuthenticated={isAuthenticated} />

        {/* Routes */}
        <div className="app-content">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route 
              path="/dashboard" 
              element={
                <ProtectedRoute>
                  <Dashboard />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/transactions" 
              element={
                <ProtectedRoute>
                  <Transactions />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/budgets" 
              element={
                <ProtectedRoute>
                  <Budget />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/accounts" 
              element={
                <ProtectedRoute>
                  <Accounts />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/categories" 
              element={
                <ProtectedRoute>
                  <Categories />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/profile" 
              element={
                <ProtectedRoute>
                  <Profile />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/banks" 
              element={
                <ProtectedRoute>
                  <BankConnections />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/banks/callback" 
              element={
                <ProtectedRoute>
                  <BankCallback />
                </ProtectedRoute>
              } 
            />
            
            {/* Admin Routes */}
            <Route 
              path="/admin" 
              element={
                <AdminRoute>
                  <Admin />
                </AdminRoute>
              } 
            />
            <Route 
              path="/admin/users" 
              element={
                <AdminRoute>
                  <UserList />
                </AdminRoute>
              } 
            />
            <Route 
              path="/admin/transactions" 
              element={
                <AdminRoute>
                  <AllTransactions />
                </AdminRoute>
              } 
            />
            <Route 
              path="/admin/accounts" 
              element={
                <AdminRoute>
                  <AllAccounts />
                </AdminRoute>
              } 
            />
            <Route 
              path="/admin/categories" 
              element={
                <AdminRoute>
                  <AllCategories />
                </AdminRoute>
              } 
            />
            <Route 
              path="/admin/budgets" 
              element={
                <AdminRoute>
                  <AllBudgets />
                </AdminRoute>
              } 
            />
            
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
          </Routes>
        </div>
        <Footer />
      </div>
    </Router>
  );
}

export default App;