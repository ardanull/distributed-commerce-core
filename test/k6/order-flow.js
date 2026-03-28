import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "30s",
};

export default function () {
  const payload = JSON.stringify({
    customer_id: `cust-${__VU}`,
    currency: "TRY",
    items: [{ sku: "keyboard", quantity: 1, unit_price: 249900 }],
  });

  const res = http.post("http://localhost:8080/v1/orders", payload, {
    headers: {
      "Content-Type": "application/json",
      "X-Correlation-ID": `k6-${__VU}-${__ITER}`,
    },
  });

  check(res, {
    "accepted": (r) => r.status === 202,
  });

  sleep(1);
}
