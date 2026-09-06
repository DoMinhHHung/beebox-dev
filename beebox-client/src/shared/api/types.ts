export type ApiErrorCode =
  | "INVALID_INPUT"
  | "UNAUTHENTICATED"
  | "FORBIDDEN"
  | "TENANT_ACCESS_DENIED"
  | "NOT_FOUND"
  | "CONFLICT"
  | "CREDENTIAL_INVALID"
  | "CREDENTIAL_REVOKED"
  | "CREDENTIAL_TYPE_UNSUPPORTED"
  | "INTERNAL";

export type ApiErrorBody = {
  error: {
    code: ApiErrorCode;
    message: string;
  };
};

export class ApiError extends Error {
  constructor(
    public readonly code: ApiErrorCode,
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}