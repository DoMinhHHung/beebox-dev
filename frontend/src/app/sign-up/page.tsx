import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import { SignUpForm } from "@/components/auth/SignUpForm";
import Link from "next/link";

export const metadata = { title: "Create account · BeeBox" };

export default function SignUpPage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center mb-6">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Create your account</h1>
          <p className="mt-1.5 text-sm text-on-surface-variant">Get started in a few moments</p>
        </div>
        <SignUpForm />
        <p className="mt-6 text-center text-sm text-on-surface-variant">
          Already have an account?{" "}
          <Link href="/sign-in" className="font-medium text-primary hover:underline">Sign in</Link>
        </p>
      </AuthCard>
    </AuthShell>
  );
}
