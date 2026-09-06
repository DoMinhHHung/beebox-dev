export type ProjectTier = "free" | "pro" | "enterprise";

export type Project = {
  id: string;
  name: string;
  tier: ProjectTier;
  ownerId: string;
};

// Backend: snake_case
export type ProjectDto = {
  id: string;
  name: string;
  tier: string;
  owner_id: string;
};

export function mapProject(dto: ProjectDto): Project {
  return {
    id: dto.id,
    name: dto.name,
    tier: dto.tier as ProjectTier,
    ownerId: dto.owner_id,
  };
}