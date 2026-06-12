import { iconRegistry } from '../../lib/icons'

export type ActionTarget =
  | 'add-trip'
  | 'add-item'
  | 'add-set'
  | 'add-person'
  | 'add-packing-list'
  | 'manage-manufacturers'
  | 'manage-categories'
  | 'import-csv'
  | 'settings'
  | 'logout'
  | 'trips'
  | 'planner'
  | 'sets'
  | 'lists'
  | 'gear'
  | 'persons'

export interface ActionOption {
  value: ActionTarget
  label: string
  description: string
  icon: string
}

// Shared quick-action list, consumed by both the mobile actions menu and the
// desktop action rail so the two stay in sync.
export const actionOptions: ActionOption[] = [
  { value: 'add-trip', label: 'Add Trip', description: 'Plan a new trip.', icon: `pi ${iconRegistry.navigation.trips}` },
  { value: 'add-item', label: 'Add Item', description: 'Create a new gear item.', icon: `pi ${iconRegistry.navigation.gear}` },
  { value: 'add-set', label: 'Add Set', description: 'Create a new gear set.', icon: `pi ${iconRegistry.navigation.sets}` },
  { value: 'add-person', label: 'Add Person', description: 'Create a new person.', icon: `pi ${iconRegistry.navigation.person}` },
  { value: 'add-packing-list', label: 'Add Packing List', description: 'Create a new packing list.', icon: `pi ${iconRegistry.navigation.planner}` },
  { value: 'manage-manufacturers', label: 'Manage Manufacturers', description: 'Create and edit manufacturers.', icon: `pi ${iconRegistry.content.building}` },
  { value: 'manage-categories', label: 'Manage Categories', description: 'Create and edit custom categories.', icon: `pi ${iconRegistry.content.tag}` },
  { value: 'import-csv', label: 'Import from CSV', description: 'Preview and import gear from CSV.', icon: `pi ${iconRegistry.action.upload}` },
  { value: 'settings', label: 'Settings', description: 'Configure app preferences.', icon: `pi ${iconRegistry.navigation.settings}` },
  { value: 'logout', label: 'Logout', description: 'Sign out of your account.', icon: `pi ${iconRegistry.navigation.logout}` },
]
