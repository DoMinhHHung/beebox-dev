import { forwardRef } from "react";

type FormFieldProps = {
  id: string;
  label: string;
  type?: string;
  placeholder?: string;
  autoComplete?: string;
  error?: string;
  required?: boolean;
} & React.InputHTMLAttributes<HTMLInputElement>;

export const FormField = forwardRef<HTMLInputElement, FormFieldProps>(function FormField(
  { id, label, type = "text", placeholder, autoComplete, error, required, className = "", ...rest },
  ref
) {
  return (
    <div className="w-full">
      <label htmlFor={id} className="block text-sm font-medium text-on-surface mb-1.5">{label}</label>
      <input
        ref={ref}
        id={id}
        type={type}
        placeholder={placeholder}
        autoComplete={autoComplete}
        required={required}
        aria-invalid={!!error}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`w-full h-11 px-3.5 rounded-lg border bg-surface-container-lowest text-on-surface text-sm placeholder:text-on-surface-variant/70 focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary transition-colors ${error ? "border-error" : "border-outline-variant"} ${className}`}
        {...rest}
      />
      {error && <p id={`${id}-error`} className="mt-1.5 text-xs text-error" role="alert">{error}</p>}
    </div>
  );
});
