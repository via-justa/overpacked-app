/**
 * Centralized icon registry for the application.
 * Maps semantic icon names to PrimeIcons class names.
 * 
 * Usage:
 * - Import registry constants for type-safe icon references
 * - Use with AppIcon component for consistent rendering
 * - Refer to iconRegistry for full icon mappings
 */

/**
 * Action icons for user interactions
 */
export const actionIcons = {
  create: 'pi-plus',
  delete: 'pi-trash',
  edit: 'pi-pencil',
  editField: 'pi-file-edit',
  cancel: 'pi-times',
  close: 'pi-times',
  confirm: 'pi-check',
  submit: 'pi-check',
  refresh: 'pi-refresh',
  reset: 'pi-refresh',
  upload: 'pi-upload',
  duplicate: 'pi-file-edit',
  menu: 'pi-ellipsis-h',
  menuBars: 'pi-bars',
  search: 'pi-search',
} as const

/**
 * Navigation icons for app sections and routing
 */
export const navigationIcons = {
  dashboard: 'pi-home',
  home: 'pi-home',
  gear: 'pi-box',
  items: 'pi-box',
  sets: 'pi-sitemap',
  packs: 'pi-briefcase',
  lists: 'pi-list',
  trips: 'pi-map',
  planner: 'pi-check-square',
  person: 'pi-user',
  persons: 'pi-users',
  settings: 'pi-cog',
  menu: 'pi-bars',
  login: 'pi-sign-in',
  logout: 'pi-sign-out',
} as const

/**
 * Status and state icons for values and feedback
 */
export const statusIcons = {
  success: 'pi-check-circle',
  error: 'pi-times-circle',
  warning: 'pi-exclamation-circle',
  notSet: 'pi-minus-circle',
  complete: 'pi-check-square',
  incomplete: 'pi-square',
  active: 'pi-check-circle',
  inactive: 'pi-times-circle',
} as const

/**
 * Content and media icons
 */
export const contentIcons = {
  image: 'pi-image',
  imagePlaceholder: 'pi-image',
  tag: 'pi-tag',
  label: 'pi-tag',
  externalLink: 'pi-external-link',
  file: 'pi-file',
  folder: 'pi-folder',
  building: 'pi-building',
  routeKomoot: 'pi-map-marker',
  routeStrava: 'pi-bolt',
  routeWanderer: 'pi-compass',
  routeLink: 'pi-link',
} as const

/**
 * Directional and navigational indicators
 */
export const directionalIcons = {
  chevronRight: 'pi-chevron-right',
  chevronLeft: 'pi-chevron-left',
  chevronDown: 'pi-chevron-down',
  chevronUp: 'pi-chevron-up',
  arrowRight: 'pi-arrow-right',
  arrowLeft: 'pi-arrow-left',
  arrowUp: 'pi-arrow-up',
  arrowDown: 'pi-arrow-down',
} as const

/**
 * Loading and feedback icons
 */
export const feedbackIcons = {
  spinner: 'pi-spinner',
  loading: 'pi-spinner',
  info: 'pi-info-circle',
} as const

/**
 * Combined icon registry
 */
export const iconRegistry = {
  action: actionIcons,
  navigation: navigationIcons,
  status: statusIcons,
  content: contentIcons,
  directional: directionalIcons,
  feedback: feedbackIcons,
} as const

/**
 * Type utilities for icon references
 */
export type IconCategory = keyof typeof iconRegistry
export type ActionIcon = keyof typeof actionIcons
export type NavigationIcon = keyof typeof navigationIcons
export type StatusIcon = keyof typeof statusIcons
export type ContentIcon = keyof typeof contentIcons
export type DirectionalIcon = keyof typeof directionalIcons
export type FeedbackIcon = keyof typeof feedbackIcons

export type IconName<T extends IconCategory> = keyof (typeof iconRegistry)[T]

/**
 * Helper function to get full PrimeIcon class string
 */
export function getIconClass(category: IconCategory, name: string): string {
  const iconName = iconRegistry[category]?.[name as keyof (typeof iconRegistry)[typeof category]]
  return iconName ? `pi ${iconName}` : ''
}

/**
 * Helper function to get just the icon name (without 'pi' prefix)
 */
export function getIconName(category: IconCategory, name: string): string {
  return iconRegistry[category]?.[name as keyof (typeof iconRegistry)[typeof category]] ?? ''
}

/**
 * Size utilities for consistent icon sizing
 */
export const iconSizes = {
  xs: 'text-xs',      // 12px
  sm: 'text-sm',      // 14px
  md: 'text-base',    // 16px
  lg: 'text-lg',      // 18px
  xl: 'text-xl',      // 20px
  '2xl': 'text-2xl',  // 24px
} as const

export type IconSize = keyof typeof iconSizes

/**
 * Get size class for icon
 */
export function getIconSizeClass(size: IconSize = 'md'): string {
  return iconSizes[size]
}
