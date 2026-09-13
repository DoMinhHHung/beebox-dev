import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import { SignInForm } from "@/components/auth/SignInForm";
import Link from "next/link";

export const metadata = { title: "Sign in · BeeBox" };

export default function SignInPage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center mb-6">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Welcome back</h1>
          <p className="mt-1.5 text-sm text-on-surface-variant">Sign in to continue to your account</p>
        </div>
        <SignInForm />
        <p className="mt-6 text-center text-sm text-on-surface-variant">
          Don’t have an account?{" "}
          <Link href="/sign-up" className="font-medium text-primary hover:underline">Sign up</Link>
        </p>
      </AuthCard>
    </AuthShell>
  );
}
