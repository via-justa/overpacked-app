import { setupServer } from 'msw/node'
import { authHandlers } from './handlers/auth'
import { settingsHandlers } from './handlers/settings'
import { personsHandlers } from './handlers/persons'
import { itemsHandlers } from './handlers/items'
import { listsHandlers } from './handlers/lists'
import { setsHandlers } from './handlers/sets'
import { tripsHandlers } from './handlers/trips'

// One shared MSW server. It intercepts at fetch, covering both the openapi-fetch
// apiClient and the raw fetch() in src/lib/api/auth.ts. Default handlers return
// empty/logged-out responses; tests override with server.use(...) per case.
export const server = setupServer(
  ...authHandlers,
  ...settingsHandlers,
  ...personsHandlers,
  ...itemsHandlers,
  ...listsHandlers,
  ...setsHandlers,
  ...tripsHandlers,
)
