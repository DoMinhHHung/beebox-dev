import { mapProject, type Project, type ProjectDto, type ProjectTier } from "@/entities/project/types";
import { apiRequest } from "@/shared/api/http";

export async function createProject(name: string, tier?: ProjectTier): Promise<Project> {
  const dto = await apiRequest<ProjectDto>("/dashboard/projects", {
    method: "POST",
    body: { name, tier: tier ?? "free" },
  });
  return mapProject(dto);
}