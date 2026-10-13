import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom Metrics for Capacity Invariant Verification
const confirmedRsvps = new Counter('confirmed_rsvps');
const rejectedConflicts = new Counter('rejected_conflicts');
const unexpectedErrors = new Counter('unexpected_errors');
const rsvpLatency = new Trend('rsvp_duration_ms');

export const options = {
  scenarios: {
    ticket_drop_spike: {
      executor: 'per-vu-iterations',
      vus: 100, // 100 concurrent virtual users competing simultaneously
      iterations: 10, // 10 iterations each -> 1,000 total requests
      maxDuration: '30s',
    },
  },
  thresholds: {
    'unexpected_errors': ['count==0'], // Zero 5xx errors permitted
    'rsvp_duration_ms': ['p(95)<350'],  // Sub-350ms p95 latency invariant
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080/api/v1';
const EVENT_ID = __ENV.EVENT_ID || '00000000-0000-0000-0000-000000000000';
const EVENT_CAPACITY = parseInt(__ENV.EVENT_CAPACITY || '200', 10);

export default function (data) {
  const vuId = __VU;
  const iterId = __ITER;
  const timestamp = Date.now();

  const payload = JSON.stringify({
    event_id: EVENT_ID,
    guest_name: `Attendee VU-${vuId}-Iter-${iterId}`,
    guest_email: `loadtest_${timestamp}_vu${vuId}_iter${iterId}@hackways.in`,
    notes: 'k6 concurrency capacity stress verification',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    timeout: '2500ms', // Must fall within 2000ms statement_timeout budget
  };

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/rsvps`, payload, params);
  const duration = Date.now() - startTime;
  rsvpLatency.add(duration);

  if (res.status === 201) {
    confirmedRsvps.add(1);
  } else if (res.status === 409) {
    rejectedConflicts.add(1);
  } else {
    unexpectedErrors.add(1);
    console.error(`Unexpected response HTTP ${res.status}: ${res.body}`);
  }

  check(res, {
    'is expected status code (201 or 409)': (r) => r.status === 201 || r.status === 409,
    'response under 2500ms': () => duration < 2500,
  });

  // Short jitter sleep between iterations
  sleep(0.05);
}

export function handleSummary(data) {
  const confirmed = data.metrics.confirmed_rsvps ? data.metrics.confirmed_rsvps.values.count : 0;
  const conflicts = data.metrics.rejected_conflicts ? data.metrics.rejected_conflicts.values.count : 0;
  const errors = data.metrics.unexpected_errors ? data.metrics.unexpected_errors.values.count : 0;
  const oversold = Math.max(0, confirmed - EVENT_CAPACITY);

  console.log(`\n========================================================`);
  console.log(`HACKWAYS CONCURRENCY LOAD TEST SUMMARY`);
  console.log(`========================================================`);
  console.log(`Target Event Capacity: ${EVENT_CAPACITY}`);
  console.log(`Confirmed RSVPs (201): ${confirmed}`);
  console.log(`Capacity Conflicts (409): ${conflicts}`);
  console.log(`Unexpected Errors: ${errors}`);
  console.log(`Oversold Tickets: ${oversold}`);
  console.log(`P95 Latency: ${data.metrics.rsvp_duration_ms ? data.metrics.rsvp_duration_ms.values['p(95)'].toFixed(2) : 'N/A'}ms`);
  console.log(`========================================================\n`);

  return {
    stdout: JSON.stringify({
      target_capacity: EVENT_CAPACITY,
      confirmed_rsvps: confirmed,
      conflicts: conflicts,
      unexpected_errors: errors,
      tickets_oversold: oversold,
      invariant_passed: oversold === 0 && errors === 0,
    }, null, 2),
  };
}
