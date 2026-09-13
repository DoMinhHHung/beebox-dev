import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import { ResetPasswordForm } from "@/components/auth/ResetPasswordForm";
import Link from "next/link";

export const metadata = { title: "Set new password · BeeBox" };

type Props = { searchParams: Promise<{ reset_id?: string; token?: string }> };

export default async function ResetPasswordPage({ searchParams }: Props) {
  const params = await searchParams;
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center mb-6">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface">Set a new password</h1>
          <p className="mt-1.5 text-sm text-on-surface-variant">Choose a strong password you haven’t used before</p>
        </div>
        <ResetPasswordForm resetId={params.reset_id ?? ""} token={params.token ?? ""} />
        <p className="mt-6 text-center text-sm text-on-surface-variant">
          <Link href="/sign-in" className="font-medium text-primary hover:underline">Back to sign in</Link>
        </p>
      </AuthCard>
    </AuthShell>
  );
}
