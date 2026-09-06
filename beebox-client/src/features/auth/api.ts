import { apiRequest, setAccessToken } from "@/shared/api/http";

export async function signUp(email: string, password: string) {
  return apiRequest<{ id: string; email: string }>("/auth/sign-up", {
    method: "POST",
    body: { email, password },
    auth: false,
  });
}

export async function signIn(email: string, password: string) {
  const res = await apiRequest<{ token: string }>("/auth/sign-in", {
    method: "POST",
    body: { email, password },
    auth: false,
  });
  setAccessToken(res.token);
  return res;
}

export async function signOut() {
  try {
    await apiRequest<void>("/auth/sign-out", { method: "POST", auth: true });
  } finally {
    setAccessToken(null);
  }
}