export type FieldKind = "STRING" | "NUMBER" | "BOOLEAN";

export type FieldDefinition = {
  name: string;
  kind: FieldKind;
  required: boolean;
};

export type FieldSchema = {
  projectId: string;
  version: number;
  fields: FieldDefinition[];
};

export type FieldSchemaDto = {
  project_id: string;
  version: number;
  fields: { name: string; kind: string; required: boolean }[];
};

export function mapFieldSchema(dto: FieldSchemaDto): FieldSchema {
  return {
    projectId: dto.project_id,
    version: dto.version,
    fields: dto.fields.map((f) => ({
      name: f.name,
      kind: f.kind as FieldKind,
      required: f.required,
    })),
  };
}