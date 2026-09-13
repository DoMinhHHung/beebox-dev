import { z } from "zod";

const serverEnvSchema = z.object({
  BEEBOX_IDENTITY_URL: z.string().min(1).default("http://127.0.0.1:8081"),
  SESSION_COOKIE_NAME: z.string().default("beebox_session"),
  SESSION_COOKIE_SECURE: z
    .string()
    .optional()
    .transform((v) => v === "true" || v === "1"),
  NODE_ENV: z
    .enum(["development", "production", "test"])
    .default("development"),
});

export type ServerEnv = z.infer<typeof serverEnvSchema>;

export function getServerEnv(): ServerEnv {
  const parsed = serverEnvSchema.safeParse({
    BEEBOX_IDENTITY_URL: process.env.BEEBOX_IDENTITY_URL,
    SESSION_COOKIE_NAME: process.env.SESSION_COOKIE_NAME,
    SESSION_COOKIE_SECURE: process.env.SESSION_COOKIE_SECURE,
    NODE_ENV: process.env.NODE_ENV,
  });

  if (!parsed.success) {
    console.error("[env] invalid server environment:", parsed.error.flatten());
    throw new Error("Invalid server environment configuration");
  }

  // Normalize: allow host without protocol by mistake
  let url = parsed.data.BEEBOX_IDENTITY_URL.trim();
  if (!/^https?:\/\//i.test(url)) {
    url = `http://${url}`;
  }

  return { ...parsed.data, BEEBOX_IDENTITY_URL: url };
}
