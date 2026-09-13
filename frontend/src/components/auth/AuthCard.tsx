type AuthCardProps = { children: React.ReactNode; className?: string };

export function AuthCard({ children, className = "" }: AuthCardProps) {
  return (
    <div className={`w-full max-w-[420px] relative ${className}`}>
      <div className="absolute -bottom-1 left-2 right-2 h-2 bg-surface-container-low rounded-xl -z-10" aria-hidden />
      <div className="bg-surface-container-lowest rounded-xl shadow-[0_1px_2px_rgba(18,19,22,0.03),0_8px_24px_-4px_rgba(18,19,22,0.05)] p-6 sm:p-8">
        {children}
      </div>
    </div>
  );
}
