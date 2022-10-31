import { check } from "k6";
import http from "k6/http";

export const options = {
  vus: 10,
  duration: "2m",
};

export default function () {
  const res = http.get(__ENV.LOAD_TEST_URI);
  check(res, {
    "is status 200": (r) => r.status === 200,
  });
}
