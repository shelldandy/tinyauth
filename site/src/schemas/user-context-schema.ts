import { z } from "zod";

export const userContextSchema = z.object({
  isLoggedIn: z.boolean(),
  username: z.string(),
  oauth: z.boolean(),
  provider: z.string(),
  configuredProviders: z.array(z.string()),
  disableContinue: z.boolean(),
  title: z.string(),
  genericName: z.string(),
});

export type UserContextSchemaType = z.infer<typeof userContextSchema>;
