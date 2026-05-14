import { toTypedSchema } from '@vee-validate/zod'
import type { ZodTypeAny } from 'zod'

export const buildTypedSchema = <TSchema extends ZodTypeAny>(schema: TSchema) => {
  return toTypedSchema(schema)
}