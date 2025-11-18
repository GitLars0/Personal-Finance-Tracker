import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up
    { duration: '5m', target: 100 },  // Stress level
    { duration: '5m', target: 200 },  // Beyond normal
    { duration: '2m', target: 0 },    // Recovery
  ],
  thresholds: {
    'http_req_duration': ['p(99)<1000'], // 99% under 1s
    'errors': ['rate<0.05'],              // 5% error tolerance
  },
};

const BASE_URL = 'http://localhost:8080';

const testUser = {
  username: 'demo',
  password: 'demo123'
};

export default function () {
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });

  const loginSuccess = check(loginRes, {
    'login status 200': (r) => r.status === 200,
  });

  errorRate.add(!loginSuccess ? 1 : 0);

  if (loginSuccess) {
    const token = loginRes.json('token');
    const authHeaders = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    };

    // Stress test multiple endpoints
    const responses = http.batch([
      ['GET', `${BASE_URL}/api/reports/account-balances`, null, { headers: authHeaders }],
      ['GET', `${BASE_URL}/api/transactions`, null, { headers: authHeaders }],
      ['GET', `${BASE_URL}/api/budgets`, null, { headers: authHeaders }],
      ['GET', `${BASE_URL}/api/accounts`, null, { headers: authHeaders }],
    ]);

    // Track errors for each batched request
    responses.forEach((res) => {
      const success = check(res, {
        'status is 200': (r) => r.status === 200,
      });
      errorRate.add(!success ? 1 : 0);
    });
  } else {
    // If login failed, count the 4 batch requests as errors too
    errorRate.add(1); // account-balances
    errorRate.add(1); // transactions
    errorRate.add(1); // budgets
    errorRate.add(1); // accounts
  }

  sleep(0.5);
}