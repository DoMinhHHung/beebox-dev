import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import { ForgotPasswordForm } from "@/components/auth/ForgotPasswordForm";
import Link from "next/link";

export const metadata = { title: "Forgot password · BeeBox" };

export default function ForgotPasswordPage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center mb-6">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Reset your password</h1>
          <p className="mt-1.5 text-sm text-on-surface-variant">Enter your email and we’ll send you a reset link</p>
        </div>
        <ForgotPasswordForm />
        <p className="mt-6 text-center text-sm text-on-surface-variant">
          <Link href="/sign-in" className="font-medium text-primary hover:underline">Back to sign in</Link>
        </p>
      </AuthCard>
    </AuthShell>
  );
}
