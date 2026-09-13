type ButtonProps = {
  children: React.ReactNode;
  type?: "button" | "submit" | "reset";
  variant?: "primary" | "secondary" | "ghost";
  loading?: boolean;
  disabled?: boolean;
  className?: string;
  onClick?: () => void;
};

export function Button({ children, type = "button", variant = "primary", loading = false, disabled = false, className = "", onClick }: ButtonProps) {
  const base = "w-full h-11 rounded-lg font-medium text-sm flex items-center justify-center transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:opacity-60 disabled:cursor-not-allowed";
  const variants = {
    primary: "bg-primary-container hover:bg-secondary text-on-primary shadow-sm",
    secondary: "bg-surface-container-low hover:bg-surface-container text-on-surface border border-outline-variant",
    ghost: "bg-transparent hover:bg-surface-container-low text-on-surface",
  };
  return (
    <button type={type} disabled={disabled || loading} onClick={onClick} className={`${base} ${variants[variant]} ${className}`}>
      {loading ? (
        <span className="flex items-center gap-2">
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
          <span>Please wait…</span>
        </span>
      ) : children}
    </button>
  );
}
