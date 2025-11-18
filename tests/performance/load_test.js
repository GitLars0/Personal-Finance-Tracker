import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '2m', target: 10 },  // Ramp up to 10 users
    { duration: '5m', target: 50 },  // Stay at 50 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests under 500ms
    'errors': ['rate<0.01'],             // Error rate under 1%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test user credentials
const testUser = {
  username: 'demo',
  password: 'demo123'
};

export default function () {
  // Login
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });

  const loginSuccess = check(loginRes, {
    'login successful': (r) => r.status === 200,
    'received token': (r) => r.json('token') !== undefined,
  });

  // Track login errors
  errorRate.add(!loginSuccess ? 1 : 0);

  if (loginSuccess) {
    const token = loginRes.json('token');
    const authHeaders = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    };

    // Get dashboard data
    const dashboardRes = http.get(`${BASE_URL}/api/reports/account-balances`, {
      headers: authHeaders,
    });

    const dashboardSuccess = check(dashboardRes, {
      'dashboard loaded': (r) => r.status === 200,
    });
    errorRate.add(!dashboardSuccess ? 1 : 0); // Track dashboard errors

    // Get transactions
    const txnRes = http.get(`${BASE_URL}/api/transactions`, {
      headers: authHeaders,
    });

    const txnSuccess = check(txnRes, {
      'transactions loaded': (r) => r.status === 200,
    });
    errorRate.add(!txnSuccess ? 1 : 0); // Track transaction errors

    // Get budgets
    const budgetRes = http.get(`${BASE_URL}/api/budgets`, {
      headers: authHeaders,
    });

    const budgetSuccess = check(budgetRes, {
      'budgets loaded': (r) => r.status === 200,
    });
    errorRate.add(!budgetSuccess ? 1 : 0); // Track budget errors
  } else {
    // If login failed, count the remaining requests as errors too
    errorRate.add(1); // dashboard
    errorRate.add(1); // transactions
    errorRate.add(1); // budgets
  }

  sleep(1);
}