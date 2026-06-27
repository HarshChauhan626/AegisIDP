import { z } from "zod";

// Auth schemas
export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
});

export type LoginFormData = z.infer<typeof loginSchema>;

// Environment schemas
export const createEnvironmentSchema = z.object({
  project_id: z.string().min(1, "Project is required"),
  name: z
    .string()
    .min(2, "Name must be at least 2 characters")
    .max(64, "Name must be under 64 characters"),
  config: z.record(z.string(), z.unknown()).optional(),
});

export type CreateEnvironmentFormData = z.infer<typeof createEnvironmentSchema>;

// User schemas
export const createUserSchema = z.object({
  email: z.string().email("Invalid email address"),
  name: z.string().min(1, "Name is required"),
  password: z.string().min(6, "Password must be at least 6 characters"),
  role: z.enum(["admin", "developer", "viewer"]),
});

export type CreateUserFormData = z.infer<typeof createUserSchema>;

export const updateUserSchema = z.object({
  name: z.string().optional(),
  role: z.enum(["admin", "developer", "viewer"]).optional(),
  active: z.boolean().optional(),
});

export type UpdateUserFormData = z.infer<typeof updateUserSchema>;
