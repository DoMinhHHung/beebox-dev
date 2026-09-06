import {
  mapCredential,
  type Credential,
  type CredentialDto,
  type CredentialEnvironment,
} from "@/entities/credential/types";
import { apiRequest } from "@/shared/api/http";

export async function createCredential(
  projectId: string,
  environment: CredentialEnvironment,
): Promise<Credential> {
  const dto = await apiRequest<CredentialDto>(
    `/dashboard/projects/${projectId}/credentials`,
    { method: "POST", body: { environment } },
  );
  return mapCredential(dto);
}

export async function getCredential(credentialId: string): Promise<Credential> {
  const dto = await apiRequest<CredentialDto>(
    `/dashboard/credentials/${credentialId}`,
  );
  return mapCredential(dto);
}

export async function rotateCredential(credentialId: string): Promise<Credential> {
  const dto = await apiRequest<CredentialDto>(
    `/dashboard/credentials/${credentialId}/rotate`,
    { method: "POST" },
  );
  return mapCredential(dto);
}

export async function revokeCredential(credentialId: string): Promise<Credential> {
  const dto = await apiRequest<CredentialDto>(
    `/dashboard/credentials/${credentialId}/revoke`,
    { method: "POST" },
  );
  return mapCredential(dto);
}