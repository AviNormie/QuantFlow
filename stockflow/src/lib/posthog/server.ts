import { PostHog } from "posthog-node";

export function getPostHogServerClient(): PostHog | null {
  const apiKey =
    process.env.POSTHOG_API_KEY ??
    process.env.NEXT_PUBLIC_POSTHOG_PROJECT_TOKEN;

  if (!apiKey) {
    return null;
  }

  return new PostHog(apiKey, {
    host:
      process.env.POSTHOG_HOST ??
      process.env.NEXT_PUBLIC_POSTHOG_HOST ??
      "https://us.i.posthog.com",
    flushAt: 1,
    flushInterval: 0,
  });
}
