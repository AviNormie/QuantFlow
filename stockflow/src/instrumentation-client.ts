import * as Sentry from "@sentry/nextjs";
import posthog from "posthog-js";

const tracesSampleRate = Number(
  process.env.SENTRY_TRACES_SAMPLE_RATE ??
    (process.env.NODE_ENV === "development" ? "1.0" : "0.1"),
);

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.SENTRY_ENVIRONMENT ?? process.env.NODE_ENV,
  tracesSampleRate,
});

const posthogKey = process.env.NEXT_PUBLIC_POSTHOG_PROJECT_TOKEN;
if (posthogKey) {
  posthog.init(posthogKey, {
    api_host:
      process.env.NEXT_PUBLIC_POSTHOG_HOST ?? "https://us.i.posthog.com",
    defaults: "2026-01-30",
  });
}

export const onRouterTransitionStart = Sentry.captureRouterTransitionStart;
