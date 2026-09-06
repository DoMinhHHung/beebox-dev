import {
  mapFieldSchema,
  type FieldDefinition,
  type FieldSchema,
  type FieldSchemaDto,
} from "@/entities/field-definition/types";
import { apiRequest } from "@/shared/api/http";

export async function defineFields(
  projectId: string,
  fields: FieldDefinition[],
): Promise<FieldSchema> {
  const dto = await apiRequest<FieldSchemaDto>(
    `/dashboard/projects/${projectId}/fields`,
    { method: "PUT", body: { fields } },
  );
  return mapFieldSchema(dto);
}

export async function getLatestFields(projectId: string): Promise<FieldSchema> {
  const dto = await apiRequest<FieldSchemaDto>(
    `/dashboard/projects/${projectId}/fields`,
  );
  return mapFieldSchema(dto);
}

export async function getFieldsVersion(
  projectId: string,
  version: number,
): Promise<FieldSchema> {
  const dto = await apiRequest<FieldSchemaDto>(
    `/dashboard/projects/${projectId}/fields/${version}`,
  );
  return mapFieldSchema(dto);
}