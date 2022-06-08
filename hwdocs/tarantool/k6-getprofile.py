import http from 'k6/http';
import { check, group, sleep } from 'k6';

export const options = {
  vus: 1,
//  duration: '5m',
};


//export const options = {
//  stages: [
//    { duration: '5m', target: 100 }, // simulate ramp-up of traffic from 1 to 100 users over 5 minutes.
//    { duration: '10m', target: 100 }, // stay at 100 users for 10 minutes
//    { duration: '5m', target: 0 }, // ramp-down to 0 users
//  ]
//};


const BASE_URL = 'http://127.0.0.1:9080';
const USERNAME = 'u1';
const PASSWORD = 'u1';

export default () => {
  const loginRes = http.post(`${BASE_URL}/login/`, {
    login: USERNAME,
    password: PASSWORD,
  });

  check(loginRes, {
    'logged in successfully': (resp) => resp.status !== 302,
  });

  const authHeaders = {
    headers: {
      Authorization: `Bearer ${loginRes.json('access')}`,
    },
  };


  for (let id = 1; id <= 50000; id++) {
    const profileRes = http.get(`${BASE_URL}/profile/?id=${id}`, {
      cookies: loginRes.cookies,
      tags: { name: 'GetProfileURL' },
    });
    check(profileRes, { 'retrieved profile': (resp) => resp.status == 200 });
  }
};
