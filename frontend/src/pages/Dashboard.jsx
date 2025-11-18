import React, { useEffect, useState } from 'react';
import '../styles/Dashboard.css';
import { DashboardHeader, AccountBalanceCards, BudgetProgressSection, SpendingSummarySection, CashflowSection, RecentTransactionsSection, EmptyDashboardState } from '../components/dashboard'

function Dashboard() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [accountBalances, setAccountBalances] = useState(null);
  const [budgetProgress, setBudgetProgress] = useState(null);
  const [spendSummary, setSpendSummary] = useState(null);
  const [cashflow, setCashflow] = useState(null);
  const [recentTransactions, setRecentTransactions] = useState([]);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  // Listen for transaction changes to refresh dashboard data
  useEffect(() => {
    const handleTransactionChange = () => {
      console.log('Dashboard: Transaction change detected, refreshing data...');
      fetchDashboardData();
    };

    // Listen for custom events
    window.addEventListener('transactionCreated', handleTransactionChange);
    window.addEventListener('transactionUpdated', handleTransactionChange);
    window.addEventListener('transactionDeleted', handleTransactionChange);

    // Listen for storage changes (in case transactions are modified in another tab)
    window.addEventListener('storage', (e) => {
      if (e.key === 'lastTransactionUpdate') {
        console.log('Dashboard: Storage change detected for transactions');
        handleTransactionChange();
      }
    });

    console.log('Dashboard: Event listeners registered');

    return () => {
      window.removeEventListener('transactionCreated', handleTransactionChange);
      window.removeEventListener('transactionUpdated', handleTransactionChange);
      window.removeEventListener('transactionDeleted', handleTransactionChange);
      window.removeEventListener('storage', handleTransactionChange);
      console.log('Dashboard: Event listeners removed');
    };
  }, []);

  const fetchDashboardData = async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      setError('Not authenticated');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);

      // Fetch all dashboard data in parallel
      const headers = { 'Authorization': `Bearer ${token}` };

      const [balancesRes, budgetRes, spendRes, cashflowRes, transactionsRes] = await Promise.all([
        fetch('/api/reports/account-balances', { headers }).catch(() => ({ ok: false, status: 404 })),
        fetch('/api/reports/budget-progress', { headers }).catch(() => ({ ok: false, status: 404 })),
        fetch('/api/reports/spend-summary', { headers }).catch(() => ({ ok: false, status: 404 })),
        fetch('/api/reports/cashflow?group_by=month', { headers }).catch(() => ({ ok: false, status: 404 })),
        fetch('/api/transactions?limit=5', { headers }).catch(() => ({ ok: false, status: 404 }))
      ]);

      // Helper function to safely parse JSON
      const safeParseJson = async (response, fallback) => {
        if (!response.ok) return fallback;
        try {
          const text = await response.text();
          return text ? JSON.parse(text) : fallback;
        } catch (err) {
          console.warn('Failed to parse JSON response:', err);
          return fallback;
        }
      };

      // Handle responses - treat 404s as empty data rather than errors
      const balances = await safeParseJson(balancesRes, { accounts: [], total_balance_cents: 0 });
      const budget = await safeParseJson(budgetRes, null);
      const spend = await safeParseJson(spendRes, { categories: [], total_spent_cents: 0 });
      const cashflowData = await safeParseJson(cashflowRes, { periods: [], summary: { total_income_cents: 0, total_expense_cents: 0, net_cents: 0 } });
      const transactions = await safeParseJson(transactionsRes, []);

      setAccountBalances(balances);
      setBudgetProgress(budget);
      setSpendSummary(spend);
      setCashflow(cashflowData);
      setRecentTransactions(transactions);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (cents) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(cents / 100);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="dashboard-container">
        <div className="loading">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard-container">
        <div className="error-message">{error}</div>
        <button onClick={fetchDashboardData}>Retry</button>
      </div>
    );
  }

  // Check if user has any data
  const hasAnyData = (
    (accountBalances?.accounts?.length > 0) ||
    (budgetProgress !== null) ||
    (spendSummary?.categories?.length > 0) ||
    (cashflow?.periods?.length > 0) ||
    (recentTransactions?.length > 0)
  );

  if (!loading && !hasAnyData) {
    return (
      <div className="dashboard-container">
        <header className="dashboard-header">
          <h1>Welcome to Your Financial Dashboard</h1>
        </header>
        <EmptyDashboardState />
      </div>
    );
  }

  return (
    <div className="dashboard-container">
      <DashboardHeader onRefresh={fetchDashboardData} />

      <AccountBalanceCards 
        accountBalances={accountBalances}
        formatCurrency={formatCurrency}
      />

      <BudgetProgressSection 
        budgetProgress={budgetProgress}
        formatCurrency={formatCurrency}
        formatDate={formatDate}
      />

      <SpendingSummarySection 
        spendSummary={spendSummary}
        formatCurrency={formatCurrency}
      />

      <CashflowSection 
        cashflow={cashflow}
        formatCurrency={formatCurrency}
      />

      <RecentTransactionsSection 
        recentTransactions={recentTransactions}
        formatCurrency={formatCurrency}
        formatDate={formatDate}
      />
    </div>
  );
}

export default Dashboard;