import { z } from "zod";

const serverEnvSchema = z.object({
  BEEBOX_IDENTITY_URL: z.string().url().default("http://localhost:8080"),
  SESSION_COOKIE_NAME: z.string().default("beebox_session"),
  SESSION_COOKIE_SECURE: z.string().optional().transform((v) => v === "true" || v === "1"),
  NODE_ENV: z.enum(["development", "production", "test"]).default("development"),
});

export function getServerEnv() {
  const parsed = serverEnvSchema.safeParse({
    BEEBOX_IDENTITY_URL: process.env.BEEBOX_IDENTITY_URL,
    SESSION_COOKIE_NAME: process.env.SESSION_COOKIE_NAME,
    SESSION_COOKIE_SECURE: process.env.SESSION_COOKIE_SECURE,
    NODE_ENV: process.env.NODE_ENV,
  });
  if (!parsed.success) throw new Error("Invalid server environment");
  return parsed.data;
}
