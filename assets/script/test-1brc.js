import { check, sleep } from 'k6';
import { FormData } from 'https://jslib.k6.io/formdata/0.0.2/index.js';
import http from 'k6/http';

const fileData = open('../sample/measurements-500.txt');

export const options = {
  discardResponseBodies: true,
  scenarios: {
    contacts: {
      executor: 'per-vu-iterations',
      vus: 100,
      iterations: 10,
      maxDuration: '10m',
    },
  },
};

export function setup() {
    const baseUrl = 'http://localhost:8080';
    return {
        baseUrl: baseUrl,
    };
}

export default function (data) {
    const url = `${data.baseUrl}/one-billion-row-challenge`;
    const formData1 = new FormData();
    formData1.append('file', http.file(fileData, 'measurements.txt', 'application/txt'));
    const res = http.post(url, formData1.body(), {
        headers: {
            'Content-Type': 'multipart/form-data; boundary=' + formData1.boundary,
        },
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
        // 'response has body': (r) => r.body && r.body.length > 0,
    });
    sleep(0.2);
}
