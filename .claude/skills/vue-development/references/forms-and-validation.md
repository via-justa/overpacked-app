# Forms and Validation

Forms use `vee-validate` with `zod` schemas. The bridge is `buildTypedSchema`
(`lib/validation/schema.ts`), a thin wrapper over `@vee-validate/zod`'s `toTypedSchema`:

```ts
import { z } from 'zod'
import { buildTypedSchema } from '@/lib/validation/schema'

const schema = buildTypedSchema(
  z.object({
    name: z.string().min(1, 'Name is required'),
    weightGrams: z.number().positive(),
    url: z.string().optional(),
  }),
)
```

Conventions:
- Define the zod schema next to the form (or in the feature) and pass it through
  `buildTypedSchema` so vee-validate gets a typed schema and the form values are inferred.
- Reuse the shared validators in `lib/validation/validators.ts` (e.g. `isValidUrl`, `isFloat`)
  rather than re-implementing common checks. Note their convention: an empty/blank value is
  treated as valid (optional-field friendly) — required-ness is expressed separately in the
  schema.
- Surface validation messages in the UI; pair them with the explicit error-state handling the
  rest of the app uses.
- Keep accessibility in mind: associate messages with their fields, and make sure required and
  invalid states are conveyed to assistive tech, not by color alone.

When a form submits, the mutation goes through the vue-query mutation composables
(`useMutationWithToast` / `useInlineMutation`) so success/error feedback and cache
invalidation stay consistent with the rest of the app — don't call the API directly from the
component's submit handler.
