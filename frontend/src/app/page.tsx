import { redirect } from "next/navigation";
import { getSessionToken } from "@/lib/session";
import { getSession } from "@/lib/beebox";

export default async function HomePage() {
  const token = await getSessionToken();
  if (token) {
    try {
      await getSession(token);
      redirect("/welcome");
    } catch {
      // fall through
    }
  }
  redirect("/sign-in");
}
