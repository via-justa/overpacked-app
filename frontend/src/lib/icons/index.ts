/**
 * Icon system exports
 * 
 * Provides centralized icon registry and utilities for type-safe,
 * semantic icon usage throughout the application.
 */

export {
  // Registry objects
  iconRegistry,
  actionIcons,
  navigationIcons,
  statusIcons,
  contentIcons,
  directionalIcons,
  feedbackIcons,
  iconSizes,
  
  // Helper functions
  getIconClass,
  getIconName,
  getIconSizeClass,
  
  // Types
  type IconCategory,
  type ActionIcon,
  type NavigationIcon,
  type StatusIcon,
  type ContentIcon,
  type DirectionalIcon,
  type FeedbackIcon,
  type IconName,
  type IconSize,
} from './registry'
