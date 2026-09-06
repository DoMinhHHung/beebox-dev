export type CredentialEnvironment = "test" | "live";
export type CredentialStatus = "ACTIVE" | "REVOKED";

export type Credential = {
  id: string;
  projectId: string;
  environment: CredentialEnvironment;
  publicKey: string;
  status: CredentialStatus;
  createdAt: string;
  revokedAt: string | null;
  secretKey?: string; // chỉ có lúc create/rotate
};

export type CredentialDto = {
  id: string;
  project_id: string;
  environment: string;
  public_key: string;
  status: string;
  created_at: string;
  revoked_at: string | null;
  secret_key?: string;
};

export function mapCredential(dto: CredentialDto): Credential {
  return {
    id: dto.id,
    projectId: dto.project_id,
    environment: dto.environment as CredentialEnvironment,
    publicKey: dto.public_key,
    status: dto.status as CredentialStatus,
    createdAt: dto.created_at,
    revokedAt: dto.revoked_at,
    secretKey: dto.secret_key,
  };
}