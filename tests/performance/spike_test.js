import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '10s', target: 10 },    // Normal load
    { duration: '1m', target: 1000 },   // Sudden spike
    { duration: '3m', target: 1000 },   // Sustain spike
    { duration: '10s', target: 10 },    // Back to normal
    { duration: '3m', target: 0 },      // Recovery
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2000'], // Allow higher latency
    'errors': ['rate<0.10'],              // 10% error tolerance
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
    'login works': (r) => r.status === 200 || r.status === 429, // Allow rate limiting
  });

  // Track login errors (429 rate limiting counts as success for spike test)
  errorRate.add(!loginSuccess ? 1 : 0);

  if (loginRes.status === 200 && loginRes.json('token')) {
    const token = loginRes.json('token');
    
    const dashboardRes = http.get(`${BASE_URL}/api/reports/account-balances`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });

    // Track dashboard request errors
    const dashboardSuccess = check(dashboardRes, {
      'dashboard loaded': (r) => r.status === 200 || r.status === 429,
    });
    errorRate.add(!dashboardSuccess ? 1 : 0);
  } else if (loginRes.status !== 429) {
    // If login failed (and not rate limited), count dashboard as error
    errorRate.add(1);
  }

  sleep(0.1);
}